package saas

import (
	"fmt"
	"net/http"

	"github.com/run-ai/preinstall-diagnostics/internal/env"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
)

const (
	RunAISaasAddress = "https://app.run.ai"
)

func CheckRunAIBackendReachable(logger *log.Logger) error {
	logger.TitleF("Run:AI service access: %s", RunAISaasAddress)
	saas := env.EnvOrDefault(env.RunAISaasEnvVar, RunAISaasAddress)

	res, err := http.Get(saas)
	if err != nil {
		return err
	}

	if res.StatusCode >= 300 && res.StatusCode < 200 {
		return fmt.Errorf("status code of %d was returned from %s", res.StatusCode, saas)
	}

	logger.LogF("Run:AI service is accessible from within the cluster")
	return nil
}
