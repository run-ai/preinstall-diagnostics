package job

import (
	"context"
	"encoding/json"
	v2 "github.com/run-ai/preinstall-diagnostics/internal"
	"github.com/run-ai/preinstall-diagnostics/internal/env"
	"github.com/run-ai/preinstall-diagnostics/internal/k8sclient"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"github.com/run-ai/preinstall-diagnostics/internal/resources"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Main(logger *log.Logger) {
	testResults := runTestsAndAppendResults(logger)

	err := deleteConfigMapIfExists()
	if err != nil {
		panic(err)
	}

	err = createConfigMapWithTestResults(testResults)
	if err != nil {
		panic(err)
	}
}

func deleteConfigMapIfExists() error {
	k8s, err := k8sclient.ClientSet()
	if err != nil {
		return err
	}

	err = k8s.CoreV1().ConfigMaps(resources.Namespace.Name).Delete(context.TODO(),
		"runai-diagnostics-"+env.EnvOrDefault("NODE_NAME", ""),
		metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}

func createConfigMapWithTestResults(results []v2.TestResult) error {
	k8s, err := k8sclient.ClientSet()
	if err != nil {
		return err
	}

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return err
	}

	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "runai-diagnostics-" + env.EnvOrDefault("NODE_NAME", ""),
			Namespace: resources.Namespace.Name,
		},
		Data: map[string]string{
			"results": string(resultsJSON),
		},
	}

	_, err = k8s.CoreV1().ConfigMaps(resources.Namespace.Name).Create(context.TODO(), &cm, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}
