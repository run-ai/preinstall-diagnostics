package external_cluster_tests

import (
	"context"
	"github.com/run-ai/preinstall-diagnostics/internal/k8sclient"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListPods() ([]v1.Pod, error) {
	k8s, err := k8sclient.ClientSet()
	if err != nil {
		return nil, err
	}

	var podList *v1.PodList
	var pods []v1.Pod
	cont := ""

	for podList == nil || podList.Continue != "" {
		var err error
		podList, err = k8s.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			Limit:    500,
			Continue: cont,
		})
		if err != nil {
			return nil, err
		}

		cont = podList.Continue
		pods = append(pods, podList.Items...)
	}

	return pods, nil
}
