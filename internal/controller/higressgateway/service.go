package higressgateway

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/alibaba/higress/api/v1alpha1"
)

const (
	HigressGatewayServiceName = "higress-gateway"
)

func initService(svc *apiv1.Service, instance *v1alpha1.HigressGateway) *apiv1.Service {
	*svc = apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        HigressGatewayServiceName,
			Namespace:   instance.Namespace,
			Labels:      instance.Labels,
			Annotations: instance.Annotations,
		},
	}

	updateServiceSpec(svc, instance)
	return svc
}

func updateServiceSpec(svc *apiv1.Service, instance *v1alpha1.HigressGateway) {
	svc.Spec.Selector = instance.Spec.SelectorLabels

	if instance.Spec.NetWorkGateway != "" {
		svc.ObjectMeta.Labels["topology.istio.io/network"] = instance.Spec.NetWorkGateway
	}

	if instance.Spec.Service != nil {
		if ip := instance.Spec.Service.LoadBalancerIP; ip != "" {
			svc.Spec.LoadBalancerIP = ip
		}
		if ranges := instance.Spec.Service.LoadBalancerSourceRanges; len(ranges) > 0 {
			svc.Spec.LoadBalancerSourceRanges = ranges
		}
		if policy := instance.Spec.Service.ExternalTrafficPolicy; policy != "" {
			svc.Spec.ExternalTrafficPolicy = apiv1.ServiceExternalTrafficPolicy(policy)
		}

		svc.Spec.Type = apiv1.ServiceType(instance.Spec.Service.Type)
		svc.Spec.Ports = instance.Spec.Service.Ports
	}

	if instance.Spec.NetWorkGateway != "" {
		svc.Spec.Ports = []apiv1.ServicePort{
			{
				Name:       "status-port",
				Port:       15021,
				TargetPort: intstr.FromInt(15021),
			},
			{
				Name:       "tls",
				Port:       15443,
				TargetPort: intstr.FromInt(15443),
			},
			{
				Name:       "tls-istiod",
				Port:       15012,
				TargetPort: intstr.FromInt(15012),
			},
			{
				Name:       "tls-webhook",
				Port:       15017,
				TargetPort: intstr.FromInt(15017),
			},
		}
	}
}

func muteService(svc *apiv1.Service, instance *v1alpha1.HigressGateway) controllerutil.MutateFn {
	return func() error {
		updateServiceSpec(svc, instance)
		return nil
	}
}
