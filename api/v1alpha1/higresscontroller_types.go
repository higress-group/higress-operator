/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HigressControllerSpec defines the desired state of HigressController
type HigressControllerSpec struct {
	CRDCommonFields `json:",inline"`

	Controller ControllerSpec `json:"controller"`
	Pilot      PilotSpec      `json:"pilot"`
}

// HigressControllerStatus defines the observed state of HigressController
type HigressControllerStatus struct {
	Deployed bool `json:"deployed"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HigressController is the Schema for the higresscontrollers API
type HigressController struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HigressControllerSpec   `json:"spec,omitempty"`
	Status HigressControllerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HigressControllerList contains a list of HigressController
type HigressControllerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HigressController `json:"items"`
}

type ControllerSpec struct {
	ContainerCommonFields `json:",inline"`

	GatewayName    string `json:"gatewayName"`
	IngressClass   string `json:"ingressClass"`
	WatchNamespace string `json:"watchNamespace"`
}

type PilotSpec struct {
	ContainerCommonFields `json:",inline"`

	TraceSampling                     string   `json:"traceSampling"`
	JwksResolveExtraRootCA            string   `json:"jwksResolveExtraRootCA"`
	Plugins                           []string `json:"plugins"`
	KeepaliveMaxServerConnectionAge   int      `json:"keepaliveMaxServerConnectionAge"`
	ClusterDomain                     string   `json:"clusterDomain"`
	OneNamespace                      bool     `json:"oneNamespace"`
	JwtPolicy                         string   `json:"jwtPolicy"`
	EnableProtocolSniffingForOutbound bool     `json:"enableProtocolSniffingForOutbound"`
	EnableProtocolSniffingForInbound  bool     `json:"enableProtocolSniffingForInbound"`
}

func init() {
	SchemeBuilder.Register(&HigressController{}, &HigressControllerList{})
}
