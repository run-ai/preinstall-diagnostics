package resources

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ClusterRoleBinding = rbacv1.ClusterRoleBinding{
	TypeMeta: metav1.TypeMeta{
		APIVersion: rbacGV,
		Kind:       "ClusterRoleBinding",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: defaultResourceName,
		Labels: map[string]string{
			defaultResourceName: "",
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: rbacAPIGroup,
		Kind:     ClusterRole.Kind,
		Name:     ClusterRole.Name,
	},
	Subjects: []rbacv1.Subject{
		{
			Kind:      ServiceAccount.Kind,
			Name:      ServiceAccount.Name,
			Namespace: ServiceAccount.Namespace,
		},
	},
}
