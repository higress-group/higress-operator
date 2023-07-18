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

	// +kubebuilder:validation:Optional
	NetWorkGateway string `json:"netWorkGateway"`
	// +kubebuilder:validation:Optional
	Skywalking *Skywalking `json:"skywalking"`
	// +kubebuilder:validation:Optional
	RollingMaxSurge intstr.IntOrString `json:"rollingMaxSurge"`
	// +kubebuilder:validation:Optional
	RollingMaxUnavailable intstr.IntOrString `json:"rollingMaxUnavailable"`
	// +kubebuilder:validation:Optional
	MeshConfig MeshConfig `json:"meshConfig"`
	// +kubebuilder:validation:Optional
	MeshNetworks map[string]Network `json:"meshNetworks"`
	// +kubebuilder:validation:Optional
	VolumeWasmPlugins []string `json:"volumeWasmPlugins"`
	// +kubebuilder:validation:Optional
	HostNetwork bool `json:"hostNetwork"`
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
	Enable bool `json:"enable"`
	// +kubebuilder:validation:Optional
	Port *int32 `json:"port"`
	// +kubebuilder:validation:Optional
	Address string `json:"address"`
	// +kubebuilder:validation:Optional
	CustomBootStrap string `json:"customBootStrap"`
}

type MeshConfig struct {
	TrustDomain              string         `json:"trustDomain"`
	AccessLogEncoding        string         `json:"accessLogEncoding"`
	AccessLogFile            string         `json:"accessLogFile"`
	IngressControllerMode    string         `json:"ingressControllerMode"`
	AccessLogFormat          string         `json:"accessLogFormat"`
	DnsRefreshRate           string         `json:"dnsRefreshRate"`
	EnableAutoMtls           bool           `json:"enableAutoMtls"`
	EnablePrometheusMerge    bool           `json:"enablePrometheusMerge"`
	ProtocolDetectionTimeout string         `json:"protocolDetectionTimeout"`
	ConfigSources            []ConfigSource `json:"configSources"`
	DefaultConfig            ProxyConfig    `json:"defaultConfig"`
	// +kubebuilder:validation:Optional
	RootNamespace string `json:"rootNamespace"`
}

type Network struct {
	Endpoints []Endpoint `json:"endpoints"`
	Gateways  []Gateway  `json:"gateways"`
}
type Endpoint struct {
	FromCidr     string `json:"fromCidr"`
	FromRegistry string `json:"fromRegistry"`
}
type Gateway struct {
	Address             string `json:"address"`
	RegistryServiceName string `json:"registryServiceName"`
	Port                int32  `json:"port"`
}
type ConfigSource struct {
	Address string `json:"address"`
}
type ProxyConfig struct {
	// +kubebuilder:validation:Optional
	DisableAlpnH2 bool `json:"disableAlpnH2"`
	// +kubebuilder:validation:Optional
	MeshId string `json:"meshId"`
	// +kubebuilder:validation:Optional
	Tracing *Tracing `json:"tracing"`
	// +kubebuilder:validation:Optional
	DiscoveryAddress string `json:"discoveryAddress"`
	// +kubebuilder:validation:Optional
	ProxyStatsMatcher *ProxyStatsMatcher `json:"proxyStatsMatcher"`
}
type ProxyStatsMatcher struct {
	// +kubebuilder:validation:Optional
	InclusionPrefixes []string `json:"inclusionPrefixes"`
	// +kubebuilder:validation:Optional
	InclusionSuffixes []string `json:"inclusionSuffixes"`
	// +kubebuilder:validation:Optional
	InclusionRegexps []string `json:"inclusionRegexps"`
}

type Tracing struct {
	// +kubebuilder:validation:Optional
	Zipkin *TracingZipkin `json:"zipkin"`
	// +kubebuilder:validation:Optional
	Lightstep *TracingLightstep `json:"lightstep"`
	// +kubebuilder:validation:Optional
	Datadog *TracingDatadog `json:"datadog"`
	// +kubebuilder:validation:Optional
	Stackdriver *TracingStackdriver `json:"stackdriver"`
	// +kubebuilder:validation:Optional
	OpenCensusAgent *TracingOpencensusagent `json:"openCensusAgent"`
}

type TracingZipkin struct {
	Address string `json:"address"`
}
type TracingLightstep struct {
	Address     string `json:"address"`
	AccessToken string `json:"accessToken"`
}
type TracingDatadog struct {
	Address string `json:"address"`
}
type TracingStackdriver struct {
	Debug bool `json:"debug"`
	// +kubebuilder:validation:Optional
	MaxNumberOfAttributes *int `json:"maxNumberOfAttributes"`
	// +kubebuilder:validation:Optional
	MaxNumberOfAnnotations *int `json:"maxNumberOfAnnotations"`
	// +kubebuilder:validation:Optional
	MaxNumberOfMessageEvents *int `json:"maxNumberOfMessageEvents"`
}
type TracingOpencensusagent struct {
	Address string `json:"address"`
}

func init() {
	SchemeBuilder.Register(&HigressGateway{}, &HigressGatewayList{})
}
