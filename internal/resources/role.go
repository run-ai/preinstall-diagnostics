package resources

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Role = rbacv1.ClusterRole{
	TypeMeta: metav1.TypeMeta{
		APIVersion: rbacGV,
		Kind:       "Role",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      defaultResourceName,
		Namespace: defaultResourceName,
		Labels: map[string]string{
			defaultResourceName: "",
		},
	},
	Rules: []rbacv1.PolicyRule{
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"configmaps",
			},
			Verbs: []string{
				"get",
				"list",
				"create",
				"delete",
			},
		},
	},
}
