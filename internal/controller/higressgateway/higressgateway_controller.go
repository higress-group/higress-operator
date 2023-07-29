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

package higressgateway

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/alibaba/higress/higress-operator/api/v1alpha1"
	. "github.com/alibaba/higress/higress-operator/internal/controller"
)

const (
	finalizer = "higressgateway.higress.io/finalizer"
)

// HigressGatewayReconciler reconciles a HigressGateway object
type HigressGatewayReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=operator.higress.io,resources=higressgateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.higress.io,resources=higressgateways/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.higress.io,resources=higressgateways/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=pods;services;services/finalizers;endpoints;persistentvolumeclaims;events;configmaps;secrets;serviceaccounts;namespaces,verbs=create;update;get;list;watch;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HigressGateway object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *HigressGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	instance := &operatorv1alpha1.HigressGateway{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("HigressGateway(%v) resource not found", req.NamespacedName))
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Failed to get resource HigressGateway(%v)", req.NamespacedName)
		return ctrl.Result{}, err
	}

	r.setDefaultValues(instance)

	// if deletionTimeStamp is not nil, it means is marked to be deleted
	if instance.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(instance, finalizer) {
			if err := r.finalizeHigressGateway(instance, logger); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(instance, finalizer)

			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	// check if the namespace still exists during the reconciling
	ns, nn := &apiv1.Namespace{}, types.NamespacedName{Name: instance.Namespace, Namespace: apiv1.NamespaceAll}
	err := r.Get(ctx, nn, ns)
	if (err != nil && errors.IsNotFound(err)) || (ns.Status.Phase == apiv1.NamespaceTerminating) {
		logger.Info(fmt.Sprintf("The namespace (%s) doesn't exist or is in Terminating status, canceling Reconciling", instance.Namespace))
		return ctrl.Result{}, nil
	} else if err != nil {
		logger.Error(err, "Failed to check if namespace exists")
		return ctrl.Result{}, nil
	}

	// add finalizer for this CR
	if controllerutil.AddFinalizer(instance, finalizer) {
		if err = r.Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err = r.createServiceAccount(ctx, instance, logger); err != nil {
		return ctrl.Result{}, err
	}

	if err = r.createRBAC(ctx, instance, logger); err != nil {
		return ctrl.Result{}, err
	}

	if err = r.createConfigMap(ctx, instance, logger); err != nil {
		return ctrl.Result{}, err
	}

	if err = r.createDeployment(ctx, instance, logger); err != nil {
		return ctrl.Result{}, err
	}

	if err = r.createService(ctx, instance, logger); err != nil {
		return ctrl.Result{}, err
	}

	if !instance.Status.Deployed {
		instance.Status.Deployed = true
		if err = r.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HigressGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.HigressGateway{}).
		Owns(&appsv1.Deployment{}).
		Owns(&apiv1.Service{}).
		Owns(&apiv1.ConfigMap{}).
		Owns(&apiv1.ServiceAccount{}).
		Complete(r)
}

func (r *HigressGatewayReconciler) createServiceAccount(ctx context.Context, instance *operatorv1alpha1.HigressGateway, logger logr.Logger) error {
	sa := initServiceAccount(&apiv1.ServiceAccount{}, instance)
	if err := ctrl.SetControllerReference(instance, sa, r.Scheme); err != nil {
		return err
	}

	exists, err := CreateIfNotExits(ctx, r.Client, sa)
	if err != nil {
		return err
	}

	if !exists {
		logger.Info(fmt.Sprintf("create serviceAccount for HigressGateway(%v)", instance.Name))
	}

	return nil
}

func (r *HigressGatewayReconciler) createRBAC(ctx context.Context, instance *operatorv1alpha1.HigressGateway, logger logr.Logger) error {
	if !instance.Spec.RBAC.Enable || !instance.Spec.ServiceAccount.Enable {
		return nil
	}

	var (
		role = &rbacv1.Role{}
		rb   = &rbacv1.RoleBinding{}
		cr   = &rbacv1.ClusterRole{}
		crb  = &rbacv1.ClusterRoleBinding{}
		err  error
	)
	// reconcile clusterRole
	cr = initClusterRole(cr, instance)
	if err = CreateOrUpdate(ctx, r.Client, "clusterrole", cr, muteClusterRole(cr, instance), logger); err != nil {
		return err
	}

	// reconcile clusterRoleBinding
	initClusterRoleBinding(crb, instance)
	if err = CreateOrUpdate(ctx, r.Client, "clusterRoleBinding", crb,
		muteClusterRoleBinding(crb, instance), logger); err != nil {
		return err
	}

	initRole(role, instance)
	if err = CreateOrUpdate(ctx, r.Client, "role", cr, muteRole(role, instance), logger); err != nil {
		return err
	}

	initRoleBinding(rb, instance)
	if err = CreateOrUpdate(ctx, r.Client, "roleBinding", rb,
		muteRoleBinding(rb, instance), logger); err != nil {
		return err
	}

	return nil
}

func (r *HigressGatewayReconciler) createDeployment(ctx context.Context, instance *operatorv1alpha1.HigressGateway, logger logr.Logger) error {
	deploy := initDeployment(&appsv1.Deployment{}, instance)
	if err := ctrl.SetControllerReference(instance, deploy, r.Scheme); err != nil {
		return err
	}

	return CreateOrUpdate(ctx, r.Client, "Deployment", deploy, muteDeployment(deploy, instance), logger)
}

func (r *HigressGatewayReconciler) createService(ctx context.Context, instance *operatorv1alpha1.HigressGateway, logger logr.Logger) error {
	svc := initService(&apiv1.Service{}, instance)
	if err := ctrl.SetControllerReference(instance, svc, r.Scheme); err != nil {
		return err
	}

	return CreateOrUpdate(ctx, r.Client, "Service", svc, muteService(svc, instance), logger)
}

func (r *HigressGatewayReconciler) finalizeHigressGateway(instance *operatorv1alpha1.HigressGateway, logger logr.Logger) error {
	var (
		ctx = context.TODO()
		crb = &rbacv1.ClusterRoleBinding{}
	)

	name := getServiceAccount(instance)
	nn := types.NamespacedName{Name: name, Namespace: apiv1.NamespaceAll}
	if err := r.Get(ctx, nn, crb); err != nil {
		return err
	}

	var subjects []rbacv1.Subject
	for _, subject := range crb.Subjects {
		if subject.Name != name || subject.Namespace != instance.Namespace {
			subjects = append(subjects, subject)
		}
	}
	crb.Subjects = subjects
	if err := r.Update(ctx, crb); err != nil {
		return err
	}

	return nil
}

func (r *HigressGatewayReconciler) createConfigMap(ctx context.Context, instance *operatorv1alpha1.HigressGateway, logger logr.Logger) error {
	gatewayConfigMap, err := initGatewayConfigMap(&apiv1.ConfigMap{}, instance)
	if err != nil {
		return err
	}
	if err = CreateOrUpdate(ctx, r.Client, "gatewayConfigMap", gatewayConfigMap,
		muteConfigMap(gatewayConfigMap, instance, updateGatewayConfigMapSpec), logger); err != nil {
		return err
	}

	if instance.Spec.Skywalking.Enable {
		skywalkingConfigMap, err := initSkywalkingConfigMap(&apiv1.ConfigMap{}, instance)
		if err != nil {
			return err
		}

		if err = ctrl.SetControllerReference(instance, skywalkingConfigMap, r.Scheme); err != nil {
			return err
		}

		if err = CreateOrUpdate(ctx, r.Client, "skywalkingConfigMap", skywalkingConfigMap,
			muteConfigMap(skywalkingConfigMap, instance, updateSkywalkingConfigMap), logger); err != nil {
			return err
		}
	}

	return nil
}

func (r *HigressGatewayReconciler) setDefaultValues(instance *operatorv1alpha1.HigressGateway) {
	if instance.Spec.RBAC == nil {
		instance.Spec.RBAC = &operatorv1alpha1.RBAC{Enable: true}
	}
	// serviceAccount
	if instance.Spec.ServiceAccount == nil {
		instance.Spec.ServiceAccount = &operatorv1alpha1.ServiceAccount{Enable: true, Name: "higress-gateway"}
	}
	// replicas
	if instance.Spec.Replicas == nil {
		replicas := int32(1)
		instance.Spec.Replicas = &replicas
	}
	// selectorLabels
	if len(instance.Spec.SelectorLabels) == 0 {
		instance.Spec.SelectorLabels = map[string]string{
			"app":     "higress-gateway",
			"higress": "higress-system-higress-gateway",
		}
	}
	// service
	if instance.Spec.Service == nil {
		instance.Spec.Service = &operatorv1alpha1.Service{
			Type: "LoadBalancer",
			Ports: []apiv1.ServicePort{
				{
					Name:       "http2",
					Port:       80,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(80),
				},
				{
					Name:       "https",
					Port:       443,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(443),
				},
			},
		}
	}
	// skywalking
	if instance.Spec.Skywalking == nil {
		instance.Spec.Skywalking = &operatorv1alpha1.Skywalking{Enable: false}
	}
}
