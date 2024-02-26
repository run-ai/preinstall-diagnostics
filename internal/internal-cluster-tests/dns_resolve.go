package internal_cluster_tests

import (
	"context"
	"fmt"
	"github.com/run-ai/preinstall-diagnostics/internal/env"
	"io"
	"net"
	"os"
	"time"
)

func DNSRecordsResolvable() (bool, error) {
	externalDNSServers := []string{
		"1.1.1.1:53",
		"8.8.8.8:53",
	}

	for _, dnsServer := range externalDNSServers {
		resolver := &net.Resolver{
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Millisecond * time.Duration(10000),
				}
				return d.DialContext(ctx, network, dnsServer)
			},
		}

		ip, err := resolver.LookupHost(context.Background(), RunAISaasHostname)
		if err != nil {
			return false, err
		}

		if len(ip) == 0 {
			return false, fmt.Errorf("no addresses resolved for [%s] by [%s]",
				RunAISaasHostname, dnsServer)
		}
	}

	return true, nil
}

func BackendFQDNResolvable() ([]net.IP, error) {
	backendFQDN := env.EnvOrDefault(env.BackendFQDNEnvVar, "")
	if backendFQDN == "" {
		return nil, fmt.Errorf("Backend FQDN was not provided using the --domain flag, skipping test")
	}

	ips, err := net.DefaultResolver.LookupIP(context.TODO(), "ip", backendFQDN)
	if err != nil {
		return nil, err
	}

	return ips, nil
}

func DNSResolvConf() (string, error) {
	f, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return "", err
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
