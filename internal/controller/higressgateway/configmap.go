package higressgateway

import (
	"fmt"

	"gopkg.in/yaml.v2"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1alpha1 "github.com/alibaba/higress/api/v1alpha1"
)

func initGatewayConfigMap(cm *apiv1.ConfigMap, instance *operatorv1alpha1.HigressGateway) (*apiv1.ConfigMap, error) {
	cm = &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "higress-config",
			Namespace: instance.Namespace,
			Labels:    instance.Labels,
		},
	}

	var (
		data              map[string]string
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
		if meshConfig.RootNamespace != "" {
			meshConfig.RootNamespace = instance.Spec.IstioNamespace
		}
	} else {
		meshConfig.RootNamespace = instance.Namespace
	}

	if meshConfigBytes, err = yaml.Marshal(meshConfig); err == nil {
		data["mesh"] = string(meshConfigBytes)
	}
	if err != nil {
		return nil, err
	}

	return cm, nil
}

func initSkywalkingConfigMap(cm *apiv1.ConfigMap, instance *operatorv1alpha1.HigressGateway) (*apiv1.ConfigMap, error) {
	cm = &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "todo(lql)",
			Namespace: instance.Namespace,
			Labels:    instance.Labels,
		},
	}

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
