package higressgateway

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1alpha1 "github.com/alibaba/higress/api/v1alpha1"
)

const (
	clusterRole = "higress-gateway"
)

func defaultRules() []rbacv1.PolicyRule {
	rules := []rbacv1.PolicyRule{
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{""},
			Resources: []string{"secrets"},
		},
	}

	return rules
}

func initClusterRole(cr *rbacv1.ClusterRole, _ *operatorv1alpha1.HigressGateway) *rbacv1.ClusterRole {
	*cr = rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRole,
		},
		Rules: defaultRules(),
	}
	return cr
}

func muteClusterRole(cr *rbacv1.ClusterRole, instance *operatorv1alpha1.HigressGateway) controllerutil.MutateFn {
	return func() error {
		cr.Name = getServiceAccount(instance)
		cr.Rules = defaultRules()
		return nil
	}
}

func initClusterRoleBinding(crb *rbacv1.ClusterRoleBinding, instance *operatorv1alpha1.HigressGateway) *rbacv1.ClusterRoleBinding {
	*crb = rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: getServiceAccount(instance),
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     clusterRole,
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      getServiceAccount(instance),
				Namespace: instance.Namespace,
			},
		},
	}

	return crb
}

func muteClusterRoleBinding(crb *rbacv1.ClusterRoleBinding, instance *operatorv1alpha1.HigressGateway) controllerutil.MutateFn {
	return func() error {
		initClusterRoleBinding(crb, instance)
		return nil
	}
}

func initRoleBinding(rb *rbacv1.RoleBinding, instance *operatorv1alpha1.HigressGateway) *rbacv1.RoleBinding {
	*rb = rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: getServiceAccount(instance),
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     getServiceAccount(instance),
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      getServiceAccount(instance),
				Namespace: instance.Namespace,
			},
		},
	}
	return rb
}

func muteRoleBinding(rb *rbacv1.RoleBinding, instance *operatorv1alpha1.HigressGateway) controllerutil.MutateFn {
	return func() error {
		initRoleBinding(rb, instance)
		return nil
	}
}

func initRole(role *rbacv1.Role, instance *operatorv1alpha1.HigressGateway) *rbacv1.Role {
	*role = rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getServiceAccount(instance),
			Namespace: instance.Namespace,
		},
		Rules: defaultRules(),
	}

	return role
}

func muteRole(role *rbacv1.Role, instance *operatorv1alpha1.HigressGateway) controllerutil.MutateFn {
	return func() error {
		initRole(role, instance)
		return nil
	}
}
