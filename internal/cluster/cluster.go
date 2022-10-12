package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	ocpv1 "github.com/openshift/api/config/v1"
	"github.com/run-ai/preinstall-diagnostics/internal/client"
	"github.com/run-ai/preinstall-diagnostics/internal/env"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"github.com/run-ai/preinstall-diagnostics/internal/resources"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

var (
	nvidiaLabels = map[string]string{
		"nvidia.com/cuda.driver.major":  "",
		"nvidia.com/cuda.driver.minor":  "",
		"nvidia.com/cuda.driver.rev":    "",
		"nvidia.com/cuda.runtime.major": "",
		"nvidia.com/cuda.runtime.minor": "",
		"nvidia.com/gfd.timestamp":      "",
		"nvidia.com/gpu.compute.major":  "",
		"nvidia.com/gpu.compute.minor":  "",
		"nvidia.com/gpu.count":          "",
		"nvidia.com/gpu.family":         "",
		"nvidia.com/gpu.machine":        "",
		"nvidia.com/gpu.memory":         "",
		"nvidia.com/gpu.product":        "",

		// Not all of our custumers support MIG?
		"nvidia.com/mig.strategy": "",
	}

	nginxLabels = map[string]string{
		"app": "nginx-ingress",
	}

	nfdLabels = map[string]string{
		"app.kubernetes.io/name": "node-feature-discovery",
	}

	gfdLabels = map[string]string{
		"app.kubernetes.io/name": "gpu-feature-discovery",
	}
)

func startPingPongServer(logger *log.Logger) {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		tjs, err := json.Marshal(t)
		if err != nil {
			logger.ErrorF("failed to marshal system time: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte(tjs))
	})

	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			panic(err)
		}
	}()
}

func GetCompletePodLogs(pod *v1.Pod, logger *log.Logger) (string, error) {
	client, err := client.Clientset()
	if err != nil {
		return "", err
	}

	logsReady := false
	fetchLogAttempts := attempts
	logs := ""

	for fetchLogAttempts > 0 && !logsReady {
		req := client.CoreV1().Pods(pod.Namespace).
			GetLogs(pod.Name, &v1.PodLogOptions{})
		res, err := req.DoRaw(context.TODO())
		if err != nil {
			return "", err
		}

		logs = string(res)
		logsReady = strings.Contains(logs, log.CompleteTag)

		if !logsReady {
			logger.LogF("logs for [%s/%s] are not ready yet, retrying in %d seconds...",
				pod.Spec.NodeName, pod.Name, sleepInterval/time.Second)
		}

		fetchLogAttempts--
		time.Sleep(sleepInterval)
	}

	if logsReady {
		return fmt.Sprintf("Logs for [%s/%s]:\n%s", pod.Spec.NodeName, pod.Name, logs), nil
	} else {
		return "", fmt.Errorf("timed out waiting for [%s/%s] logs to be ready",
			pod.Spec.NodeName, pod.Name)
	}
}

func GetDaemonsetPods(client *kubernetes.Clientset) ([]v1.Pod, error) {
	labelSelector := strings.ReplaceAll(labels.FormatLabels(resources.DaemonSet.Spec.Template.Labels), "=", "")
	pods, err := client.CoreV1().Pods(resources.DaemonSet.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func WaitDaemonSetAvailable(logger *log.Logger) error {
	client, err := client.Clientset()
	if err != nil {
		return err
	}

	dsAvailabilityAttempts := attempts
	available := false

	for dsAvailabilityAttempts > 0 && !available {
		ds, err := client.AppsV1().DaemonSets(resources.DaemonSet.Namespace).Get(context.TODO(),
			resources.DaemonSet.Name, metav1.GetOptions{})
		if err != nil {
			logger.LogF("fetching daemonset failed with %v, retrying in %d seconds",
				err, sleepInterval/time.Second)
		} else {
			if ds.Status.DesiredNumberScheduled != 0 &&
				ds.Status.NumberAvailable == ds.Status.DesiredNumberScheduled {
				logger.LogF("all daemonset pods are available")
				return nil
			} else {
				logger.LogF("not all pods are ready [%d/%d], retrying in %d seconds",
					ds.Status.NumberAvailable, ds.Status.DesiredNumberScheduled, sleepInterval/time.Second)
			}
		}

		dsAvailabilityAttempts--
		time.Sleep(sleepInterval)
	}

	return fmt.Errorf("timed out waiting for daemonset to be ready")
}

func waitAllPodsPingable(logger *log.Logger) error {
	client, err := client.Clientset()
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

	pods, err := GetDaemonsetPods(client)
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

		pods, err = GetDaemonsetPods(client)
		if err != nil {
			return err
		}
	}

	if !pingable {
		return fmt.Errorf("failed to ping all pods")
	}

	return nil
}

func CheckNodeConnectivity(logger *log.Logger) error {
	logger.TitleF("Node connectivity check")

	startPingPongServer(logger)

	err := WaitDaemonSetAvailable(logger)
	if err != nil {
		return err
	}

	err = waitAllPodsPingable(logger)
	if err != nil {
		return err
	}

	return nil
}

func isOpenShift() (bool, string, error) {
	dclient, err := client.DynamicClient()
	if err != nil {
		return false, "", err
	}

	ver, err := dclient.Resource(schema.GroupVersionResource{
		Group:    "config.openshift.io",
		Version:  "v1",
		Resource: "clusterversions",
	}).Get(context.TODO(), "version", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, "", nil
		} else {
			return false, "", err
		}
	}

	verJSON, err := json.Marshal(ver.Object)
	if err != nil {
		return false, "", err
	}

	verCR := &ocpv1.ClusterVersion{}

	err = json.Unmarshal(verJSON, verCR)
	if err != nil {
		return false, "", err
	}

	return true, verCR.Status.History[0].Version, nil
}

func ShowClusterVersion(logger *log.Logger) error {
	logger.TitleF("Cluster Version")

	dc, err := client.Clientset()
	if err != nil {
		return err
	}

	ver, err := dc.ServerVersion()
	if err != nil {
		return err
	}

	logger.LogF("Kubernetes Cluster Version: %s", ver.String())

	isOpenShift, version, err := isOpenShift()
	if err != nil {
		return err
	}

	if isOpenShift {
		logger.LogF("Openshift Cluster Version: %s", version)
	}

	return nil
}

func PrintDNSResolvConf(logger *log.Logger) error {
	logger.TitleF("Print resolv.conf")

	f, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return err
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	logger.LogF(string(content))

	return nil
}

func ShowGPUNodes(logger *log.Logger) error {
	logger.TitleF("GPU Nodes")

	client, err := client.Clientset()
	if err != nil {
		return err
	}

	labelSelector := strings.ReplaceAll(labels.FormatLabels(nvidiaLabels), "=", "")

	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return err
	}

	for _, node := range nodes.Items {
		logger.LogF("Node name: %s", node.Name)
	}

	logger.LogF("please verify that the list above includes all GPU nodes in the cluster")
	logger.LogF("if you suspect GPU nodes are missing from the list above, gpu-feature-discovery might be malfunctioning")

	return nil
}

func PrometheusNotInstalled(logger *log.Logger) error {
	logger.TitleF("Prometheus check")

	dclient, err := client.DynamicClient()
	if err != nil {
		return err
	}

	_, err = dclient.Resource(schema.GroupVersionResource{
		Group:    "monitoring.coreos.com",
		Version:  "v1",
		Resource: "prometheuses",
	}).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		} else {
			return nil
		}
	}

	return fmt.Errorf("prometheus is installed in the cluster")
}

func NodeFeatureDiscoveryNotInstalled(logger *log.Logger) error {
	logger.TitleF("Node Feature Discovery")

	client, err := client.Clientset()
	if err != nil {
		return err
	}

	pods, err := client.CoreV1().Pods("").
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: labels.FormatLabels(nfdLabels),
		})
	if err != nil {
		return err
	}

	if len(pods.Items) != 0 {
		return fmt.Errorf("node-feature-discovery is installed in the cluster")
	}

	return nil
}

func GPUFeatureDiscoveryNotInstalled(logger *log.Logger) error {
	logger.TitleF("GPU Feature Discovery")

	client, err := client.Clientset()
	if err != nil {
		return err
	}

	pods, err := client.CoreV1().Pods("").
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: labels.FormatLabels(gfdLabels),
		})
	if err != nil {
		return err
	}

	if len(pods.Items) != 0 {
		return fmt.Errorf("gpu-feature-discovery is installed in the cluster")
	}

	return nil
}

func ShowStorageClasses(logger *log.Logger) error {
	logger.TitleF("Storage Classes")

	client, err := client.Clientset()
	if err != nil {
		return err
	}

	scs, err := client.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(scs.Items) == 0 {
		logger.LogF("No storage classes defined in the cluster")
		return nil
	}

	logger.LogF("StorageClasses in cluster:")

	for _, sc := range scs.Items {
		logger.LogF("	%s", sc.Name)
	}

	return nil
}

func NvidiaDevicePluginNotInstalled(logger *log.Logger) error {
	logger.TitleF("Nvidia device plugin")

	client, err := client.Clientset()
	if err != nil {
		return err
	}

	dss, err := client.AppsV1().DaemonSets("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, ds := range dss.Items {
		if strings.Contains(ds.Name, nvidiaDevicePluginDaemonset) {
			return fmt.Errorf("nvidia device plugin is installed in the cluster")
		}
	}

	return nil
}

func DCGMExporterNotInstalled(logger *log.Logger) error {
	logger.TitleF("DCGM Exporter")

	client, err := client.Clientset()
	if err != nil {
		return err
	}

	dss, err := client.AppsV1().DaemonSets("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, ds := range dss.Items {
		if strings.Contains(ds.Name, dcgmExporterDaemonset) {
			return fmt.Errorf("dcgm exporter is installed in the cluster")
		}
	}

	return nil
}

func ResolveBackendFQDN(logger *log.Logger) error {
	logger.TitleF("DNS Resolver")

	backendFQDN := env.EnvOrDefault(env.BackendFQDNEnvVar, "")
	if backendFQDN == "" {
		logger.WarningF("Backend FQDN was not provided using the --domain flag, skipping test")
		logger.Skip()
		return nil
	}

	ips, err := net.DefaultResolver.LookupIP(context.TODO(), "ip", backendFQDN)
	if err != nil {
		return err
	}

	logger.LogF("Resolved IP addresses for %s", backendFQDN)
	for _, ip := range ips {
		logger.LogF(ip.String())
	}

	client, err := client.Clientset()
	if err != nil {
		return err
	}

	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, ip := range ips {
		for _, node := range nodes.Items {
			for _, nodeIP := range node.Status.Addresses {
				if nodeIP.Address == ip.String() {
					logger.LogF("%s ip address is resolved to the IP of node %s", backendFQDN, node.Name)
				}
			}
		}
	}

	return nil
}

func NginxIngressControllerNotInstalled(logger *log.Logger) error {
	logger.TitleF("Nginx Ingress Controller")

	client, err := client.Clientset()
	if err != nil {
		return err
	}

	_, err = client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		LabelSelector: labels.FormatLabels(nginxLabels),
	})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		} else {
			return nil
		}
	}

	return fmt.Errorf("nginx ingress controller is installed in the cluster")
}

func ShowOSInfo(logger *log.Logger) error {
	logger.TitleF("OS Information")

	uname := exec.Command("uname", "-a")
	output, err := uname.Output()
	if err != nil {
		switch e := err.(type) {
		case *exec.ExitError:
			return fmt.Errorf("%s", string(e.Stderr))
		default:
			return err
		}
	}

	logger.LogF("Os Info: %s", string(output))
	return nil
}

func checkURLAvailable(url string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	} else {
		if res.StatusCode >= 400 && res.StatusCode < 200 {
			return fmt.Errorf("%s is not reachable, got status code %d", url, res.StatusCode)
		}
	}

	return nil
}

func RunAIHelmRepositoryReachable(logger *log.Logger) error {
	logger.TitleF("Run:AI Helm Repository")

	const runaiCharts = "https://runai-charts.storage.googleapis.com"

	return checkURLAvailable(runaiCharts)
}

func DockerHubReachable(logger *log.Logger) error {
	logger.TitleF("DockerHub")

	const dockerHub = "https://hub.docker.com"

	return checkURLAvailable(dockerHub)
}

func QuayIOReachable(logger *log.Logger) error {
	logger.TitleF("Quay.io")

	const quay = "https://quay.io"

	return checkURLAvailable(quay)
}

func RunAIPrometheusReachable(logger *log.Logger) error {
	logger.TitleF("Run:AI Prometheus")

	const prom = "https://prometheus-us-central1.grafana.net"

	return checkURLAvailable(prom)
}

func RunAIAuthProviderReachable(logger *log.Logger) error {
	logger.TitleF("Run:AI Auth Provider")

	const auth = "https://runai-prod.auth0.com"

	return checkURLAvailable(auth)
}

func ListPods(logger *log.Logger) error {
	logger.TitleF("List Pods")

	client, err := client.Clientset()
	if err != nil {
		return err
	}

	var podList *v1.PodList
	var pods []v1.Pod
	cont := ""

	for podList == nil || podList.Continue != "" {
		var err error
		podList, err = client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			Limit:    500,
			Continue: cont,
		})
		if err != nil {
			return err
		}

		cont = podList.Continue
		pods = append(pods, podList.Items...)
	}

	logger.LogF("Namespace/Name/Phase")

	for _, pod := range pods {
		logger.LogF("%s/%s/%s", pod.Namespace, pod.Name, pod.Status.Phase)
	}

	return nil
}
