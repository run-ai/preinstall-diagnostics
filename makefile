OUT_DIR=_out
BIN=preinstall-diagnostics

REGISTRY=gcr.io/run-ai-lab
VERSION=v0.1.2

IMAGE_DOCKER_FILE=cmd/preinstall-diagnostics/Dockerfile
IMAGE=$(REGISTRY)/preinstall-diagnostics:$(VERSION)

BUILDER_IMAGE_DOCKER_FILE=cmd/preinstall-diagnostics/builder-image.Dockerfile
BUILDER_IMAGE=$(REGISTRY)/preinstall-diagnostics-builder:$(VERSION)

.PHONY: binary
binary:
	IMAGE=${IMAGE} OUT_DIR=$(OUT_DIR) BIN=$(BIN) VERSION=$(VERSION) ./scripts/build-binary.sh

.PHONY: builder-image
builder-image:
	DOCKER_FILE=$(BUILDER_IMAGE_DOCKER_FILE) IMAGE=$(BUILDER_IMAGE) ./scripts/build-image.sh

.PHONY: image
image:
	DOCKER_FILE=$(IMAGE_DOCKER_FILE) IMAGE=$(IMAGE) ./scripts/build-image.sh

.PHONY: push-builder
push-builder:
	docker push $(BUILDER_IMAGE)

.PHONY: push
push:
	docker push $(IMAGE)

.PHONY: all
all: builder-image push-builder image push binary