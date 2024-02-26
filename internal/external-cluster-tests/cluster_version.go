package external_cluster_tests

import (
	"context"
	"encoding/json"
	"fmt"
	ocpv1 "github.com/openshift/api/config/v1"
	"github.com/run-ai/preinstall-diagnostics/internal/k8sclient"
	"golang.org/x/mod/semver"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func ShowClusterVersion() (version string, err error) {
	dc, err := k8sclient.ClientSet()
	if err != nil {
		return "", err
	}

	ver, err := dc.ServerVersion()
	if err != nil {
		return "", err
	}

	if semver.Compare(ver.String(), "v1.20.0") < 0 {
		err := fmt.Errorf("Kubernetes Cluster Version is lower than 1.20.0")
		return "", err
	}

	isOpenShift, ocpVersion, err := isOpenShift()
	if err != nil {
		return "", err
	}

	if isOpenShift {
		return fmt.Sprintf("Openshift Cluster Version: %s", ocpVersion), nil
	}

	return ver.String(), nil
}

func isOpenShift() (bool, string, error) {
	dclient, err := k8sclient.DynamicClient()
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
