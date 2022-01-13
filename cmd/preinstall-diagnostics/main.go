package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/run-ai/preinstall-diagnostics/internal/client"
	"github.com/run-ai/preinstall-diagnostics/internal/cluster"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"github.com/run-ai/preinstall-diagnostics/internal/registry"
	"github.com/run-ai/preinstall-diagnostics/internal/resources"
	"github.com/run-ai/preinstall-diagnostics/internal/saas"
	"github.com/run-ai/preinstall-diagnostics/internal/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

var (
	externalClusterTests = []func() error{
		cluster.ShowClusterVersion,
		cluster.ShowGPUNodes, // gpu-feature-discovery // NVIDIA pre-requisites on GPU nodes. (gpu-feature-discovery checks this)
		cluster.NvidiaDevicePluginAvailable,
		cluster.DCGMExporterAvailable,
		cluster.PrometheusAvailable,
		cluster.NodeFeatureDiscoveryAvailable,
		cluster.ShowStorageClasses,
		cluster.NginxIngressControllerAvailable,
	}

	internalClusterTests = []func() error{
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

const (
	cleanArgName                  = "clean"
	backendFQDNArgName            = "backend-fqdn"
	imageArgName                  = "image"
	dryRunArgName                 = "dry-run"
	runaiContainerRegistryArgName = "registry"
	runaiSaasArgName              = "saas-address"
	outputArgName                 = "output"
)

var (
	daemonsetMode bool

	clean                  bool
	backendFQDN            string
	image                  string
	dryRun                 bool
	runaiContainerRegistry string
	runaiSaas              string
	output                 string

	outputFile *os.File
)

func init() {
	daemonsetModeStr, _ := util.EnvOrError(resources.DamonsetModeEnvName)
	daemonsetMode = daemonsetModeStr != ""

	flag.BoolVar(&clean, cleanArgName, false, "Clean all runai diagnostics tools from the cluster")
	flag.StringVar(&backendFQDN, backendFQDNArgName, "", "FQDN of the runai backend to resolve (required for DNS resolve test)")
	flag.StringVar(&image, imageArgName, registry.RunAIDiagnosticsImage, "Diagnostics image to use (for air-gapped environments)")
	flag.BoolVar(&dryRun, dryRunArgName, false, "Print the diagnostics resources without executing")
	flag.StringVar(&runaiContainerRegistry, runaiContainerRegistryArgName, registry.RunAIProdRegistryURL, "URL to container image registry to check connectivity to")
	flag.StringVar(&runaiSaas, runaiSaasArgName, saas.RunAISaasAddress, "URL the Run:AI service to check connectivity to")
	flag.StringVar(&output, outputArgName, "", "File to save the output to, if omitted, prints to stdout")
	flag.Parse()

	if output != "" {
		err := os.RemoveAll(output)
		if err != nil {
			panic(err)
		}

		outputFile, err = os.OpenFile(output,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		log.LogTarget = outputFile
	}
}

func main() {
	if daemonsetMode {
		daemonsetMain()
	} else {
		cliMain(clean, backendFQDN, image, runaiContainerRegistry, runaiSaas)
	}
}

func runTests(tests []func() error) {
	for _, test := range tests {
		err := test()
		if err != nil {
			log.ErrorF("%v", err)
			log.Fail()
		} else {
			log.Pass()
		}
	}
}

func cliMain(clean bool, backendFQDN, image, imageRegistry, runaiSaas string) {
	templateResources(backendFQDN, image, imageRegistry, runaiSaas)

	if dryRun {
		resources.PrintResources(resources.CreationOrder())
		return
	}

	dynClient, err := client.DynamicClient()
	if err != nil {
		panic(err)
	}

	client, err := client.Clientset()
	if err != nil {
		panic(err)
	}

	fmt.Println("cleaning up previous deployment if it exists...")
	err = deleteResources(dynClient)
	if err != nil {
		panic(err)
	}

	if clean {
		return
	}

	fmt.Println("deploying runai diagnostics tool...")
	err = createResources(backendFQDN, dynClient)
	if err != nil {
		panic(err)
	}

	fmt.Println("running external cluster tests...")
	runTests(externalClusterTests)

	fmt.Print()

	fmt.Println("running internal cluster tests...")
	err = cluster.WaitDaemonSetAvailable()
	if err != nil {
		panic(err)
	}

	pods, err := cluster.GetDaemonsetPods(client)
	if err != nil {
		panic(err)
	}

	for _, pod := range pods {
		podLogs, err := cluster.GetCompletePodLogs(&pod)
		if err != nil {
			panic(err)
		}

		if output != "" {
			_, err = outputFile.WriteString(podLogs)
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Println(podLogs)
		}
	}

	fmt.Println("cleaning up...")
	err = deleteResources(dynClient)
	if err != nil {
		panic(err)
	}
}

func templateResources(backendFQDN, image, imageRegistry, runaiSaas string) {
	ds := &resources.DaemonSet
	if backendFQDN != "" {
		ds.Spec.Template.Spec.Containers[0].Env =
			append(ds.Spec.Template.Spec.Containers[0].Env,
				v1.EnvVar{
					Name:  cluster.BackendFQDNEnvVar,
					Value: backendFQDN,
				})
	}

	if imageRegistry != "" {
		ds.Spec.Template.Spec.Containers[0].Env =
			append(ds.Spec.Template.Spec.Containers[0].Env,
				v1.EnvVar{
					Name:  registry.RegistryEnvVar,
					Value: imageRegistry,
				})
	}

	if runaiSaas != "" {
		ds.Spec.Template.Spec.Containers[0].Env =
			append(ds.Spec.Template.Spec.Containers[0].Env,
				v1.EnvVar{
					Name:  saas.RunAISaasEnvVar,
					Value: runaiSaas,
				})
	}

	if image != "" {
		ds.Spec.Template.Spec.Containers[0].Image = image
	}
}

func deleteResources(dynClient dynamic.Interface) error {
	err := resources.DeleteResources(resources.DeletionOrder(), dynClient)
	if err != nil {
		return err
	}

	fmt.Println("waiting for all resources to be deleted...")

	// wait for ns to be deleted
	client, err := client.Clientset()
	if err != nil {
		return err
	}

	for {
		_, err := client.CoreV1().Namespaces().Get(context.TODO(),
			resources.Namespace.Name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				break
			}

			return err
		}

		time.Sleep(time.Second)
	}

	fmt.Println("all resources were successfully deleted")
	return nil
}

func createResources(backendFQDN string, dynClient dynamic.Interface) error {
	return resources.CreateResources(resources.CreationOrder(), dynClient)
}

func daemonsetMain() {
	runTests(internalClusterTests)
	log.Complete()

	// Sleep
	<-make(chan struct{})
}
