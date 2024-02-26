package external_cluster_tests

import (
	"context"
	"fmt"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/run-ai/preinstall-diagnostics/internal/k8sclient"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime_client "sigs.k8s.io/controller-runtime/pkg/client"
)

func PrometheusInstalled() (bool, error) {
	k8s, err := k8sclient.NewClient()
	if err != nil {
		return false, err
	}

	testProm := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	err = k8s.Get(context.TODO(), runtime_client.ObjectKeyFromObject(testProm), testProm)
	if err != nil {
		if meta.IsNoMatchError(err) {
			return false, fmt.Errorf("prometheus is not installed in the cluster")
		} else if !errors.IsNotFound(err) {
			return false, err
		}
	}

	return true, nil
}
