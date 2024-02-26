package env

import (
	"fmt"
	"os"
)

const (
	KubeConfigEnvVar = "KUBECONFIG"

	NodeNameEnvVar     = "NODE_NAME"
	PodNameEnvVar      = "POD_NAME"
	PodNamespaceEnvVar = "POD_NAMESPACE"

	BackendFQDNEnvVar = "BACKEND_FQDN"

	RegistryEnvVar = "REGISTRY"

	RunAISaasEnvVar = "RUNAI_SAAS"
)

func EnvOrError(envVar string) (string, error) {
	val, defined := os.LookupEnv(envVar)
	if !defined {
		return "", fmt.Errorf("env var %s is not defined", envVar)
	}

	return val, nil
}

func EnvOrDefault(envVar, def string) string {
	val, err := EnvOrError(envVar)
	if err != nil {
		return def
	}

	return val
}
