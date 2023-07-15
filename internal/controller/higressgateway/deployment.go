package higressgateway

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/alibaba/higress/api/v1alpha1"
)

const (
	instanceName = "higress-gateway"
)

func initDeployment(deploy *appsv1.Deployment, instance *v1alpha1.HigressGateway) *appsv1.Deployment {
	deploy = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name,
			Namespace:   instance.Namespace,
			Labels:      instance.Labels,
			Annotations: instance.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: instance.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: instance.Spec.SelectorLabels,
			},
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &instance.Spec.RollingMaxUnavailable,
					MaxSurge:       &instance.Spec.RollingMaxSurge,
				},
			},
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					ImagePullSecrets:   instance.Spec.ImagePullSecrets,
					ServiceAccountName: instance.Spec.ServiceAccount.Name,
					SecurityContext:    genSecurityContextForPod(instance),
					NodeSelector:       instance.Spec.NodeSelector,
					Affinity:           instance.Spec.Affinity,
					Tolerations:        instance.Spec.Toleration,
					Containers: []apiv1.Container{
						{
							Name:            instanceName,
							Image:           genImage(instance),
							Args:            genArgs(instance),
							SecurityContext: genSecurityContextForContainer(instance),
							Env:             genEnv(instance),
							Ports:           genPorts(instance),
							ReadinessProbe:  genProbe(instance),
							VolumeMounts:    genVolumeMounts(instance),
						},
					},
					Volumes: genVolumes(instance),
				},
			},
		},
	}

	// resource
	if !instance.Spec.Local && instance.Spec.Resources != nil {
		deploy.Spec.Template.Spec.Containers[0].Resources = *instance.Spec.Resources
	}

	// hostnetwork
	if instance.Spec.HostNetwork {
		deploy.Spec.Template.Spec.HostNetwork = instance.Spec.HostNetwork
		deploy.Spec.Template.Spec.DNSPolicy = apiv1.DNSClusterFirstWithHostNet
	}
	return deploy
}

func muteDeployment(deploy *appsv1.Deployment, instance *v1alpha1.HigressGateway) controllerutil.MutateFn {
	return func() error {
		initDeployment(deploy, instance)
		return nil
	}
}

func genProbe(instance *v1alpha1.HigressGateway) *apiv1.Probe {
	return &apiv1.Probe{
		FailureThreshold: 30,
		ProbeHandler: apiv1.ProbeHandler{
			HTTPGet: &apiv1.HTTPGetAction{
				Path:   "healthz/ready",
				Port:   intstr.FromInt(15021),
				Scheme: "HTTP",
			},
		},
		InitialDelaySeconds: 1,
		PeriodSeconds:       2,
		SuccessThreshold:    1,
		TimeoutSeconds:      3,
	}
}

func genPorts(instance *v1alpha1.HigressGateway) []apiv1.ContainerPort {
	if len(instance.Spec.Ports) > 0 {
		return instance.Spec.Ports
	}

	ports := []apiv1.ContainerPort{
		{
			Name:          "http-envoy-prom",
			Protocol:      "TCP",
			ContainerPort: 15090,
		},
	}

	if instance.Spec.Local {
		ports = append(ports, []apiv1.ContainerPort{
			{
				Name:          "http",
				Protocol:      "TCP",
				ContainerPort: 80,
				HostPort:      80,
			},
			{
				Name:          "https",
				Protocol:      "TCP",
				ContainerPort: 443,
				HostPort:      443,
			},
		}...)
	}

	return ports
}

func genEnv(instance *v1alpha1.HigressGateway) []apiv1.EnvVar {
	envs := []apiv1.EnvVar{
		{
			Name: "NODE_NAME",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "spec.nodeName",
				},
			},
		},
		{
			Name: "POD_NAME",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		{
			Name: "INSTANCE_IP",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
		{
			Name: "HOST_IP",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.hostIP",
				},
			},
		},
		{
			Name: "SERVICE_ACCOUNT",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "spec.serviceAccountName",
				},
			},
		},
		{
			Name:  "PROXY_XDS_VIA_AGENT",
			Value: "true",
		},
		{
			Name:  "ENABLE_INGRESS_GATEWAY_SDS",
			Value: "false",
		},
		{
			Name:  "JWT_POLICY",
			Value: instance.Spec.JwtPolicy,
		},
		{
			Name:  "ISTIO_META_HTTP10",
			Value: "1",
		},
		{
			Name:  "ISTIO_META_CLUSTER_ID",
			Value: "kubernetes",
		},
		{
			Name:  "INSTANCE_NAME",
			Value: instanceName,
		},
	}

	skywalking := instance.Spec.Skywalking
	if skywalking != nil && skywalking.Enable {
		envs = append(envs, apiv1.EnvVar{
			Name:  "ISTIO_BOOTSTRAP_OVERRIDE",
			Value: "/etc/istio/custom-bootstrap/custom_bootstrap.json",
		})
	}

	if instance.Spec.NetWorkGateway != "" {
		envs = append(envs, apiv1.EnvVar{
			Name:  "ISTIO_META_REQUESTED_NETWORK_VIEW",
			Value: instance.Spec.NetWorkGateway,
		})
	}

	for k, v := range instance.Spec.Env {
		envs = append(envs, apiv1.EnvVar{Name: k, Value: v})
	}

	return envs
}

func genImage(instance *v1alpha1.HigressGateway) string {
	return fmt.Sprintf("%v:%v", instance.Spec.Image.Repository, instance.Spec.Image.Tag)
}

func genArgs(instance *v1alpha1.HigressGateway) []string {
	return []string{
		"proxy",
		"router",
		"--domain",
		instance.Namespace + "svc.cluster.local",
		"--proxyLogLevel=warning",
		"--proxyComponetLogLevel=misc:error",
		"--log_output_level=all:info",
		"--serviceCluster=higress-gateway",
	}
}

// todo(lql): Version adaptation
func genSecurityContextForContainer(instance *v1alpha1.HigressGateway) *apiv1.SecurityContext {
	if instance.Spec.SecurityContext != nil {
		return instance.Spec.SecurityContext
	}

	readOnlyRootFileSystem := true
	runAsGroup := int64(1337)
	runAsUser := int64(1337)
	runAsNonRoot := true

	return &apiv1.SecurityContext{
		ReadOnlyRootFilesystem: &readOnlyRootFileSystem,
		RunAsNonRoot:           &runAsNonRoot,
		RunAsUser:              &runAsUser,
		RunAsGroup:             &runAsGroup,
		Capabilities: &apiv1.Capabilities{
			Drop: []apiv1.Capability{"ALL"},
		},
	}
}

// todo(lql): Version adaptation
func genSecurityContextForPod(instance *v1alpha1.HigressGateway) *apiv1.PodSecurityContext {
	if instance.Spec.PodSecurityContext != nil {
		return instance.Spec.PodSecurityContext
	}
	return &apiv1.PodSecurityContext{
		Sysctls: []apiv1.Sysctl{
			{
				Name:  "net.ipv4.ip_unprivileged_port_start",
				Value: "0",
			},
		},
	}
}

func genVolumeMounts(instance *v1alpha1.HigressGateway) []apiv1.VolumeMount {
	mounts := []apiv1.VolumeMount{
		{
			Name:      "config",
			MountPath: "/etc/istio/config",
		},
		{
			Name:      "istio-ca-root-cert",
			MountPath: "/var/run/secrets/istio",
		},
		{
			Name:      "istio-data",
			MountPath: "/var/lib/istio/data",
		},
		{
			Name:      "podinfo",
			MountPath: "/etc/istio/pod",
		},
		{
			Name:      "proxy-socket",
			MountPath: "/etc/istio/proxy",
		},
	}

	if instance.Spec.JwtPolicy == "third-party-jwt" {
		mounts = append(mounts, apiv1.VolumeMount{
			Name:      "istio-token",
			MountPath: "/var/run/secrets/tokens",
			ReadOnly:  true,
		})
	}

	skywalking := instance.Spec.Skywalking
	if skywalking != nil && skywalking.Enable {
		mounts = append(mounts, apiv1.VolumeMount{
			Name:      "custom-bootstrap-volume",
			MountPath: "/etc/istio/custom-bootstrap",
		})
	}

	return mounts
}

func genVolumes(instance *v1alpha1.HigressGateway) []apiv1.Volume {
	var volumes []apiv1.Volume

	volumes = append(volumes, apiv1.Volume{
		Name: "istio-ca-root-cert",
		VolumeSource: apiv1.VolumeSource{
			ConfigMap: &apiv1.ConfigMapVolumeSource{
				LocalObjectReference: apiv1.LocalObjectReference{
					Name: "higress-config",
				},
			},
		},
	})

	caRootCertName := "istio-ca-root-cert"
	if instance.Spec.EnableHigressIstio {
		caRootCertName = "higress-ca-root-cert"
	}
	mode := int32(420)
	quantiy := resource.MustParse("1m")
	expirationSeconds := int64(43200)
	hostPathType := apiv1.HostPathDirectory
	volumes = append(volumes, []apiv1.Volume{
		{
			Name: "config",
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: caRootCertName,
					},
				},
			},
		},
		{
			Name: "istio-data",
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "proxy-socket",
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "podinfo",
			VolumeSource: apiv1.VolumeSource{
				DownwardAPI: &apiv1.DownwardAPIVolumeSource{
					DefaultMode: &mode,
					Items: []apiv1.DownwardAPIVolumeFile{
						{
							Path: "labels",
							FieldRef: &apiv1.ObjectFieldSelector{
								APIVersion: "v1",
								FieldPath:  "metadata.labels",
							},
						},
						{
							Path: "annotations",
							FieldRef: &apiv1.ObjectFieldSelector{
								APIVersion: "v1",
								FieldPath:  "metadata.annotations",
							},
						},
						{
							Path: "cpu-request",
							ResourceFieldRef: &apiv1.ResourceFieldSelector{
								ContainerName: instanceName,
								Divisor:       quantiy,
								Resource:      "request.cpu",
							},
						},
						{
							Path: "cpu-limit",
							ResourceFieldRef: &apiv1.ResourceFieldSelector{
								ContainerName: instanceName,
								Divisor:       quantiy,
								Resource:      "limits.cpu",
							},
						},
					},
				},
			},
		},
	}...)

	if instance.Spec.JwtPolicy == "third-party-jwt" {
		volumes = append(volumes, apiv1.Volume{
			Name: "istio-token",
			VolumeSource: apiv1.VolumeSource{
				Projected: &apiv1.ProjectedVolumeSource{
					Sources: []apiv1.VolumeProjection{
						{
							ServiceAccountToken: &apiv1.ServiceAccountTokenProjection{
								Path:              "istio-token",
								Audience:          "istio-ca",
								ExpirationSeconds: &expirationSeconds,
							},
						},
					},
				},
			},
		})
	}

	if skywalking := instance.Spec.Skywalking; skywalking != nil && skywalking.Enable {
		volumes = append(volumes, apiv1.Volume{
			Name: "custom-bootstrap-volume",
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: "higress-custom-bootstrap",
					},
					DefaultMode: &mode,
				},
			},
		})
	}

	if len(instance.Spec.VolumeWasmPlugins) > 0 {
		volumes = append(volumes, apiv1.Volume{
			Name: "local-wasmplugins-volume",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/opt/plugins",
					Type: &hostPathType,
				},
			},
		})
	}

	return volumes
}
