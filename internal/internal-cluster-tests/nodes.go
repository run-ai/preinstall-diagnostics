package internal_cluster_tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/run-ai/preinstall-diagnostics/internal/env"
	"github.com/run-ai/preinstall-diagnostics/internal/k8sclient"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"github.com/run-ai/preinstall-diagnostics/internal/resources"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

const (
	// Maximum amount of times to test for availability
	attempts = 100

	// Interval to wait between availability checks
	sleepInterval = 5 * time.Second

	nvidiaDevicePluginDaemonset = "nvidia-device-plugin"
	dcgmExporterDaemonset       = "dcgm-exporter"

	maximumAllowedTimeDiff = time.Minute
)

func ShowOSInfo() (string, error) {
	uname := exec.Command("uname", "-a")
	output, err := uname.Output()
	if err != nil {
		switch e := err.(type) {
		case *exec.ExitError:
			return "", fmt.Errorf("%s", string(e.Stderr))
		default:
			return "", err
		}
	}

	return strings.Join(strings.Split(string(output), " "), "\n"), nil
}

func CheckNodeConnectivity(logger *log.Logger) error {

	startPingPongServer()

	err := WaitForJobPodsToBeRunning(logger)
	if err != nil {
		return err
	}

	err = waitAllPodsPingable(logger)
	if err != nil {
		return err
	}

	return nil
}

func startPingPongServer() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		tjs, err := json.Marshal(t)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		_, _ = w.Write(tjs)
	})

	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			panic(err)
		}
	}()
}

func WaitForJobPodsToBeRunning(logger *log.Logger) error {
	k8s, err := k8sclient.ClientSet()
	if err != nil {
		return err
	}

	nodeList, err := k8s.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	nodeCount := len(nodeList.Items)

	podAvailabilityAttempts := attempts

	for podAvailabilityAttempts > 0 {
		logger.LogF("waiting for jobs to be available...")

		jobs, err := k8s.BatchV1().Jobs("runai-diagnostics").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		logger.LogF("waiting for pods to be available...")
		availablePods := 0
		for _, job := range jobs.Items {
			if job.Status.Ready == nil {
				continue
			}
			availablePods += int(*job.Status.Ready)
		}

		if availablePods == nodeCount {
			return nil
		}

		podAvailabilityAttempts--
		time.Sleep(sleepInterval)
	}

	return fmt.Errorf("timed out waiting for testing pods to be ready")
}

func GetJobsPods(client *kubernetes.Clientset) ([]v1.Pod, error) {
	labelSelector := strings.ReplaceAll(labels.FormatLabels(map[string]string{
		"runai-diagnostics": "",
	}), "=", "")
	pods, err := client.CoreV1().Pods(resources.Namespace.Name).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func waitAllPodsPingable(logger *log.Logger) error {
	k8s, err := k8sclient.ClientSet()
	if err != nil {
		return err
	}

	nodeName, err := env.EnvOrError(env.NodeNameEnvVar)
	if err != nil {
		return err
	}

	podName, err := env.EnvOrError(env.PodNameEnvVar)
	if err != nil {
		return err
	}

	pingedPods := map[string]struct{}{}
	podPingAttempts := attempts
	pingable := false

	pods, err := GetJobsPods(k8s)
	if err != nil {
		return err
	}

	for podPingAttempts > 0 && !pingable {
		for _, pod := range pods {
			_, pinged := pingedPods[pod.Name]
			if pinged {
				continue
			}

			logger.LogF("attempting to ping pod [%s/%s]...", pod.Spec.NodeName, pod.Name)

			ip := pod.Status.PodIP
			url := fmt.Sprintf("%s//%s:%s/%s", "http:", ip, "8080", "ping")

			res, err := http.Get(url)
			if err != nil {
				logger.ErrorF("[%s/%s] -> [%s/%s]: could not ping [%s/%s] due to %v, retrying in %d seconds",
					nodeName, podName, pod.Spec.NodeName, pod.Name,
					pod.Spec.NodeName, pod.Name, err, sleepInterval/time.Second)
			} else {
				if res.StatusCode != 200 {
					logger.ErrorF("[%s/%s] -> [%s/%s]: http ping failed got status code %d",
						nodeName, podName, pod.Spec.NodeName, pod.Name, res.StatusCode)
				} else {
					logger.LogF("[%s/%s] -> [%s/%s]: successfully pinged",
						nodeName, podName, pod.Spec.NodeName, pod.Name)

					targetTimeJSON, err := io.ReadAll(res.Body)
					if err != nil {
						return fmt.Errorf("[%s/%s] -> [%s/%s]: could not read pod ping response body: %v",
							nodeName, podName, pod.Spec.NodeName, pod.Name, err)
					}

					myTime := time.Now()
					targetTime := time.Time{}
					err = json.Unmarshal(targetTimeJSON, &targetTime)
					if err != nil {
						return fmt.Errorf("[%s/%s] -> [%s/%s]: could not parse target pod time: %v",
							nodeName, podName, pod.Spec.NodeName, pod.Name, err)
					}

					var diff time.Duration
					if targetTime.After(myTime) {
						diff = targetTime.Sub(myTime)
					} else {
						diff = myTime.Sub(targetTime)
					}

					if diff > maximumAllowedTimeDiff {
						logger.ErrorF("[%s/%s] -> [%s/%s]: node clocks are out of sync",
							nodeName, podName, pod.Spec.NodeName, pod.Name)
					} else {
						logger.LogF("[%s/%s] -> [%s/%s]: node clocks are in sync",
							nodeName, podName, pod.Spec.NodeName, pod.Name)
						pingedPods[pod.Name] = struct{}{}
					}
				}
			}
		}

		podPingAttempts--
		pingable = len(pingedPods) == len(pods)

		time.Sleep(sleepInterval)

		pods, err = GetJobsPods(k8s)
		if err != nil {
			return err
		}
	}

	if !pingable {
		return fmt.Errorf("failed to ping all pods")
	}

	return nil
}
