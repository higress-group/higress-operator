package higressgateway

import (
	"fmt"

	"gopkg.in/yaml.v2"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1alpha1 "github.com/alibaba/higress/higress-operator/api/v1alpha1"
	"github.com/alibaba/higress/higress-operator/internal/controller"
	"github.com/alibaba/higress/higress-operator/internal/controller/higresscontroller"
)

func initGatewayConfigMap(cm *apiv1.ConfigMap, instance *operatorv1alpha1.HigressGateway) (*apiv1.ConfigMap, error) {
	*cm = apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      controller.HigressGatewayConfig,
			Namespace: instance.Namespace,
			Labels:    instance.Labels,
		},
	}

	if _, err := updateGatewayConfigMapSpec(cm, instance); err != nil {
		return nil, err
	}

	return cm, nil
}

func updateGatewayConfigMapSpec(cm *apiv1.ConfigMap, instance *operatorv1alpha1.HigressGateway) (*apiv1.ConfigMap, error) {
	var (
		data              = map[string]string{}
		err               error
		meshNetworksBytes []byte
		networksBytes     []byte
		meshConfigBytes   []byte
	)

	if networksBytes, err = yaml.Marshal(instance.Spec.MeshNetworks); err == nil {
		networksMap := make(map[string]string)
		networksMap["networks"] = string(networksBytes)
		if meshNetworksBytes, err = yaml.Marshal(networksMap); err == nil {
			data["meshNetworks"] = string(meshNetworksBytes)
		}
	}

	// rootNamespace
	meshConfig := instance.Spec.MeshConfig
	if instance.Spec.EnableHigressIstio {
		if meshConfig.RootNamespace == "" {
			meshConfig.RootNamespace = instance.Spec.IstioNamespace
		}
	} else {
		meshConfig.RootNamespace = instance.Namespace
	}

	// configSources
	if instance.Spec.EnableIstioAPI {
		meshConfig.ConfigSources = append(meshConfig.ConfigSources, operatorv1alpha1.ConfigSource{
			Address: "k8s://",
		})
	}

	// defaultConfig.tracing
	// defaultConfig.discoveryAddress
	if instance.Spec.EnableHigressIstio {
		instance.Spec.MeshConfig.DefaultConfig.DiscoveryAddress =
			fmt.Sprintf("%s.%s.svc:15012", "istiod", instance.Namespace)
	} else {
		instance.Spec.MeshConfig.DefaultConfig.DiscoveryAddress =
			fmt.Sprintf("%s.%s.svc:15012", higresscontroller.HigressControllerServiceName, instance.Namespace)
	}
	if meshConfigBytes, err = yaml.Marshal(meshConfig); err == nil {
		data["mesh"] = string(meshConfigBytes)
	}

	if err != nil {
		return nil, err
	}

	cm.Data = data
	return cm, nil
}

func initSkywalkingConfigMap(cm *apiv1.ConfigMap, instance *operatorv1alpha1.HigressGateway) (*apiv1.ConfigMap, error) {
	*cm = apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "skywalking-config",
			Namespace: instance.Namespace,
			Labels:    instance.Labels,
		},
	}

	if _, err := updateSkywalkingConfigMap(cm, instance); err != nil {
		return nil, err
	}

	return cm, nil
}

func updateSkywalkingConfigMap(cm *apiv1.ConfigMap, instance *operatorv1alpha1.HigressGateway) (*apiv1.ConfigMap, error) {
	skywalking := instance.Spec.Skywalking
	if skywalking != nil && skywalking.Enable && len(skywalking.CustomBootStrap) == 0 {
		return nil, fmt.Errorf("error empty skywalking custom bootstrap scripts")
	}

	cm.Data["custom_bootstrap.json"] = skywalking.CustomBootStrap

	return cm, nil
}

func muteConfigMap(cm *apiv1.ConfigMap, instance *operatorv1alpha1.HigressGateway,
	fn func(*apiv1.ConfigMap, *operatorv1alpha1.HigressGateway) (*apiv1.ConfigMap, error)) controllerutil.MutateFn {
	return func() error {
		if _, err := fn(cm, instance); err != nil {
			return err
		}
		return nil
	}
}
