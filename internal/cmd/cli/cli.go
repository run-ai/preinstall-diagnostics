package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	v2 "github.com/run-ai/preinstall-diagnostics/internal"
	"github.com/run-ai/preinstall-diagnostics/internal/k8sclient"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"github.com/run-ai/preinstall-diagnostics/internal/resources"
	"github.com/run-ai/preinstall-diagnostics/internal/utils"
	ver "github.com/run-ai/preinstall-diagnostics/internal/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func Main(clean, dryRun bool, backendFQDN, clusterFQDN, image, imagePullSecretName, imageRegistry,
	runaiSaas string, version bool, logger *log.Logger) {
	if version {
		fmt.Println(ver.Version)
		return
	}

	creationOrder, deletionOrder := resources.TemplateResources(backendFQDN, image,
		imagePullSecretName, imageRegistry, runaiSaas)

	if dryRun {
		err := resources.PrintResources(creationOrder)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		return
	}

	dynClient, err := k8sclient.DynamicClient()
	if err != nil {
		panic(err)
	}

	//k8s, err := k8sclient.ClientSet()
	//if err != nil {
	//	panic(err)
	//}

	_, _ = logger.WriteStringF("cleaning up previous deployment if it exists...")
	err = utils.DeleteResources(deletionOrder, dynClient, logger)
	if err != nil {
		panic(err)
	}

	if clean {
		return
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Test Name", "Result", "Test Message"})

	RunTestsAndAppendToTable(t, clusterFQDN)

	_, _ = logger.WriteStringF("deploying runai diagnostics tool...")
	err = utils.CreateResources(creationOrder, dynClient)
	if err != nil {
		panic(err)
	}

	// wait for job tests to complete and collect results
	err = utils.WaitForJobsToComplete(10*time.Second, 5*time.Minute)
	if err != nil {
		panic(err)
	}

	nodesResults, err := getNodesTestsResultsTables()
	if err != nil {
		panic(err)
	}

	for _, nodeResult := range nodesResults {
		t.AppendSeparator()
		utils.AppendRowToTable(t, "Node "+nodeResult.Name, nodeResult.CalculatedResult, nodeResult.TestResultsTable.Render())
	}

	_, _ = logger.WriteStringF("cleaning up...")
	err = utils.DeleteResources(deletionOrder, dynClient, logger)
	if err != nil {
		panic(err)
	}

	// compile results into a table
	logger.WriteStringF("%s", t.Render())
}

type NodeResult struct {
	Name             string
	TestResultsTable table.Writer
	CalculatedResult bool
}

func getNodesTestsResultsTables() ([]NodeResult, error) {
	nodesResults := []NodeResult{}

	k8s, err := k8sclient.ClientSet()
	if err != nil {
		panic(err)
	}

	nodeList, err := k8s.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, node := range nodeList.Items {
		cm, err := k8s.CoreV1().ConfigMaps("runai-diagnostics").
			Get(context.TODO(), "runai-diagnostics-"+node.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		nodeResultsJSON := cm.Data["results"]
		var nodeResult []v2.TestResult
		err = json.Unmarshal([]byte(nodeResultsJSON), &nodeResult)
		if err != nil {
			return nil, err
		}

		t := table.NewWriter()
		t.AppendHeader(table.Row{"Test Name", "Result", "Test Message"})

		pass := true

		for _, res := range nodeResult {
			utils.AppendRowToTable(t, res.Name, res.Result, res.Message)
			t.AppendSeparator()

			if !res.Result {
				pass = false
			}
		}

		nodesResults = append(nodesResults, NodeResult{
			Name:             node.Name,
			TestResultsTable: t,
			CalculatedResult: pass,
		})
	}

	return nodesResults, nil
}
