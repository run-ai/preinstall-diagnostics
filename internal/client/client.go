package client

import (
	"github.com/run-ai/preinstall-diagnostics/internal/util"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const kubeConfigEnvVar = "KUBECONFIG"

var (
	config        *rest.Config
	clientset     *kubernetes.Clientset
	dynamicClient dynamic.Interface
)

func getConfig() (*rest.Config, error) {
	if config != nil {
		return config, nil
	}

	// Try to build internal cluster config, and if that fails,
	// try external cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeConfigPath, kubeConfErr := util.EnvOrError(kubeConfigEnvVar)
		if kubeConfErr != nil {
			return nil, kubeConfErr
		}

		config, kubeConfErr = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if kubeConfErr != nil {
			return nil, kubeConfErr
		}
	}

	return config, nil
}

func Clientset() (*kubernetes.Clientset, error) {
	if clientset != nil {
		return clientset, nil
	}

	conf, err := getConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(conf)
}

func DynamicClient() (dynamic.Interface, error) {
	if dynamicClient != nil {
		return dynamicClient, nil
	}

	conf, err := getConfig()
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(conf)
}
