package higresscontroller

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1alpha1 "github.com/alibaba/higress/higress-operator/api/v1alpha1"
)

const (
	HigressControllerServiceName = "higress-controller"
)

func initService(svc *apiv1.Service, instance *operatorv1alpha1.HigressController) *apiv1.Service {
	*svc = apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      HigressControllerServiceName,
			Namespace: instance.Namespace,
			Labels:    instance.Labels,
		},
	}

	updateServiceSpec(svc, instance)
	return svc
}

func updateServiceSpec(svc *apiv1.Service, instance *operatorv1alpha1.HigressController) {
	svc.Spec.Selector = instance.Spec.SelectorLabels
	svc.Spec.Type = apiv1.ServiceTypeClusterIP
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
		set := make(map[string]struct{})
		for _, port := range svc.Spec.Ports {
			set[port.Name] = struct{}{}
		}
		for _, port := range ports {
			if _, ok := set[port.Name]; !ok {
				svc.Spec.Ports = append(svc.Spec.Ports, port)
			}
		}
	}
}

func muteService(svc *apiv1.Service, instance *operatorv1alpha1.HigressController) controllerutil.MutateFn {
	return func() error {
		updateServiceSpec(svc, instance)
		return nil
	}
}
