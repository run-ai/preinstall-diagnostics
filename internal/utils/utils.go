package utils

import (
	"context"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/run-ai/preinstall-diagnostics/internal/k8sclient"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"github.com/run-ai/preinstall-diagnostics/internal/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func CheckURLAvailable(url string) (bool, error) {
	res, err := http.Get(url)
	if err != nil {
		return false, err
	} else {
		if res.StatusCode >= 400 && res.StatusCode < 200 {
			return false, fmt.Errorf("%s is not reachable, got status code %d", url, res.StatusCode)
		}
	}

	return true, nil
}

func WaitForJobsToComplete(interval, timeout time.Duration) error {
	k8s, err := k8sclient.ClientSet()
	if err != nil {
		return err
	}

	for ; timeout > 0; timeout -= interval {
		time.Sleep(interval)

		jobs, err := k8s.BatchV1().Jobs("runai-preinstall-diagnostics").List(context.TODO(),
			metav1.ListOptions{})
		if err != nil {
			return err
		}

		completedJobs := 0
		for _, job := range jobs.Items {
			if job.Status.CompletionTime != nil {
				completedJobs++
			}
		}

		if completedJobs == len(jobs.Items) {
			return nil
		}
	}

	return fmt.Errorf("timed out waiting for jobs to be completed")
}

func AppendRowToTable(t table.Writer, testName string, testResult bool, testMessage string) {
	testResultStr := func() string {
		if testResult {
			return log.Green("PASS")
		} else {
			return log.Red("FAIL")
		}
	}
	t.AppendRow(table.Row{testName, testResultStr(), testMessage})
}

func DeleteResources(resourcesToDelete []client.Object, dynClient dynamic.Interface, logger *log.Logger) error {
	err := resources.DeleteResources(resourcesToDelete, dynClient)
	if err != nil {
		return err
	}

	logger.WriteStringF("all resources were successfully deleted")
	return nil
}

func CreateResources(resourcesToCreate []client.Object, dynClient dynamic.Interface) error {
	return resources.CreateResources(resourcesToCreate, dynClient)
}
