package job

import (
	v2 "github.com/run-ai/preinstall-diagnostics/internal"
	"github.com/run-ai/preinstall-diagnostics/internal/env"
	internal_cluster_tests2 "github.com/run-ai/preinstall-diagnostics/internal/internal-cluster-tests"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
)

func runTestsAndAppendResults(logger *log.Logger) []v2.TestResult {
	var testResults []v2.TestResult
	osInfo, err := internal_cluster_tests2.ShowOSInfo()
	if err != nil {
		testResults = append(testResults, v2.TestResult{
			Name:    "OS Info",
			Result:  false,
			Message: err.Error(),
		})
	} else {
		testResults = append(testResults, v2.TestResult{
			Name:    "OS Info",
			Result:  true,
			Message: osInfo,
		})
	}

	ips, err := internal_cluster_tests2.BackendFQDNResolvable()
	if err != nil {
		testResults = append(testResults, v2.TestResult{
			Name:    "Backend FQDN Resolve",
			Result:  false,
			Message: err.Error(),
		})
	} else {
		ipsStr := ""
		for i := range ips {
			ipsStr += ips[i].String()
			if i < len(ips)-1 {
				ipsStr += "\n"
			}
		}
		testResults = append(testResults, v2.TestResult{
			Name:    "Backend FQDN Resolve",
			Result:  true,
			Message: ipsStr,
		})
	}

	dnsResolveConf, err := internal_cluster_tests2.DNSResolvConf()
	if err != nil {
		testResults = append(testResults, v2.TestResult{
			Name:    "DNS Resolve Conf",
			Result:  false,
			Message: err.Error(),
		})
	} else {
		testResults = append(testResults, v2.TestResult{
			Name:    "DNS Resolve Conf",
			Result:  true,
			Message: dnsResolveConf,
		})
	}

	airgapped := env.EnvOrDefault(env.AirgappedEnvVar, "false")
	logger.WriteStringF("airgapped = " + airgapped)
	//if airgapped != "true" {
	reachable, err := internal_cluster_tests2.RunAIAuthProviderReachable()
	if err != nil {
		testResults = append(testResults, v2.TestResult{
			Name:    "RunAI Auth Provider Reachable",
			Result:  false,
			Message: err.Error(),
		})
	} else {
		testResults = append(testResults, v2.TestResult{
			Name:    "RunAI Auth Provider Reachable",
			Result:  reachable,
			Message: "",
		})
	}

	reachable, err = internal_cluster_tests2.RunAIPrometheusReachable()
	if err != nil {
		testResults = append(testResults, v2.TestResult{
			Name:    "RunAI Prometheus Reachable",
			Result:  false,
			Message: err.Error(),
		})
	} else {
		testResults = append(testResults, v2.TestResult{
			Name:    "RunAI Prometheus Reachable",
			Result:  reachable,
			Message: "",
		})
	}

	reachable, err = internal_cluster_tests2.RunAIBackendReachable()
	if err != nil {
		testResults = append(testResults, v2.TestResult{
			Name:    "RunAI Backend Reachable",
			Result:  false,
			Message: err.Error(),
		})
	} else {
		testResults = append(testResults, v2.TestResult{
			Name:    "RunAI Backend Reachable",
			Result:  reachable,
			Message: "",
		})
	}
	//}
	err = internal_cluster_tests2.CheckNodeConnectivity(logger)
	if err != nil {
		testResults = append(testResults, v2.TestResult{
			Name:    "Node Connectivity",
			Result:  false,
			Message: err.Error(),
		})
	} else {
		testResults = append(testResults, v2.TestResult{
			Name:    "Node Connectivity",
			Result:  true,
			Message: "",
		})
	}

	return testResults
}
