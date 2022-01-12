go mod tidy
go mod vendor

# Building a Windows binary
GOOS=windows GOARCH=amd64 go build \
    -ldflags="-X 'github.com/run-ai/preinstall-diagnostics/internal/registry.RunAIDiagnosticsImage=${IMAGE}'" \
    -o ${OUT_DIR}/${BIN}-windows-amd64 cmd/preinstall-diagnostics/main.go

# Building a MacOS x86 binary
GOOS=darwin GOARCH=amd64 go build \
    -ldflags="-X 'github.com/run-ai/preinstall-diagnostics/internal/registry.RunAIDiagnosticsImage=${IMAGE}'" \
    -o ${OUT_DIR}/${BIN}-darwin-amd64 cmd/preinstall-diagnostics/main.go

# Building a MacOS arm binary
GOOS=darwin GOARCH=arm64 go build \
    -ldflags="-X 'github.com/run-ai/preinstall-diagnostics/internal/registry.RunAIDiagnosticsImage=${IMAGE}'" \
    -o ${OUT_DIR}/${BIN}-darwin-arm64 cmd/preinstall-diagnostics/main.go

#Building a Linux binary
GOOS=linux GOARCH=amd64 go build \
    -ldflags="-X 'github.com/run-ai/preinstall-diagnostics/internal/registry.RunAIDiagnosticsImage=${IMAGE}'" \
    -o ${OUT_DIR}/${BIN}-linux-amd64 cmd/preinstall-diagnostics/main.go