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
	MeshConfig            MeshConfig         `json:"meshConfig"`
	MeshNetworks          map[string]Network `json:"meshNetworks"`
	VolumeWasmPlugins     []string           `json:"volumeWasmPlugins"`
	HostNetwork           bool               `json:"hostNetwork"`
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
	Enable          bool   `json:"enable"`
	Port            *int32 `json:"port"`
	Address         string `json:"address"`
	CustomBootStrap string `json:"customBootStrap"`
}

type MeshConfig struct {
	TrustDomain              string         `json:"trustDomain"`
	AccessLogEncoding        string         `json:"accessLogEncoding"`
	AccessLogFile            string         `json:"accessLogFile"`
	IngresssControllerMode   string         `json:"ingresssControllerMode"`
	AccessLogFormat          string         `json:"accessLogFormat"`
	DnsRefreshRate           *int64         `json:"dnsRefreshRate"`
	EnableAutoMtls           bool           `json:"enableAutoMtls"`
	EnablePrometheusMerge    bool           `json:"enablePrometheusMerge"`
	ProtocolDetectionTimeout *int64         `json:"protocolDetectionTimeout"`
	RootNamespace            string         `json:"rootNamespace"`
	ConfigSources            []ConfigSource `json:"configSources"`
	DefaultConfig            ProxyConfig    `json:"defaultConfig"`
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
	DisableAlpnH2     bool               `json:"disableAlpnH2"`
	MeshId            string             `json:"meshId"`
	Tracing           *Tracing           `json:"tracing"`
	DiscoveryAddress  string             `json:"discoveryAddress"`
	ProxyStatsMatcher *ProxyStatsMatcher `json:"proxyStatsMatcher"`
}
type ProxyStatsMatcher struct {
	InclusionPrefixes []string `json:"inclusionPrefixes"`
	InclusionSuffixes []string `json:"inclusionSuffixes"`
	InclusionRegexps  []string `json:"inclusionRegexps"`
}

type Tracing struct {
	Zipkin          *TracingZipkin          `json:"zipkin"`
	Lightstep       *TracingLightstep       `json:"lightstep"`
	Datadog         *TracingDatadog         `json:"datadog"`
	Stackdriver     *TracingStackdriver     `json:"stackdriver"`
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
	Debug                    bool `json:"debug"`
	MaxNumberOfAttributes    *int `json:"maxNumberOfAttributes"`
	MaxNumberOfAnnotations   *int `json:"maxNumberOfAnnotations"`
	MaxNumberOfMessageEvents *int `json:"maxNumberOfMessageEvents"`
}
type TracingOpencensusagent struct {
	Address string `json:"address"`
}

func init() {
	SchemeBuilder.Register(&HigressGateway{}, &HigressGatewayList{})
}
