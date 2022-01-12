# Run:AI Preinstall Diagnostics Tool

## Diagnostics tool for Run:AI installations
This tool is used prior to Run:AI installations to collect cluster information to
help fine-tune the installation process

## Build
```
REGISTRY=gcr.io/run-ai-lab VERSION=<your-private-tag> make -e all
```

## Running at client environments
- The binary should be provided to the customer according to their PC's platform (windows/linux/apple x86/ apple m1), the server platform is assumed to be `x86_64`
- The container image should be provided to the customer as well:
    * If the client environment is air-gapped, the image should be provided to them offline and pushed to their internal registry, and the binary should be run with `--image` to modify the image used.
    * If the client is connected to the internet, the proper image will be pulled from the `run-ai-prod` registry.

## Help
```
❯ ./_out/preinstall-diagnostics-darwin-arm64 --help
Usage of ./_out/preinstall-diagnostics-darwin-arm64:
  -backend-fqdn string
    	FQDN of the runai backend to resolve (required for DNS resolve test)
  -clean
    	Clean all runai diagnostics tools from the cluster
  -dry-run
    	Print the diagnostics resources without executing
  -image string
    	Diagnostics image to use (for air-gapped environments) (default "gcr.io/run-ai-prod/preinstall-diagnostics:v0.1.0") # Version might varys
  -output string
    	File to save the output to, if omitted, prints to stdout
  -registry string
    	URL to container image registry to check connectivity to (default "https://gcr.io/run-ai-prod")
  -saas-address string
    	URL the Run:AI service to check connectivity to (default "https://app.run.ai")
```