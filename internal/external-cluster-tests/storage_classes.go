package external_cluster_tests

import (
	"context"
	"fmt"
	"github.com/run-ai/preinstall-diagnostics/internal/k8sclient"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ShowStorageClasses() ([]v1.StorageClass, error) {
	k8s, err := k8sclient.ClientSet()
	if err != nil {
		return nil, err
	}

	scs, err := k8s.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(scs.Items) == 0 {
		return nil, fmt.Errorf("No storage classes defined in the cluster")
	}

	return scs.Items, nil
}
