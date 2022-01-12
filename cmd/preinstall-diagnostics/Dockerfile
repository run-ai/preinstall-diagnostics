FROM gcr.io/run-ai-prod/preinstall-diagnostics-builder:v0.1.0 AS builder

# Set up env for go compiling
ENV GOPATH=/home/preinstall-diagnostics/go
ENV PROJECT_PATH=${GOPATH}/src/github.com/run-ai/preinstall-diagnostics
RUN mkdir -p ${PROJECT_PATH}
WORKDIR ${PROJECT_PATH}

# Copy all code
COPY . .

RUN make binary

FROM fedora

ENV GOPATH=/home/preinstall-diagnostics/go
ENV PROJECT_PATH=${GOPATH}/src/github.com/run-ai/preinstall-diagnostics

COPY --from=builder ${PROJECT_PATH}/_out/preinstall-diagnostics-linux-amd64 /preinstall-diagnostics

ENTRYPOINT [ "/preinstall-diagnostics" ]