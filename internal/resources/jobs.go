package resources

import (
	"github.com/run-ai/preinstall-diagnostics/internal/env"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RunInternalClusterTestsEnvVarName = "RUN_INTERNAL_CLUSTER_TESTS"
)

func TemplateJobsForNodes(nodeNames []string, backendFQDN string) []*batchv1.Job {
	jobs := []*batchv1.Job{}
	for _, nodeName := range nodeNames {
		jobs = append(jobs, templateJobForNode(nodeName, backendFQDN))
	}

	return jobs
}

func templateJobForNode(nodeName, backendFQDN string) *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: batchv1.SchemeGroupVersion.Group + "/" + batchv1.SchemeGroupVersion.Version,
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultResourceName + "-" + nodeName,
			Namespace: Namespace.Name,
			Labels: map[string]string{
				defaultResourceName: "",
			},
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						defaultResourceName: nodeName,
					},
				},
				Spec: v1.PodSpec{
					RestartPolicy: v1.RestartPolicyNever,
					Tolerations: []v1.Toleration{
						{
							Key: "node-role.kubernetes.io/master",
						},
					},
					NodeName:           nodeName,
					ServiceAccountName: ServiceAccount.Name,
					Containers: []v1.Container{
						{
							Name:            defaultResourceName,
							Image:           defaultImage,
							ImagePullPolicy: v1.PullAlways,
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  env.BackendFQDNEnvVar,
									Value: backendFQDN,
								},
								{
									Name:  RunInternalClusterTestsEnvVarName,
									Value: "true",
								},
								{
									Name: env.NodeNameEnvVar,
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
								{
									Name: env.PodNameEnvVar,
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: env.PodNamespaceEnvVar,
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
