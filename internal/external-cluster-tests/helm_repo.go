package external_cluster_tests

import (
	"github.com/run-ai/preinstall-diagnostics/internal/utils"
)

func RunAIHelmRepositoryReachable() (bool, error) {
	const runaiCharts = "https://run-ai-charts.storage.googleapis.com"

	available, err := utils.CheckURLAvailable(runaiCharts)
	if err != nil {
		return false, err
	}

	return available, nil
}
