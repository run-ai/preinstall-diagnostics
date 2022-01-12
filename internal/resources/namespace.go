package resources

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Namespace = v1.Namespace{
	TypeMeta: metav1.TypeMeta{
		APIVersion: coreAPIVersion,
		Kind:       "Namespace",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: defaultResourceName,
		Labels: map[string]string{
			defaultResourceName: "",
		},
	},
}
