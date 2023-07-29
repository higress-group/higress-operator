package higresscontroller

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1alpha1 "github.com/alibaba/higress/higress-operator/api/v1alpha1"
)

const (
	clusterRole = "higress-controller"
)

func defaultRules() []rbacv1.PolicyRule {
	rules := []rbacv1.PolicyRule{
		// ingress controller
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"networking.k8s.io", "extensions"},
			Resources: []string{"ingresses", "ingressclasses"},
		},
		{
			Verbs:     []string{"*"},
			APIGroups: []string{"networking.k8s.io", "extensions"},
			Resources: []string{"ingresses/status"},
		},
		// Needed for multi-cluster secret reading, possibly ingress certs in the future
		{
			Verbs:     []string{"get", "list", "watch", "create", "update"},
			APIGroups: []string{""},
			Resources: []string{"secrets"},
		},
		// required for CA's namespace controller
		{
			Verbs:     []string{"get", "list", "watch", "update", "create"},
			APIGroups: []string{""},
			Resources: []string{"configmaps"},
		},
		// Use for Kubernetes Service APIs
		{
			APIGroups: []string{"networking.x-k8s.io", "gateway.networking.k8s.io"},
			Resources: []string{"*"},
			Verbs:     []string{"get", "watch", "list", "update"},
		},
		//  discovery and routing
		{
			Verbs:     []string{"list", "watch", "get"},
			APIGroups: []string{""},
			Resources: []string{"pods", "nodes", "namespaces", "services", "endpoints"},
		},
		{
			APIGroups: []string{"discovery.k8s.io"},
			Resources: []string{"endpointslices"},
			Verbs:     []string{"get", "watch", "list"},
		},
		{
			Verbs:     []string{"create", "patch"},
			APIGroups: []string{""},
			Resources: []string{"events"},
		},
		// Istiod and bootstrap
		{
			Verbs:     []string{"update", "create", "get", "delete", "watch"},
			APIGroups: []string{"certificates.k8s.io"},
			Resources: []string{"certificatesigningrequests", "certificatesigningrequests/approval", "certificatesigningrequests/status"},
		},
		{
			Verbs:         []string{"approve"},
			APIGroups:     []string{"certificates.k8s.io"},
			Resources:     []string{"signers"},
			ResourceNames: []string{"kubernetes.io/legacy-unknown"},
		},
		// Used by Istiod to verify the JWT tokens
		{
			Verbs:     []string{"create"},
			APIGroups: []string{"authentication.k8s.io"},
			Resources: []string{"tokenreviews"},
		},
		// Used by Istiod to verify gateway SDS
		{
			Verbs:     []string{"create"},
			APIGroups: []string{"authorization.k8s.io"},
			Resources: []string{"subjectaccessreviews"},
		},
		// Used for MCS serviceexport management
		{
			APIGroups: []string{"multicluster.x-k8s.io"},
			Resources: []string{"serviceexports"},
			Verbs:     []string{"get", "watch", "list", "create", "delete"},
		},
		// Used for MCS serviceimport management
		{
			APIGroups: []string{"multicluster.x-k8s.io"},
			Resources: []string{"serviceimports"},
			Verbs:     []string{"get", "watch", "list"},
		},
		// sidecar injection controller
		{
			APIGroups: []string{"admissionregistration.k8s.io"},
			Resources: []string{"mutatingwebhookconfigurations"},
			Verbs:     []string{"get", "list", "watch", "update", "patch"},
		},
		// configuration validation webhook controller
		{
			APIGroups: []string{"admissionregistration.k8s.io"},
			Resources: []string{"validatingwebhookconfigurations"},
			Verbs:     []string{"get", "list", "watch", "update"},
		},
		// istio configuration
		// removing CRD permissions can break older versions of Istio running alongside this control plane (https://github.com/istio/istio/issues/29382)
		// please proceed with caution
		{
			APIGroups: []string{"config.istio.io", "security.istio.io", "networking.istio.io", "authentication.istio.io", "rbac.istio.io", "telemetry.istio.io", "extensions.istio.io"},
			Resources: []string{"*"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			Verbs:     []string{"get", "create", "watch", "list", "update", "patch"},
			APIGroups: []string{"extensions.higress.io"},
			Resources: []string{"wasmplugins"},
		},
		{
			Verbs:     []string{"get", "create", "watch", "list", "update", "patch"},
			APIGroups: []string{"networking.higress.io"},
			Resources: []string{"http2rpcs", "mcpbridges"},
		},
		// auto-detect installed CRD definitions
		{
			Verbs:     []string{"get", "watch", "list"},
			APIGroups: []string{"apiextensions.k8s.io"},
			Resources: []string{"customresourcedefinitions"},
		},
	}

	return rules
}

func initClusterRole(cr *rbacv1.ClusterRole, instance *operatorv1alpha1.HigressController) *rbacv1.ClusterRole {
	if cr == nil {
		return nil
	}

	*cr = rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRole,
		},
		Rules: defaultRules(),
	}
	return cr
}

func muteClusterRole(cr *rbacv1.ClusterRole, instance *operatorv1alpha1.HigressController) controllerutil.MutateFn {
	return func() error {
		cr.Name = clusterRole
		cr.Rules = defaultRules()
		return nil
	}
}

func initClusterRoleBinding(crb *rbacv1.ClusterRoleBinding, instance *operatorv1alpha1.HigressController) *rbacv1.ClusterRoleBinding {
	if crb == nil {
		return nil
	}

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

func muteClusterRoleBinding(crb *rbacv1.ClusterRoleBinding, instance *operatorv1alpha1.HigressController) controllerutil.MutateFn {
	return func() error {
		crb = initClusterRoleBinding(crb, instance)
		return nil
	}
}

func initRoleBinding(rb *rbacv1.RoleBinding, instance *operatorv1alpha1.HigressController) *rbacv1.RoleBinding {
	*rb = rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getServiceAccount(instance),
			Namespace: instance.Namespace,
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

func muteRoleBinding(rb *rbacv1.RoleBinding, instance *operatorv1alpha1.HigressController) controllerutil.MutateFn {
	return func() error {
		initRoleBinding(rb, instance)
		return nil
	}
}

func initRole(role *rbacv1.Role, instance *operatorv1alpha1.HigressController) *rbacv1.Role {
	*role = rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getServiceAccount(instance),
			Namespace: instance.Namespace,
		},
		Rules: defaultRules(),
	}

	return role
}

func muteRole(role *rbacv1.Role, instance *operatorv1alpha1.HigressController) controllerutil.MutateFn {
	return func() error {
		initRole(role, instance)
		return nil
	}
}
