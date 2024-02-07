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
	domainArgName                 = "domain"
	imageArgName                  = "image"
	imagePullSecretArgName        = "image-pull-secret"
	dryRunArgName                 = "dry-run"
	runaiContainerRegistryArgName = "registry"
	runaiSaasArgName              = "saas-address"
	outputArgName                 = "output"
	versionArgName                = "version"
)

const (
	defaultOutputFileName = "runai-preinstall-diagnostics.txt"
)

var (
	daemonsetMode bool

	clean                  bool
	domainFQDN             string
	image                  string
	imagePullSecretName    string
	dryRun                 bool
	runaiContainerRegistry string
	runaiSaas              string
	output                 string
	version                bool

	outputFile *os.File
)

func init() {
	daemonsetModeStr, _ := env.EnvOrError(resources.DamonsetModeEnvName)
	daemonsetMode = daemonsetModeStr != ""

	flag.BoolVar(&clean, cleanArgName, false, "Clean all runai diagnostics tools from the cluster")
	flag.StringVar(&domainFQDN, domainArgName, "", "FQDN of the runai backend to resolve (required for DNS resolve test)")
	flag.StringVar(&image, imageArgName, registry.RunAIDiagnosticsImage, "Diagnostics image to use (for air-gapped environments)")
	flag.StringVar(&imagePullSecretName, imagePullSecretArgName, "", "Secret name (within the 'runai-preinstall-diagnostics' namespace) that contains container-registry credentials")
	flag.BoolVar(&dryRun, dryRunArgName, false, "Print the diagnostics resources without executing")
	flag.StringVar(&runaiContainerRegistry, runaiContainerRegistryArgName, registry.RunAIProdRegistryURL, "URL to container image registry to check connectivity to")
	flag.StringVar(&runaiSaas, runaiSaasArgName, saas.RunAISaasAddress, "URL the Run:AI service to check connectivity to")
	flag.StringVar(&output, outputArgName, defaultOutputFileName, "File to save the output to")
	flag.BoolVar(&version, versionArgName, false, "Prints the binary version")

	flag.Parse()
}

func main() {
	if daemonsetMode {
		daemonset.DaemonsetMain(log.NewLogger(nil))
	} else {
		err := os.RemoveAll(output)
		if err != nil {
			panic(err)
		}

		outputFile, err = os.OpenFile(output,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		logger := log.NewLogger(outputFile)

		cli.CliMain(clean, dryRun, domainFQDN, image, imagePullSecretName, runaiContainerRegistry, runaiSaas, version, logger)
	}
}
