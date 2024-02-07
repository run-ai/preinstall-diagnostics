package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/run-ai/preinstall-diagnostics/internal/client"
	"github.com/run-ai/preinstall-diagnostics/internal/cluster"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"github.com/run-ai/preinstall-diagnostics/internal/resources"
	"github.com/run-ai/preinstall-diagnostics/internal/util"
	ver "github.com/run-ai/preinstall-diagnostics/internal/version"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var (
	externalClusterTests = []func(*log.Logger) error{
		// cluster.NvidiaDevicePluginNotInstalled,
		// cluster.DCGMExporterNotInstalled,
		cluster.ShowClusterVersion,
		cluster.ListPods,
		cluster.ShowGPUNodes,
		cluster.NGINXIngressControllerInstalled,
		cluster.PrometheusInstalled,
		cluster.StorageClassExists,
		cluster.RunAIHelmRepositoryReachable,
	}
)

func CliMain(clean, dryRun bool, backendFQDN, image, imagePullSecretName, imageRegistry,
	runaiSaas string, version bool, logger *log.Logger) {
	if version {
		fmt.Println(ver.Version)
		return
	}

	util.TemplateResources(backendFQDN, image, imagePullSecretName, imageRegistry, runaiSaas)

	if dryRun {
		err := resources.PrintResources(resources.CreationOrder())
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		return
	}

	err := client.Init(logger)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	dynClient, err := client.DynamicClient()
	if err != nil {
		panic(err)
	}

	k8s, err := client.ClientSet()
	if err != nil {
		panic(err)
	}

	_, _ = logger.WriteStringF("cleaning up previous deployment if it exists...")
	err = util.DeleteResources(k8s, dynClient, logger)
	if err != nil {
		panic(err)
	}

	if clean {
		return
	}

	_, _ = logger.WriteStringF("deploying runai diagnostics tool...")
	err = util.CreateResources(backendFQDN, dynClient)
	if err != nil {
		panic(err)
	}

	logger.TitleF("running external cluster tests...")
	errs := util.RunTests(externalClusterTests, logger)

	_, _ = logger.WriteStringF("")

	logger.TitleF("running internal cluster tests using image %s...", image)
	err = cluster.WaitDaemonSetAvailable(logger)
	if err != nil {
		panic(err)
	}

	pods, err := cluster.GetDaemonsetPods(k8s)
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

		_, _ = logger.WriteStringF("")
		_, _ = logger.WriteStringF("========================== LOGS FROM NODE %s ==========================",
			pod.Spec.NodeName)
		_, _ = logger.WriteStringF("")

		_, err = logger.WriteStringF(podLogs)
		if err != nil {
			panic(err)
		}
	}

	_, _ = logger.WriteStringF("cleaning up...")
	err = util.DeleteResources(k8s, dynClient, logger)
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
