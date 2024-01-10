package client

import (
	"errors"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/run-ai/preinstall-diagnostics/internal/env"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	config        *rest.Config
	clientSet     *kubernetes.Clientset
	dynamicClient dynamic.Interface

	scheme *runtime.Scheme
)

func init() {
	scheme = runtime.NewScheme()
	utilruntime.Must(monitoringv1.AddToScheme(scheme))
}

func getConfig() (*rest.Config, error) {
	if config != nil {
		return config, nil
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	defaultKubeConfigPath := path.Join(userHomeDir, "/.kube/config")
	kubeConfigPath := env.EnvOrDefault(env.KubeConfigEnvVar, defaultKubeConfigPath)

	if _, err := os.Stat(kubeConfigPath); errors.Is(err, os.ErrNotExist) {
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		kubeconfigBytes, err := ioutil.ReadFile(kubeConfigPath)
		if err != nil {
			return nil, err
		}

		config, err = clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func Init(logger *log.Logger) error {
	logger.LogF("initializing Kubernetes client...")
	_, err := getConfig()
	if err != nil {
		return err
	}

	logger.LogF("successfully initialized Kubernetes client")
	return nil
}

func ClientSet() (*kubernetes.Clientset, error) {
	if clientSet != nil {
		return clientSet, nil
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

func NewClient() (client.Client, error) {
	return client.New(ctrl.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
}
