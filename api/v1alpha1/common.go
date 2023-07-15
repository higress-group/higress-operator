package v1alpha1

import apiv1 "k8s.io/api/core/v1"

// +k8s:deepcopy-gen=true

type CRDCommonFields struct {
	Replicas           *int32                    `json:"replicas,omitempty"`
	SelectorLabels     map[string]string         `json:"selectorLabels"`
	NodeSelector       map[string]string         `json:"nodeSelector"`
	Affinity           *apiv1.Affinity           `json:"affinity"`
	Toleration         []apiv1.Toleration        `json:"toleration"`
	Service            *Service                  `json:"service"`
	RBAC               RBAC                      `json:"rbac"`
	ServiceAccount     ServiceAccount            `json:"serviceAccount"`
	AutoScaling        *AutoScaling              `json:"autoScaling"`
	PodSecurityContext *apiv1.PodSecurityContext `json:"podSecurityContext"`

	EnableStatus       bool         `json:"enableStatus"`
	EnableHigressIstio bool         `json:"enableHigressIstio"`
	EnableIstioAPI     bool         `json:"enableIstioAPI"`
	IstioNamespace     string       `json:"istioNamespace"`
	Revision           string       `json:"revision"`
	Istiod             Istio        `json:"istiod"`
	MultiCluster       MultiCluster `json:"multiCluster"`
	Local              bool         `json:"local"`
	JwtPolicy          string       `json:"jwtPolicy"`
}

// +k8s:deepcopy-gen=true

type ContainerCommonFields struct {
	Name             string                       `json:"name"`
	Annotations      map[string]string            `json:"annotations"`
	Image            Image                        `json:"image"`
	ImagePullSecrets []apiv1.LocalObjectReference `json:"imagePullSecrets"`
	Env              map[string]string            `json:"env"`
	ReadinessProbe   *apiv1.Probe                 `json:"readinessProbe"`
	Ports            []apiv1.ContainerPort        `json:"ports"`
	Resources        *apiv1.ResourceRequirements  `json:"resources"`
	SecurityContext  *apiv1.SecurityContext       `json:"securityContext"`
	LogLevel         string                       `json:"logLevel"`
	LogAsJson        bool                         `json:"logAsJson"`
}

// +k8s:deepcopy-gen=true

type Image struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	// +kubebuilder:validation:Enum="";Always;Never;IfNotPresent
	ImagePullPolicy apiv1.PullPolicy `json:"imagePullPolicy"`
}

// +k8s:deepcopy-gen=true

type ServiceAccount struct {
	Enable      bool              `json:"enable"`
	Name        string            `json:"name"`
	Annotations map[string]string `json:"annotations"`
}

// +k8s:deepcopy-gen=true

type AutoScaling struct {
	Enable                         bool   `json:"enable"`
	MinReplicas                    *int32 `json:"minReplicas"`
	MaxReplicas                    int32  `json:"maxReplicas"`
	TargetCPUUtilizationPercentage *int32 `json:"targetCPUUtilizationPercentage"`
}

// +k8s:deepcopy-gen=true

type RBAC struct {
	Enable bool `json:"enable"`
}

// +k8s:deepcopy-gen=true

type Istio struct {
	EnableAnalysis bool `json:"enableAnalysis"`
}

// +k8s:deepcopy-gen=true

type MultiCluster struct {
	Enable      bool   `json:"enable"`
	ClusterName string `json:"clusterName"`
}

// +k8s:deepcopy-gen=true

type Service struct {
	Type                     string              `json:"type"`
	Ports                    []apiv1.ServicePort `json:"ports"`
	Annotations              map[string]string   `json:"annotations"`
	LoadBalancerIP           string              `json:"loadBalancerIP"`
	LoadBalancerSourceRanges []string            `json:"loadBalancerSourceRanges"`
	ExternalTrafficPolicy    string              `json:"externalTrafficPolicy"`
}
