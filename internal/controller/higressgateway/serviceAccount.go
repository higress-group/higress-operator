package higressgateway

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/alibaba/higress/api/v1alpha1"
)

func initServiceAccount(sa *apiv1.ServiceAccount, instance *v1alpha1.HigressGateway) *apiv1.ServiceAccount {
	*sa = apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   instance.Namespace,
			Name:        getServiceAccount(instance),
			Labels:      instance.Labels,
			Annotations: instance.Spec.ServiceAccount.Annotations,
		},
	}

	return sa
}

func getServiceAccount(instance *v1alpha1.HigressGateway) string {
	if instance.Spec.ServiceAccount.Name != "" {
		return instance.Spec.ServiceAccount.Name
	}
	return instance.Name
}
