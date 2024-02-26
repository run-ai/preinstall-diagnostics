package external_cluster_tests

import (
	"context"
	"fmt"
	"github.com/run-ai/preinstall-diagnostics/internal/k8sclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NGINXIngressControllerInstalled() (bool, error) {
	k8s, err := k8sclient.ClientSet()
	if err != nil {
		return false, err
	}

	ics, err := k8s.NetworkingV1().IngressClasses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	if len(ics.Items) > 0 {
		return true, nil
	}

	return false, fmt.Errorf("an ingress controller is not installed in the cluster")
}
