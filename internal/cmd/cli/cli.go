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
)

var (
	externalClusterTests = []func(*log.Logger) error{
		cluster.ShowClusterVersion,
		cluster.ShowGPUNodes, // gpu-feature-discovery // NVIDIA pre-requisites on GPU nodes. (gpu-feature-discovery checks this)
		cluster.NvidiaDevicePluginAvailable,
		cluster.DCGMExporterAvailable,
		cluster.PrometheusAvailable,
		cluster.NodeFeatureDiscoveryAvailable,
		cluster.ShowStorageClasses,
		cluster.NginxIngressControllerAvailable,
	}
)

func CliMain(clean, dryRun bool, backendFQDN, image, imageRegistry, runaiSaas string, logger *log.Logger) {
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

	fmt.Println("cleaning up previous deployment if it exists...")
	err = util.DeleteResources(client, dynClient)
	if err != nil {
		panic(err)
	}

	if clean {
		return
	}

	fmt.Println("deploying runai diagnostics tool...")
	err = util.CreateResources(backendFQDN, dynClient)
	if err != nil {
		panic(err)
	}

	fmt.Println("running external cluster tests...")
	errs := util.RunTests(externalClusterTests, logger)

	fmt.Printf("running internal cluster tests using image %s...\n", image)
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

		_, err = logger.WriteString(podLogs)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("cleaning up...")
	err = util.DeleteResources(client, dynClient)
	if err != nil {
		panic(err)
	}

	if len(errs) > 0 || podLogsHaveFailures {
		log.NewLogger(os.Stdout).ErrorF("some tests have failed, see above logs or output file for more information")
	}
}
