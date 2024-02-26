package registry

import (
	"fmt"
	"net/http"

	"github.com/run-ai/preinstall-diagnostics/internal/env"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
)

const (
	RunAIProdRegistry    = "gcr.io/run-ai-prod"
	RunAIProdRegistryURL = "https://" + RunAIProdRegistry
)

var (
	// This is modified at build time by `make binary`
	RunAIDiagnosticsImage = "gcr.io/run-ai-lab/preinstall-diagnostics:latest"
)

func CheckRunAIRegistryReachable(logger *log.Logger) error {
	registry := env.EnvOrDefault(env.RegistryEnvVar, RunAIProdRegistryURL)

	logger.TitleF("Connectivity to runai container registry: %s", registry)

	res, err := http.Get(registry)
	if err != nil {
		return err
	}

	if res.StatusCode >= 400 && res.StatusCode < 200 {
		return fmt.Errorf("status code of %d was returned from %s", res.StatusCode, registry)
	}

	logger.LogF("Run:AI container registry is accessible")
	return nil
}
