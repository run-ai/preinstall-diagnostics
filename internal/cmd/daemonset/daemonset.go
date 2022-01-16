package daemonset

import (
	"github.com/run-ai/preinstall-diagnostics/internal/client"
	"github.com/run-ai/preinstall-diagnostics/internal/cluster"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"github.com/run-ai/preinstall-diagnostics/internal/registry"
	"github.com/run-ai/preinstall-diagnostics/internal/saas"
	"github.com/run-ai/preinstall-diagnostics/internal/util"
)

var (
	internalClusterTests = []func(*log.Logger) error{
		saas.CheckRunAIBackendReachable,
		registry.CheckRunAIRegistryReachable,
		cluster.CheckDNSResolve, // DNS: verify that the DNS provided as input can be resolved to an IP (optional, write/warn if the IP is one of the nodes)
		cluster.GetResolvConf,
		cluster.RunAIHelmRepositoryReachable,
		cluster.DockerHubReachable,
		cluster.QuayIOReachable,
		cluster.RunAIPrometheusReachable,
		cluster.RunAIAuthProviderReachable, // verify post-install URLs: Prerequisites
		// Rancher requires special instructions Troubleshooting
		cluster.ShowOSInfo, // Advanced: The operating system of each node (NVIDIA drivers were not working due to an unsupported Ubuntu version. )
		// Advanced: Tanzu and any Kubernetes 1.22 and up requires systemd.
		cluster.CheckNodeConnectivity, // clock sync.
	}
)

func DaemonsetMain(logger *log.Logger) {
	client.Init(logger)
	util.RunTests(internalClusterTests, logger)
	logger.Complete()

	// Sleep
	<-make(chan struct{})
}
