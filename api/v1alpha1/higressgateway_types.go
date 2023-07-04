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
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HigressGatewaySpec defines the desired state of HigressGateway
type HigressGatewaySpec struct {
	CRDCommonFields       `json:",inline"`
	ContainerCommonFields `json:",inline"`

	NetWorkGateway        string             `json:"netWorkGateway"`
	Skywalking            *Skywalking        `json:"skywalking"`
	AutoScaling           *AutoScaling       `json:"autoScaling"`
	RollingMaxSurge       intstr.IntOrString `json:"rollingMaxSurge"`
	RollingMaxUnavailable intstr.IntOrString `json:"rollingMaxUnavailable"`
}

// HigressGatewayStatus defines the observed state of HigressGateway
type HigressGatewayStatus struct {
	Deployed bool `json:"deployed"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HigressGateway is the Schema for the higressgateways API
type HigressGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HigressGatewaySpec   `json:"spec,omitempty"`
	Status HigressGatewayStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HigressGatewayList contains a list of HigressGateway
type HigressGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HigressGateway `json:"items"`
}

type Skywalking struct {
	Enable  bool   `json:"enable"`
	Port    *int32 `json:"port"`
	Address string `json:"address"`
}

func init() {
	SchemeBuilder.Register(&HigressGateway{}, &HigressGatewayList{})
}
