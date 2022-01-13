package registry

import (
	"fmt"
	"net/http"

	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"github.com/run-ai/preinstall-diagnostics/internal/util"
)

const (
	RegistryEnvVar = "REGISTRY"

	RunAIProdRegistry    = "gcr.io/run-ai-prod"
	RunAIProdRegistryURL = "https://" + RunAIProdRegistry
)

var (
	// This is modified at build time by `make binary`
	RunAIDiagnosticsImage = "gcr.io/run-ai-prod/preinstall-diagnostics:latest"
)

func CheckRunAIRegistryReachable() error {
	registry := util.EnvOrDefault(RegistryEnvVar, RunAIProdRegistryURL)

	log.TitleF("Connectivity to runai container registry: %s", registry)

	res, err := http.Get(registry)
	if err != nil {
		return err
	}

	if res.StatusCode >= 400 && res.StatusCode < 200 {
		return fmt.Errorf("status code of %d was returned from %s", res.StatusCode, registry)
	}

	log.LogF("Run:AI container registry is accessible")
	return nil
}
