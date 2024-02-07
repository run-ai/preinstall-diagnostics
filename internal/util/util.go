package util

import (
	"context"
	"time"

	"github.com/run-ai/preinstall-diagnostics/internal/env"
	"github.com/run-ai/preinstall-diagnostics/internal/log"
	"github.com/run-ai/preinstall-diagnostics/internal/resources"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

func RunInformationalTests(tests []func(*log.Logger) error, logger *log.Logger) []error {
	var errs []error
	for _, test := range tests {
		err := test(logger)
		if err != nil {
			errs = append(errs, err)
			logger.WarningF("%v", err)
			logger.Warning()
		} else {
			logger.Pass()
		}
	}

	return errs
}

func RunTests(tests []func(*log.Logger) error, logger *log.Logger) []error {
	var errs []error
	for _, test := range tests {
		err := test(logger)
		if err != nil {
			errs = append(errs, err)
			logger.ErrorF("%v", err)
			logger.Fail()
		} else {
			logger.Pass()
		}
	}

	return errs
}

func TemplateResources(backendFQDN, image, imagePullSecretName, imageRegistry, runaiSaas string) {
	ds := &resources.DaemonSet
	if backendFQDN != "" {
		ds.Spec.Template.Spec.Containers[0].Env =
			append(ds.Spec.Template.Spec.Containers[0].Env,
				v1.EnvVar{
					Name:  env.BackendFQDNEnvVar,
					Value: backendFQDN,
				})
	}

	if imagePullSecretName != "" {
		ds.Spec.Template.Spec.ImagePullSecrets = []v1.LocalObjectReference{
			{
				Name: imagePullSecretName,
			},
		}
	}

	if imageRegistry != "" {
		ds.Spec.Template.Spec.Containers[0].Env =
			append(ds.Spec.Template.Spec.Containers[0].Env,
				v1.EnvVar{
					Name:  env.RegistryEnvVar,
					Value: imageRegistry,
				})
	}

	if runaiSaas != "" {
		ds.Spec.Template.Spec.Containers[0].Env =
			append(ds.Spec.Template.Spec.Containers[0].Env,
				v1.EnvVar{
					Name:  env.RunAISaasEnvVar,
					Value: runaiSaas,
				})
	}

	if image != "" {
		ds.Spec.Template.Spec.Containers[0].Image = image
	}
}

func DeleteResources(client kubernetes.Interface, dynClient dynamic.Interface, logger *log.Logger) error {
	err := resources.DeleteResources(resources.DeletionOrder(), dynClient)
	if err != nil {
		return err
	}

	logger.WriteStringF("waiting for all resources to be deleted...")

	// wait for ns to be deleted
	for {
		_, err := client.CoreV1().Namespaces().Get(context.TODO(),
			resources.Namespace.Name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				break
			}

			return err
		}

		time.Sleep(time.Second)
	}

	logger.WriteStringF("all resources were successfully deleted")
	return nil
}

func CreateResources(backendFQDN string, dynClient dynamic.Interface) error {
	return resources.CreateResources(resources.CreationOrder(), dynClient)
}
