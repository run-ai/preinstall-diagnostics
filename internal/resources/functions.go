package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

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
				metav1.DeleteOptions{})

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

		if err != nil {
			return errors.Wrap(err,
				fmt.Sprintf("create error: resource kind: %s", gvk.Kind))
		}
	}

	return nil
}

func CreationOrder() []client.Object {
	return []client.Object{
		&Namespace,
		&ClusterRole,
		&ClusterRoleBinding,
		&ServiceAccount,
		&DaemonSet,
	}
}

func DeletionOrder() []client.Object {
	return []client.Object{
		&ClusterRole,
		&ClusterRoleBinding,
		&ServiceAccount,
		&DaemonSet,
		&Namespace,
	}
}
