package resources

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/run-ai/preinstall-diagnostics/internal/env"
	"github.com/run-ai/preinstall-diagnostics/internal/k8sclient"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"strings"

	pluralize "github.com/gertd/go-pluralize"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func PrintResources(objs []client.Object) error {
	for _, res := range objs {
		resYAML, err := yaml.Marshal(res)
		if err != nil {
			return err
		}

		fmt.Println(string(resYAML))
		fmt.Println("---")
	}

	return nil
}

func DeleteResources(objs []client.Object, kubeDynamicClient dynamic.Interface) error {
	for _, res := range objs {
		gvk := res.GetObjectKind().GroupVersionKind()
		gvs := schema.GroupVersionResource{
			Group:    strings.ToLower(gvk.Group),
			Version:  strings.ToLower(gvk.Version),
			Resource: pluralize.NewClient().Pluralize(strings.ToLower(gvk.Kind), 2, false),
		}

		err := kubeDynamicClient.Resource(gvs).
			Namespace(res.GetNamespace()).
			Delete(context.TODO(),
				res.GetName(),
				metav1.DeleteOptions{
					PropagationPolicy: ptr.To(metav1.DeletePropagationBackground),
				})

		if err != nil {
			if !kerrors.IsNotFound(err) {
				return errors.Wrap(err,
					fmt.Sprintf("delete error: resource kind: %s", gvk.Kind))
			}
		}
	}

	return nil
}

func CreateResources(objs []client.Object, kubeDynamicClient dynamic.Interface) error {
	for _, res := range objs {
		unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(res)
		if err != nil {
			return err
		}

		gvk := res.GetObjectKind().GroupVersionKind()
		gvs := schema.GroupVersionResource{
			Group:    strings.ToLower(gvk.Group),
			Version:  strings.ToLower(gvk.Version),
			Resource: pluralize.NewClient().Pluralize(strings.ToLower(gvk.Kind), 2, false),
		}

		_, err = kubeDynamicClient.Resource(gvs).
			Namespace(res.GetNamespace()).
			Create(context.TODO(),
				&unstructured.Unstructured{Object: unstructuredObj},
				metav1.CreateOptions{})

		if err != nil && !kerrors.IsAlreadyExists(err) {
			return errors.Wrap(err,
				fmt.Sprintf("create error: resource kind: %s", gvk.Kind))
		}
	}

	return nil
}

func TemplateResources(backendFQDN, image, imagePullSecretName,
	imageRegistry, runaiSaas string) (creationOrder, deletionOrder []client.Object) {
	creationOrder = []client.Object{}

	k8s, err := k8sclient.ClientSet()
	if err != nil {
		panic(err)
	}

	nodeList, err := k8s.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	nodeNames := []string{}
	for _, node := range nodeList.Items {
		nodeNames = append(nodeNames, node.Name)
	}

	jobs := TemplateJobsForNodes(nodeNames, backendFQDN)

	for _, job := range jobs {
		if imagePullSecretName != "" {
			job.Spec.Template.Spec.ImagePullSecrets = []v1.LocalObjectReference{
				{
					Name: imagePullSecretName,
				},
			}
		}

		if imageRegistry != "" {
			job.Spec.Template.Spec.Containers[0].Env =
				append(job.Spec.Template.Spec.Containers[0].Env,
					v1.EnvVar{
						Name:  env.RegistryEnvVar,
						Value: imageRegistry,
					})
		}

		if runaiSaas != "" {
			job.Spec.Template.Spec.Containers[0].Env =
				append(job.Spec.Template.Spec.Containers[0].Env,
					v1.EnvVar{
						Name:  env.RunAISaasEnvVar,
						Value: runaiSaas,
					})
		}

		if image != "" {
			job.Spec.Template.Spec.Containers[0].Image = image
		}
	}

	creationOrder = append(creationOrder, &Namespace,
		&ClusterRole, &ClusterRoleBinding,
		&Role, &RoleBinding,
		&ServiceAccount)

	for _, job := range jobs {
		creationOrder = append(creationOrder, job)
	}

	deletionOrder = []client.Object{}

	for i := len(creationOrder) - 1; i >= 0; i-- {
		switch creationOrder[i].(type) {
		case *v1.Namespace:
			continue
		}

		deletionOrder = append(deletionOrder, creationOrder[i])
	}

	return creationOrder, deletionOrder
}
