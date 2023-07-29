package higressgateway

import (
	"k8s.io/api/autoscaling/v2"
	"k8s.io/api/autoscaling/v2beta1"
	"k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/alibaba/higress/higress-operator/api/v1alpha1"
)

func initHPAv2(hpa *v2.HorizontalPodAutoscaler, instance *v1alpha1.HigressGateway) *v2.HorizontalPodAutoscaler {
	*hpa = v2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name,
			Namespace:   instance.Namespace,
			Labels:      instance.Labels,
			Annotations: instance.Annotations,
		},
		Spec: v2.HorizontalPodAutoscalerSpec{
			MaxReplicas: instance.Spec.AutoScaling.MaxReplicas,
			MinReplicas: instance.Spec.AutoScaling.MinReplicas,
			ScaleTargetRef: v2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       instance.Name,
				APIVersion: "apps/v1",
			},
			Metrics: []v2.MetricSpec{
				{
					Type: v2.ResourceMetricSourceType,
					Resource: &v2.ResourceMetricSource{
						Name: "cpu",
						Target: v2.MetricTarget{
							Type:               v2.UtilizationMetricType,
							AverageUtilization: instance.Spec.AutoScaling.TargetCPUUtilizationPercentage,
						},
					},
				},
			},
		},
	}

	return hpa
}

func muteHPA(hpa *v2.HorizontalPodAutoscaler, instance *v1alpha1.HigressGateway) controllerutil.MutateFn {
	return func() error {
		initHPAv2(hpa, instance)
		return nil
	}
}

func convertV2ToV2beta1(hpa *v2.HorizontalPodAutoscaler) (res *v2beta1.HorizontalPodAutoscaler) {
	res = &v2beta1.HorizontalPodAutoscaler{
		ObjectMeta: hpa.ObjectMeta,
		Spec: v2beta1.HorizontalPodAutoscalerSpec{
			MinReplicas:    hpa.Spec.MinReplicas,
			MaxReplicas:    hpa.Spec.MaxReplicas,
			ScaleTargetRef: v2beta1.CrossVersionObjectReference(hpa.Spec.ScaleTargetRef),
		},
	}

	for _, metric := range hpa.Spec.Metrics {
		res.Spec.Metrics = append(res.Spec.Metrics, v2beta1.MetricSpec{
			Type: v2beta1.MetricSourceType(metric.Type),
			Resource: &v2beta1.ResourceMetricSource{
				Name:                     metric.Resource.Name,
				TargetAverageUtilization: metric.Resource.Target.AverageUtilization,
				TargetAverageValue:       metric.Resource.Target.AverageValue,
			},
		})
	}

	return
}

func convertV2ToV2beta2(hpa *v2.HorizontalPodAutoscaler) (res *v2beta2.HorizontalPodAutoscaler) {
	res = &v2beta2.HorizontalPodAutoscaler{
		ObjectMeta: hpa.ObjectMeta,
		Spec: v2beta2.HorizontalPodAutoscalerSpec{
			MinReplicas:    hpa.Spec.MinReplicas,
			MaxReplicas:    hpa.Spec.MaxReplicas,
			ScaleTargetRef: v2beta2.CrossVersionObjectReference(hpa.Spec.ScaleTargetRef),
		},
	}

	for _, metric := range hpa.Spec.Metrics {
		res.Spec.Metrics = append(res.Spec.Metrics, v2beta2.MetricSpec{
			Type: v2beta2.MetricSourceType(metric.Type),
			Resource: &v2beta2.ResourceMetricSource{
				Name: metric.Resource.Name,
				Target: v2beta2.MetricTarget{
					Type:               v2beta2.MetricTargetType(metric.Resource.Target.Type),
					Value:              metric.Resource.Target.Value,
					AverageValue:       metric.Resource.Target.AverageValue,
					AverageUtilization: metric.Resource.Target.AverageUtilization,
				},
			},
		})
	}

	return
}

func convertV2beta2ToV2(hpa *v2beta2.HorizontalPodAutoscaler) (res *v2.HorizontalPodAutoscaler) {
	res = &v2.HorizontalPodAutoscaler{
		ObjectMeta: hpa.ObjectMeta,
		Spec: v2.HorizontalPodAutoscalerSpec{
			MinReplicas:    hpa.Spec.MinReplicas,
			MaxReplicas:    hpa.Spec.MaxReplicas,
			ScaleTargetRef: v2.CrossVersionObjectReference(hpa.Spec.ScaleTargetRef),
		},
	}

	for _, metric := range hpa.Spec.Metrics {
		res.Spec.Metrics = append(res.Spec.Metrics, v2.MetricSpec{
			Type: v2.MetricSourceType(metric.Type),
			Resource: &v2.ResourceMetricSource{
				Name: metric.Resource.Name,
				Target: v2.MetricTarget{
					Type:               v2.MetricTargetType(metric.Resource.Target.Type),
					Value:              metric.Resource.Target.Value,
					AverageValue:       metric.Resource.Target.AverageValue,
					AverageUtilization: metric.Resource.Target.AverageUtilization,
				},
			},
		})
	}

	return
}

func convertV2beta1ToV2(hpa *v2beta1.HorizontalPodAutoscaler) (res *v2.HorizontalPodAutoscaler) {
	res = &v2.HorizontalPodAutoscaler{
		ObjectMeta: hpa.ObjectMeta,
		Spec: v2.HorizontalPodAutoscalerSpec{
			MinReplicas:    hpa.Spec.MinReplicas,
			MaxReplicas:    hpa.Spec.MaxReplicas,
			ScaleTargetRef: v2.CrossVersionObjectReference(hpa.Spec.ScaleTargetRef),
		},
	}

	for _, metric := range hpa.Spec.Metrics {
		item := v2.MetricSpec{
			Type: v2.MetricSourceType(metric.Type),
			Resource: &v2.ResourceMetricSource{
				Name: metric.Resource.Name,
			},
		}

		if metric.Resource.TargetAverageUtilization != nil {
			item.Resource.Target.Type = v2.UtilizationMetricType
			item.Resource.Target.AverageUtilization = metric.Resource.TargetAverageUtilization
		} else if metric.Resource.TargetAverageValue != nil {
			item.Resource.Target.Type = v2.AverageValueMetricType
			item.Resource.Target.AverageValue = metric.Resource.TargetAverageValue
		}

		res.Spec.Metrics = append(res.Spec.Metrics, item)
	}

	return
}
