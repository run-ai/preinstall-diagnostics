# Builder image name can be found in the makefile
ARG BUILDER_IMAGE=notavailable
FROM ${BUILDER_IMAGE} AS builder

# Set up env for go compiling
ENV GOPATH=/home/preinstall-diagnostics/go
ENV PROJECT_PATH=${GOPATH}/src/github.com/run-ai/preinstall-diagnostics
RUN mkdir -p ${PROJECT_PATH}
WORKDIR ${PROJECT_PATH}

# Copy all code
COPY . .

RUN make binary-linux

FROM fedora

ENV GOPATH=/home/preinstall-diagnostics/go
ENV PROJECT_PATH=${GOPATH}/src/github.com/run-ai/preinstall-diagnostics

COPY --from=builder ${PROJECT_PATH}/_out/preinstall-diagnostics-linux-amd64 /preinstall-diagnostics

ENTRYPOINT [ "/preinstall-diagnostics" ]