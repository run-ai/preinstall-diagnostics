package internal_cluster_tests

import (
	"github.com/run-ai/preinstall-diagnostics/internal/utils"
)

func RunAIPrometheusReachable() (bool, error) {
	const prom = "https://prometheus-us-central1.grafana.net"

	return utils.CheckURLAvailable(prom)
}
