package internal_cluster_tests

import (
	"fmt"
	"github.com/run-ai/preinstall-diagnostics/internal/env"
	"github.com/run-ai/preinstall-diagnostics/internal/utils"
	"net/http"
)

const (
	RunAISaasHostname = "app.run.ai"
	RunAISaasAddress  = "https://" + RunAISaasHostname
)

func RunAIBackendReachable() (bool, error) {
	saas := env.EnvOrDefault(env.RunAISaasEnvVar, RunAISaasAddress)

	res, err := http.Get(saas)
	if err != nil {
		return false, err
	}

	if res.StatusCode >= 300 && res.StatusCode < 200 {
		return false, fmt.Errorf("status code of %d was returned from %s", res.StatusCode, saas)
	}

	return true, nil
}

func RunAIAuthProviderReachable() (bool, error) {
	const auth = "https://runai-prod.auth0.com"

	return utils.CheckURLAvailable(auth)
}
