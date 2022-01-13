package saas

import (
	"fmt"
	"net/http"

	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"github.com/run-ai/preinstall-diagnostics/internal/util"
)

const (
	RunAISaasEnvVar = "RUNAI_SAAS"

	RunAISaasAddress = "https://app.run.ai"
)

func CheckRunAIBackendReachable() error {
	log.TitleF("Run:AI service access: %s", RunAISaasAddress)
	saas := util.EnvOrDefault(RunAISaasEnvVar, RunAISaasAddress)

	res, err := http.Get(saas)
	if err != nil {
		return err
	}

	if res.StatusCode >= 300 && res.StatusCode < 200 {
		return fmt.Errorf("status code of %d was returned from %s", res.StatusCode, saas)
	}

	log.LogF("Run:AI service is accessible from within the cluster")
	return nil
}
