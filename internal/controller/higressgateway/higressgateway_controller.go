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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/alibaba/higress/api/v1alpha1"
	. "github.com/alibaba/higress/internal/controller"
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
			logger.Info("HigressGateway(%v) resource not found")
			return ctrl.Result{}, err
		}

		logger.Error(err, "Failed to get resource HigressGateway(%v)", req.NamespacedName)
		return ctrl.Result{}, err
	}

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
		logger.Info("create serviceAccount for HigressGateway(%v)", instance.Name)
	}

	return nil
}

func (r *HigressGatewayReconciler) createRBAC(ctx context.Context, instance *operatorv1alpha1.HigressGateway, logger logr.Logger) error {
	var (
		role *rbacv1.Role
		rb   *rbacv1.RoleBinding
		cr   *rbacv1.ClusterRole
		crb  *rbacv1.ClusterRoleBinding
	)

	// reconcile role
	role = initRole(role, instance)
	exists, err := CreateIfNotExits(ctx, r.Client, role)
	if err != nil {
		return err
	}
	if !exists {
		logger.Info("create Role for HigressGateway(%v)", instance.Name)
	}

	// reconcile roleBinding
	rb = initRoleBinding(rb, instance)
	if exists, err = CreateIfNotExits(ctx, r.Client, rb); err != nil {
		return err
	}
	if !exists {
		logger.Info("create clusterRole for HigressGateway(%v)", instance.Name)
	}

	// reconcile clusterRole
	cr = initClusterRole(cr, instance)
	if exists, err = CreateIfNotExits(ctx, r.Client, cr); err != nil {
		return err
	}
	if !exists {
		logger.Info("create clusterRole for HigressGateway(%v)", instance.Name)
	}

	// reconcile clusterRoleBinding
	crb = initClusterRoleBinding(crb, instance)
	if exists, err = CreateIfNotExits(ctx, r.Client, crb); err != nil {
		return err
	}
	if !exists {
		logger.Info("create clusterRole for HigressGateway(%v)", instance.Name)
	}

	return nil
}

func (r *HigressGatewayReconciler) createDeployment(ctx context.Context, instance *operatorv1alpha1.HigressGateway, logger logr.Logger) error {
	deploy := initDeployment(&appsv1.Deployment{}, instance)
	if err := ctrl.SetControllerReference(instance, deploy, r.Scheme); err != nil {
		return err
	}

	return CreateOrUpdate(ctx, r.Client, deploy, muteDeployment(deploy, instance), logger)
}

func (r *HigressGatewayReconciler) createService(ctx context.Context, instance *operatorv1alpha1.HigressGateway, logger logr.Logger) error {
	svc := initService(&apiv1.Service{}, instance)
	if err := ctrl.SetControllerReference(instance, svc, r.Scheme); err != nil {
		return err
	}

	return CreateOrUpdate(ctx, r.Client, svc, muteService(svc, instance), logger)
}

func (r *HigressGatewayReconciler) finalizeHigressGateway(instance *operatorv1alpha1.HigressGateway, logger logr.Logger) error {
	var (
		ctx = context.TODO()
		rb  = &rbacv1.RoleBinding{}
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

	nn = types.NamespacedName{Namespace: instance.Namespace, Name: name}
	if err := r.Get(ctx, nn, rb); err != nil {
		return err
	}

	subjects = []rbacv1.Subject{}
	for _, subject := range crb.Subjects {
		if subject.Name != name || subject.Namespace != instance.Namespace {
			subjects = append(subjects, subject)
		}
	}
	rb.Subjects = subjects
	if err := r.Update(ctx, crb); err != nil {
		return err
	}
	return nil
}
