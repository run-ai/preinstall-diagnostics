package resources

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ServiceAccount = v1.ServiceAccount{
	TypeMeta: metav1.TypeMeta{
		APIVersion: coreAPIVersion,
		Kind:       "ServiceAccount",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      defaultResourceName,
		Namespace: Namespace.Name,
		Labels: map[string]string{
			defaultResourceName: "",
		},
	},
}
