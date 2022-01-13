package main

import (
	"flag"
	"os"

	"github.com/run-ai/preinstall-diagnostics/internal/cmd/cli"
	"github.com/run-ai/preinstall-diagnostics/internal/cmd/daemonset"
	"github.com/run-ai/preinstall-diagnostics/internal/env"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"github.com/run-ai/preinstall-diagnostics/internal/registry"
	"github.com/run-ai/preinstall-diagnostics/internal/resources"
	"github.com/run-ai/preinstall-diagnostics/internal/saas"
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
	daemonsetModeStr, _ := env.EnvOrError(resources.DamonsetModeEnvName)
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
		daemonset.DaemonsetMain()
	} else {
		cli.CliMain(clean, dryRun, backendFQDN, image, runaiContainerRegistry, runaiSaas, outputFile)
	}
}
