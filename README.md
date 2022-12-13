# Run:AI Preinstall Diagnostics Tool

## Diagnostics tool for Run:AI installations
This tool is used prior to Run:AI installations to collect cluster information to
help fine-tune the installation process

## Build
  ### Production
  #### Build and push
  ```
  make -e all
  ```

### Development
  #### Build and push
  ```
  REGISTRY=gcr.io/run-ai-lab VERSION=<your-private-tag> make -e all
  ```
  #### Building the binary
  ```
  REGISTRY=gcr.io/run-ai-lab VERSION=<your-private-tag> make -e binary
  ```

  #### Building the container image
  ```
  REGISTRY=gcr.io/run-ai-lab VERSION=<your-private-tag> make -e builder-image image
  ```

  #### Pushing the image
  ```
  REGISTRY=gcr.io/run-ai-lab VERSION=<your-private-tag> make -e push-image
  ```

  #### Updating the builder image (in case some new build tools are needed)
  `cmd/preinstall-diagnostics/builder-image.Dockerfile` has to be updated to install the new required tools, once it is updated, the builder image can be built and pushed:
  ```
  REGISTRY=gcr.io/run-ai-lab VERSION=<your-private-tag> make -e builder-image push-builder
  ```

## Running at client environments
  - The binary should be provided to the customer according to their PC's platform (windows/linux/apple x86/ apple m1), the server platform is assumed to be `x86_64`
  - The container image should be provided to the customer as well:
      * If the client environment is air-gapped, the image should be provided to them offline and pushed to their internal registry, and the binary should be run with `--image` to modify the image used:
        * Save the image locally
          ```
          docker save --output preinstall-diagnostics.tar gcr.io/run-ai-lab/preinstall-diagnostics:${VERSION}
          ```

        * Load the image to the client's internal registry
        
          Once the archive was provided to the user, they can load the image to their internal registry using docker:
          ```
          docker load --input preinstall-diagnostics.tar
          docker tag gcr.io/run-ai-lab/preinstall-diagnostics:${VERSION} ${CLIENT_IMAGE_AND_TAG}
          docker push ${CLIENT_IMAGE_AND_TAG}
          ```
        * Running the diagnostics tool with the internal image
          ```
          preinstall-diagnostics-linux-amd64 --image ${CLIENT_IMAGE_AND_TAG}
          ```
          - In order for the tool to work properly with EKS, please run `aws configure` prior to execution
      * If the client is connected to the internet, the proper image will be pulled from the `run-ai-prod` registry.

## Help
```
❯ ./_out/preinstall-diagnostics-darwin-arm64 --help
Usage of ./_out/preinstall-diagnostics-darwin-arm64:
  -clean
    	Clean all runai diagnostics tools from the cluster
  -domain string
    	FQDN of the runai backend to resolve (required for DNS resolve test)
  -dry-run
    	Print the diagnostics resources without executing
  -image string
    	Diagnostics image to use (for air-gapped environments) (default "gcr.io/run-ai-lab/preinstall-diagnostics:v2.4.0")
  -output string
    	File to save the output to (default "runai-preinstall-diagnostics.txt")
  -registry string
    	URL to container image registry to check connectivity to (default "https://gcr.io/run-ai-prod")
  -saas-address string
    	URL the Run:AI service to check connectivity to (default "https://app.run.ai")
  -version
    	Prints the binary version
```