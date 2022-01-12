package resources

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ClusterRole = rbacv1.ClusterRole{
	TypeMeta: metav1.TypeMeta{
		APIVersion: rbacGV,
		Kind:       "ClusterRole",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: defaultResourceName,
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
				"pods",
				"nodes",
			},
			Verbs: []string{
				"get",
				"list",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			Resources: []string{
				"daemonsets",
				"deployments",
			},
			Verbs: []string{
				"get",
				"list",
			},
		},
		{
			APIGroups: []string{
				"config.openshift.io",
			},
			Resources: []string{
				"clusterversions",
			},
			Verbs: []string{
				"get",
			},
		},
		{
			APIGroups: []string{
				"monitoring.coreos.com",
			},
			Resources: []string{
				"prometheuses",
			},
			Verbs: []string{
				"list",
			},
		},
		{
			APIGroups: []string{
				"storage.k8s.io",
			},
			Resources: []string{
				"storageclasses",
			},
			Verbs: []string{
				"list",
			},
		},
	},
}
