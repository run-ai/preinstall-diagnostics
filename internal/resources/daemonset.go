package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DamonsetModeEnvName = "DAEMONSET_MODE"
)

var DaemonSet = appsv1.DaemonSet{
	TypeMeta: metav1.TypeMeta{
		APIVersion: appsGV,
		Kind:       "DaemonSet",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      defaultResourceName,
		Namespace: Namespace.Name,
		Labels: map[string]string{
			defaultResourceName: "",
		},
	},
	Spec: appsv1.DaemonSetSpec{
		MinReadySeconds: 10,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				defaultResourceName: "",
			},
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					defaultResourceName: "",
				},
			},
			Spec: v1.PodSpec{
				Tolerations: []v1.Toleration{
					{
						Key: "node-role.kubernetes.io/master",
					},
				},
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
								Name:  DamonsetModeEnvName,
								Value: "true",
							},
							{
								Name: "NODE_NAME",
								ValueFrom: &v1.EnvVarSource{
									FieldRef: &v1.ObjectFieldSelector{
										FieldPath: "spec.nodeName",
									},
								},
							},
							{
								Name: "POD_NAME",
								ValueFrom: &v1.EnvVarSource{
									FieldRef: &v1.ObjectFieldSelector{
										FieldPath: "metadata.name",
									},
								},
							},
							{
								Name: "POD_NAMESPACE",
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
