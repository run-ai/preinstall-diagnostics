package cli

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/run-ai/preinstall-diagnostics/internal/external-cluster-tests"
	"github.com/run-ai/preinstall-diagnostics/internal/utils"
)

const (
	defaultStorageClassAnnotation = "storageclass.kubernetes.io/is-default-class"
)

func RunTestsAndAppendToTable(t table.Writer, clusterFQDN string) {
	showClusterVersion(t)
	t.AppendSeparator()
	certificatesAreValid(t, clusterFQDN)
	t.AppendSeparator()
	helmRepoReachable(t)
	t.AppendSeparator()
	ingressControllerExists(t)
	t.AppendSeparator()
	prometheusInstalled(t)
	t.AppendSeparator()
	showGPUNodes(t)
	t.AppendSeparator()
	showStorageClasses(t)
	t.AppendSeparator()
	listPods(t)
}

func showClusterVersion(t table.Writer) {
	testName := "Kubernetes Cluster Version"
	clusterVersion, err := external_cluster_tests.ShowClusterVersion()
	if err != nil {
		utils.AppendRowToTable(t, testName, false, err.Error())
	} else {
		utils.AppendRowToTable(t, testName, true, clusterVersion)
	}
}

func certificatesAreValid(t table.Writer, clusterFQDN string) {
	testName := "TLS Certificates verification"
	err := external_cluster_tests.CertificateIsValid(clusterFQDN)
	if err != nil {
		utils.AppendRowToTable(t, testName, false, err.Error())
	} else {
		utils.AppendRowToTable(t, testName, true, "")
	}
}

func helmRepoReachable(t table.Writer) {
	testName := "Helm Repository Connectivity"
	reachable, err := external_cluster_tests.RunAIHelmRepositoryReachable()
	if err != nil {
		utils.AppendRowToTable(t, testName, false, err.Error())
	} else {
		utils.AppendRowToTable(t, testName, reachable, "")
	}
}

func ingressControllerExists(t table.Writer) {
	testName := "Ingress Controller Installed"
	exists, err := external_cluster_tests.NGINXIngressControllerInstalled()
	if err != nil {
		utils.AppendRowToTable(t, testName, false, err.Error())
	} else {
		utils.AppendRowToTable(t, testName, exists, "")
	}
}

func listPods(t table.Writer) {
	testName := "Pod List"
	pods, err := external_cluster_tests.ListPods()
	if err != nil {
		utils.AppendRowToTable(t, testName, false, err.Error())
	} else {
		podsStr := ""
		for i := range pods {
			podsStr += fmt.Sprintf("%s/%s/%s",
				pods[i].Namespace, pods[i].Name, pods[i].Status.Phase)
			if i < len(pods)-1 {
				podsStr += "\n"
			}
		}
		utils.AppendRowToTable(t, testName, true, podsStr)
	}
}

func prometheusInstalled(t table.Writer) {
	testName := "Prometheus Installed"
	installed, err := external_cluster_tests.PrometheusInstalled()
	if err != nil {
		utils.AppendRowToTable(t, testName, false, err.Error())
	} else {
		utils.AppendRowToTable(t, testName, installed, "")
	}
}

func showGPUNodes(t table.Writer) {
	testName := "GPU Nodes"
	nodes, err := external_cluster_tests.ShowGPUNodes()
	if err != nil {
		utils.AppendRowToTable(t, testName, false, err.Error())
	} else {
		nodesStr := ""
		for i := range nodes {
			nodesStr += nodes[i].Name

			if i < len(nodes)-1 {
				nodesStr += "\n"
			}
		}
		utils.AppendRowToTable(t, testName, true, nodesStr)
	}
}

func showStorageClasses(t table.Writer) {
	testName := "Available StorageClasses"
	scs, err := external_cluster_tests.ShowStorageClasses()
	if err != nil {
		utils.AppendRowToTable(t, testName, false, err.Error())
	} else {
		scsStr := ""
		for i := range scs {
			scsStr += scs[i].Name

			if possibleDefault, exists := scs[i].Annotations[defaultStorageClassAnnotation]; exists {
				if possibleDefault == "true" {
					scsStr += " (default)"
				}
			}

			if i < len(scs)-1 {
				scsStr += "\n"
			}
		}
		utils.AppendRowToTable(t, testName, true, scsStr)
	}
}
