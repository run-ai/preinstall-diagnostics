package saas

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/run-ai/preinstall-diagnostics/internal/env"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
)

const (
	RunAISaasHostname = "app.run.ai"
	RunAISaasAddress  = "https://" + RunAISaasHostname
	DynuDNSService    = "https://api.dynu.com"
)

func CheckRunAIBackendReachable(logger *log.Logger) error {
	logger.TitleF("Run:AI service access: %s", RunAISaasAddress)
	saas := env.EnvOrDefault(env.RunAISaasEnvVar, RunAISaasAddress)

	res, err := http.Get(saas)
	if err != nil {
		return err
	}

	if res.StatusCode >= 300 && res.StatusCode < 200 {
		return fmt.Errorf("status code of %d was returned from %s", res.StatusCode, saas)
	}

	logger.LogF("Run:AI service is accessible from within the cluster")
	return nil
}

func DNSServersReachable(logger *log.Logger) error {
	logger.TitleF("DNS Servers access")

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
			return err
		}

		if len(ip) == 0 {
			return fmt.Errorf("no addresses resolved for [%s] by [%s]",
				RunAISaasHostname, dnsServer)
		}

		logger.LogF("Address for [%s] is [%s], resolved by [%s]",
			RunAISaasHostname, ip[0], dnsServer)
	}

	return nil
}

func DynuReachable(logger *log.Logger) error {
	logger.TitleF("Dynu DNS service access: %s", DynuDNSService)

	res, err := http.Get(DynuDNSService)
	if err != nil {
		return err
	}

	if res.StatusCode >= 300 && res.StatusCode < 200 {
		return fmt.Errorf("status code of %d was returned from %s", res.StatusCode, DynuDNSService)
	}

	logger.LogF("Dynu DNS service is accessible from within the cluster")
	return nil
}
