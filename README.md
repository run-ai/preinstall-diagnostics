# Run:AI Preinstall Diagnostics Tool

## Diagnostics tool for Run:AI installations
This tool is used prior to Run:AI installations to collect cluster information to
help fine-tune the installation process

## Usage

### Run:ai SaaS
- In order for the tool to work properly with EKS, please run `aws configure` prior to execution
```shell
chmod +x ./preinstall-diagnostics-darwin-arm64 && \
  ./preinstall-diagnostics-darwin-arm64 \
      --domain ${TENANT_NAME}.run.ai \
      --cluster-domain ${CLUSTER_FQDN}
```

### Self-Hosted deployment
```shell
chmod +x ./preinstall-diagnostics-darwin-arm64 && \ 
  ./preinstall-diagnostics-darwin-arm64 \
      --domain ${CONTROL_PLANE_FQDN} \
      --cluster-domain ${CLUSTER_FQDN} \
  # if the diagnostics image is hosted in a private registry
      --image-pull-secret ${IMAGE_PULL_SECRET_NAME} \
      --image ${PRIVATE_REGISTRY_IMAGE_URL}    
```

### Air-gapped deployment
- On an air-gapped deployment, the diagnostics image should be pulled, saved and manually pushed to the organization's registry.
- The binary should be run with `--image` parameter to modify the diagnostics image to be used:

  * Save the image locally
    ```
    docker save --output preinstall-diagnostics.tar gcr.io/run-ai-lab/preinstall-diagnostics:${VERSION}
    ```
  * Load the image to the organization's registry
     ```
     docker load --input preinstall-diagnostics.tar
     docker tag gcr.io/run-ai-lab/preinstall-diagnostics:${VERSION} ${CLIENT_IMAGE_AND_TAG} 
     docker push ${CLIENT_IMAGE_AND_TAG}
     ```
Finally, run the diagnostics tool:
```shell
chmod +x ./preinstall-diagnostics-darwin-arm64 && \ 
  ./preinstall-diagnostics-darwin-arm64 \
      --domain ${CONTROL_PLANE_FQDN} \
      --cluster-domain ${CLUSTER_FQDN} \
      --image-pull-secret ${IMAGE_PULL_SECRET_NAME} \
      --image ${PRIVATE_REGISTRY_IMAGE_URL}    
```

### Example

```shell
chmod +x ./preinstall-diagnostics-darwin-arm64 && \ 
  ./preinstall-diagnostics-darwin-arm64 \
      --domain my-control-plane.my-org.com \
      --cluster-domain my-cluster.my-org.com \
      # if the diagnostics image is hosted in a private registry
      --image-pull-secret image-pull-secret \
      --image gcr.io/my-org/containers/preinstall-diagnostics:v2.16    
```

## Help
```
❯ ./preinstall-diagnostics-darwin-arm64 --help
Usage of ./preinstall-diagnostics-darwin-arm64:
  -clean
    	Clean all runai diagnostics tools from the cluster
  -cluster-domain string
    	FQDN of the cluster
  -domain string
    	FQDN of the runai backend to resolve (required for DNS resolve test)
  -dry-run
    	Print the diagnostics resources without executing
  -image string
    	Diagnostics image to use (for air-gapped environments) (default "gcr.io/run-ai-lab/preinstall-diagnostics:v2.16.18")
  -image-pull-secret string
    	Secret name (within the 'runai-preinstall-diagnostics' namespace) that contains container-registry credentials
  -kubeconfig string
    	Paths to a kubeconfig. Only required if out-of-cluster.
  -output string
    	File to save the output to (default "runai-preinstall-diagnostics.txt")
  -registry string
    	URL to container image registry to check connectivity to (default "https://gcr.io/run-ai-prod")
  -saas-address string
    	URL the Run:AI service to check connectivity to (default "https://app.run.ai")
  -version
    	Prints the binary version
```

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
