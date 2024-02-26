package external_cluster_tests

import (
	"context"
	"fmt"
	"github.com/run-ai/preinstall-diagnostics/internal/k8sclient"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"strings"
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
)

func ShowGPUNodes() ([]v1.Node, error) {
	k8s, err := k8sclient.ClientSet()
	if err != nil {
		return nil, err
	}

	labelSelector := strings.ReplaceAll(labels.FormatLabels(nvidiaLabels), "=", "")

	nodes, err := k8s.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	if len(nodes.Items) == 0 {
		return nil, fmt.Errorf("No GPU nodes were found in the cluster")
	}

	return nodes.Items, nil
}
