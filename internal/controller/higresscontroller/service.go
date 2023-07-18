package higresscontroller

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1alpha1 "github.com/alibaba/higress/api/v1alpha1"
)

func initService(svc *apiv1.Service, instance *operatorv1alpha1.HigressController) *apiv1.Service {
	*svc = apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    instance.Labels,
		},
		Spec: apiv1.ServiceSpec{
			Selector: instance.Spec.SelectorLabels,
			Type:     apiv1.ServiceTypeClusterIP,
		},
	}

	if s := instance.Spec.Service; s != nil {
		if s.Type != "" {
			svc.Spec.Type = apiv1.ServiceType(s.Type)
		}
		svc.Spec.Ports = s.Ports
	}

	if !instance.Spec.EnableHigressIstio {
		ports := []apiv1.ServicePort{
			{
				Name:     "grpc-xds",
				Protocol: apiv1.ProtocolTCP,
				Port:     15010,
			},
			{
				Name:     "https-dns",
				Protocol: apiv1.ProtocolTCP,
				Port:     15012,
			},
			{
				Name:     "https-webhook",
				Protocol: apiv1.ProtocolTCP,
				Port:     443,
			},
			{
				Name:     "https-monitoring",
				Protocol: apiv1.ProtocolTCP,
				Port:     15014,
			},
		}
		svc.Spec.Ports = append(svc.Spec.Ports, ports...)
	}

	return svc
}

func muteService(svc *apiv1.Service, instance *operatorv1alpha1.HigressController) controllerutil.MutateFn {
	return func() error {
		initService(svc, instance)
		return nil
	}
}
