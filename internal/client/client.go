package client

import (
	"os"
	"path"

	"github.com/run-ai/preinstall-diagnostics/internal/env"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

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
	var err error
	config, err = clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		defaultKubeConfigPath := path.Join(userHomeDir, "/.kube/config")
		kubeConfigPath := env.EnvOrDefault(env.KubeConfigEnvVar, defaultKubeConfigPath)

		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func Init() error {
	log.LogF("initializing Kubernetes client...")
	_, err := getConfig()
	if err != nil {
		return err
	}

	log.LogF("successfully initizlied Kubernetes client")
	return nil
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
