package cli

import (
	"fmt"
	"strings"

	"github.com/run-ai/preinstall-diagnostics/internal/client"
	"github.com/run-ai/preinstall-diagnostics/internal/cluster"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"github.com/run-ai/preinstall-diagnostics/internal/resources"
	"github.com/run-ai/preinstall-diagnostics/internal/util"
	ver "github.com/run-ai/preinstall-diagnostics/internal/version"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	externalClusterTests = []func(*log.Logger) error{
		cluster.ShowGPUNodes, // gpu-feature-discovery // NVIDIA pre-requisites on GPU nodes. (gpu-feature-discovery checks this)
		cluster.NvidiaDevicePluginNotInstalled,
		cluster.DCGMExporterNotInstalled,
		cluster.NginxIngressControllerNotInstalled,
		cluster.ShowClusterVersion,
		cluster.ShowStorageClasses,
		cluster.PrometheusNotInstalled,
		cluster.NodeFeatureDiscoveryNotInstalled,
		cluster.GPUFeatureDiscoveryNotInstalled,
	}
)

func CliMain(clean, dryRun bool, backendFQDN, image, imageRegistry,
	runaiSaas string, version bool, logger *log.Logger) {
	if version {
		fmt.Println(ver.Version)
		return
	}

	util.TemplateResources(backendFQDN, image, imageRegistry, runaiSaas)

	if dryRun {
		resources.PrintResources(resources.CreationOrder())
		return
	}

	client.Init(logger)

	dynClient, err := client.DynamicClient()
	if err != nil {
		panic(err)
	}

	client, err := client.Clientset()
	if err != nil {
		panic(err)
	}

	logger.WriteStringF("cleaning up previous deployment if it exists...")
	err = util.DeleteResources(client, dynClient, logger)
	if err != nil {
		panic(err)
	}

	if clean {
		return
	}

	logger.WriteStringF("deploying runai diagnostics tool...")
	err = util.CreateResources(backendFQDN, dynClient)
	if err != nil {
		panic(err)
	}

	logger.TitleF("running external cluster tests...")
	errs := util.RunInformationalTests(externalClusterTests, logger)

	logger.WriteStringF("")

	logger.TitleF("running internal cluster tests using image %s...", image)
	err = cluster.WaitDaemonSetAvailable(logger)
	if err != nil {
		panic(err)
	}

	pods, err := cluster.GetDaemonsetPods(client)
	if err != nil {
		panic(err)
	}

	var podLogsHaveFailures bool
	for _, pod := range pods {
		podLogs, err := cluster.GetCompletePodLogs(&pod, logger)
		if err != nil {
			panic(err)
		}

		if strings.Contains(podLogs, log.FailTag) {
			podLogsHaveFailures = true
		}

		logger.WriteStringF("")
		logger.WriteStringF("========================== LOGS FROM NODE %s ==========================",
			pod.Spec.NodeName)
		logger.WriteStringF("")

		_, err = logger.WriteStringF(podLogs)
		if err != nil {
			panic(err)
		}
	}

	logger.WriteStringF("cleaning up...")
	err = util.DeleteResources(client, dynClient, logger)
	if err != nil {
		panic(err)
	}

	if len(errs) > 0 {
		logger.WarningF("Cluster setup includes components that will " +
			"require the customization of Run:AI installation. For more details, see installation instructions")
	}

	if podLogsHaveFailures {
		logger.ErrorF("Not all prerequisites have been met, please provide %s to Run:AI Customer Support.",
			logger.FileName())
	}
}
