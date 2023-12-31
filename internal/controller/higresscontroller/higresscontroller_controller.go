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

package higresscontroller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apixv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/alibaba/higress/higress-operator/api/v1alpha1"
	. "github.com/alibaba/higress/higress-operator/internal/controller"
)

const (
	finalizer = "higresscontroller.higress.io/finalizer"
)

// HigressControllerReconciler reconciles a HigressController object
type HigressControllerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config *rest.Config
}

//+kubebuilder:rbac:groups=operator.higress.io,resources=higresscontrollers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.higress.io,resources=higresscontrollers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.higress.io,resources=higresscontrollers/finalizers,verbs=update

//+kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods;services;services/finalizers;endpoints;persistentvolumeclaims;events;configmaps;secrets;serviceaccounts;namespaces,verbs=create;update;get;list;watch;patch;delete
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch

//+kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=mutatingwebhookconfigurations,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=validatingwebhookconfigurations,verbs=get;list;watch;update

//+kubebuilder:rbac:groups="authentication.istio.io",resources=*,verbs=get;list;watch

//+kubebuilder:rbac:groups="certificates.k8s.io",resources=certificatesigningrequests,verbs=update;create;get;delete;watch
//+kubebuilder:rbac:groups="certificates.k8s.io",resources=certificatesigningrequests/approval,verbs=update;create;get;delete;watch
//+kubebuilder:rbac:groups="certificates.k8s.io",resources=certificatesigningrequests/status,verbs=update;create;get;delete;watch
//+kubebuilder:rbac:groups="certificates.k8s.io",resources=signers;resourceNames=certificatesigningrequests/status,verbs=update;create;get;delete;watch
//+kubebuilder:rbac:groups="certificates.k8s.io",resources=signers;resourceNames=kubernetes.io/legacy-unknown,verbs=approve

//+kubebuilder:rbac:groups="config.istio.io",resources=*,verbs=get;list;watch
//+kubebuilder:rbac:groups="discovery.k8s.io",resources=endpointslices,verbs=get;list;watch

//+kubebuilder:rbac:groups="extensions",resources=ingressclasses,verbs=get;list;watch
//+kubebuilder:rbac:groups="extensions",resources=ingresses,verbs=get;list;watch
//+kubebuilder:rbac:groups="extensions",resources=ingresses/status,verbs=*

//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingressclasses,verbs=get;list;watch
//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=get;list;watch
//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses/status,verbs=*

//+kubebuilder:rbac:groups="extensions.higress.io",resources=wasmplugins,verbs=get;create;watch;list;update;patch
//+kubebuilder:rbac:groups="extensions.higress.io",resources=*,verbs=get;list;watch

//+kubebuilder:rbac:groups="gateway.networking.k8s.io",resources=*,verbs=get;list;watch;update

//+kubebuilder:rbac:groups="multicluster.x-k8s.io",resources=serviceexports,verbs=get;list;watch;update;create;delete
//+kubebuilder:rbac:groups="multicluster.x-k8s.io",resources=serviceimports,verbs=get;watch;list

//+kubebuilder:rbac:groups="networking.higress.io",resources=http2rpcs,verbs=get;create;watch;list;update;patch
//+kubebuilder:rbac:groups="networking.higress.io",resources=mcpbridges,verbs=get;create;watch;list;update;patch

//+kubebuilder:rbac:groups="networking.istio.io",resources=*,verbs=get;list;watch
//+kubebuilder:rbac:groups="rbac.istio.io",resources=*,verbs=get;list;watch
//+kubebuilder:rbac:groups="security.istio.io",resources=*,verbs=get;list;watch
//+kubebuilder:rbac:groups="telemetry.istio.io",resources=*,verbs=get;list;watch
//+kubebuilder:rbac:groups="extensions.istio.io",resources=*,verbs=get;list;watch
//+kubebuilder:rbac:groups="networking.x-k8s.io",resources=*,verbs=get;list;watch;update

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=list;watch;get
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingressclasses,verbs=get;create;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=update
//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;create;delete;update;watch;list
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings;roles;rolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HigressController object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *HigressControllerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	instance := &operatorv1alpha1.HigressController{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("HigressController(%v) resource not found. Ignoring since object must be deleted", req.NamespacedName))
			return ctrl.Result{}, nil
		}

		logger.Error(err, fmt.Sprintf("Failed to get resource HigressController(%v)", req.NamespacedName))
		return ctrl.Result{}, err
	}

	r.setDefaultValues(instance)

	// if DeletionTimestamp is not nil, it means is marked to be deleted
	if instance.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(instance, finalizer) {
			if err := r.finalizeHigressController(instance, logger); err != nil {
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
			logger.Error(err, "Failed to add finalizer for higressController")
			return ctrl.Result{}, err
		}
	}

	if err = r.createCRDs(ctx, logger); err != nil {
		logger.Error(err, "Failed to create crds")
		return ctrl.Result{}, err
	}

	if err = r.createServiceAccount(ctx, instance, logger); err != nil {
		logger.Error(err, "Failed to create serviceAccount")
		return ctrl.Result{}, err
	}

	if err = r.createRBAC(ctx, instance, logger); err != nil {
		logger.Error(err, "Failed to create rbac")
		return ctrl.Result{}, err
	}

	if err = r.createDeployment(ctx, instance, logger); err != nil {
		logger.Error(err, "Failed to create deployment")
		return ctrl.Result{}, err
	}

	if err = r.createService(ctx, instance, logger); err != nil {
		logger.Error(err, "Failed to create service")
		return ctrl.Result{}, err
	}

	if !instance.Status.Deployed {
		instance.Status.Deployed = true
		if err = r.Status().Update(ctx, instance); err != nil {
			logger.Error(err, "Failed to update higressController/status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HigressControllerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.HigressController{}).
		Owns(&appsv1.Deployment{}).
		Owns(&apiv1.Service{}).
		Owns(&apiv1.ServiceAccount{}).
		Complete(r)
}

func (r *HigressControllerReconciler) createServiceAccount(ctx context.Context, instance *operatorv1alpha1.HigressController, logger logr.Logger) error {
	if !instance.Spec.ServiceAccount.Enable {
		return nil
	}

	var (
		sa  = &apiv1.ServiceAccount{}
		err error
	)

	sa = initServiceAccount(sa, instance)
	if err = ctrl.SetControllerReference(instance, sa, r.Scheme); err != nil {
		return err
	}

	exist, err := CreateIfNotExits(ctx, r.Client, sa)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to create serviceAccount for HigressController(%v)", instance.Name))
		return err
	}

	if !exist {
		logger.Info(fmt.Sprintf("Create servieAccount for HigressController(%v)", instance.Name))
	}

	return nil
}

func (r *HigressControllerReconciler) createRBAC(ctx context.Context, instance *operatorv1alpha1.HigressController, log logr.Logger) error {
	if instance.Spec.RBAC == nil {
		instance.Spec.RBAC = &operatorv1alpha1.RBAC{Enable: true}
	}
	if instance.Spec.ServiceAccount == nil {
		instance.Spec.ServiceAccount = &operatorv1alpha1.ServiceAccount{Enable: true}
	}
	if !instance.Spec.RBAC.Enable || !instance.Spec.ServiceAccount.Enable {
		return nil
	}

	var (
		role = &rbacv1.Role{}
		rb   = &rbacv1.RoleBinding{}
		cr   = &rbacv1.ClusterRole{}
		crb  = &rbacv1.ClusterRoleBinding{}
	)

	initClusterRole(cr, instance)
	if err := CreateOrUpdate(ctx, r.Client, "ClusterRole", cr, muteClusterRole(cr, instance), log); err != nil {
		return err
	}

	initClusterRoleBinding(crb, instance)
	if err := CreateOrUpdate(ctx, r.Client, "ClusterRoleBinding", crb, muteClusterRoleBinding(crb, instance), log); err != nil {
		return err
	}

	initRole(role, instance)
	if err := CreateOrUpdate(ctx, r.Client, "role", role, muteRole(role, instance), log); err != nil {
		return err
	}

	initRoleBinding(rb, instance)
	if err := CreateOrUpdate(ctx, r.Client, "roleBinding", rb, muteRoleBinding(rb, instance), log); err != nil {
		return err
	}

	return nil
}

func (r *HigressControllerReconciler) createDeployment(ctx context.Context, instance *operatorv1alpha1.HigressController, logger logr.Logger) error {
	deploy := initDeployment(&appsv1.Deployment{}, instance)
	if err := ctrl.SetControllerReference(instance, deploy, r.Scheme); err != nil {
		return err
	}

	return CreateOrUpdate(ctx, r.Client, "Deployment", deploy, muteDeployment(deploy, instance), logger)
}

func (r *HigressControllerReconciler) createService(ctx context.Context, instance *operatorv1alpha1.HigressController, logger logr.Logger) error {
	svc := initService(&apiv1.Service{}, instance)
	if err := ctrl.SetControllerReference(instance, svc, r.Scheme); err != nil {
		return err
	}

	return CreateOrUpdate(ctx, r.Client, "Service", svc, muteService(svc, instance), logger)
}

func (r *HigressControllerReconciler) finalizeHigressController(instance *operatorv1alpha1.HigressController, logger logr.Logger) error {
	if !instance.Spec.RBAC.Enable || !instance.Spec.ServiceAccount.Enable {
		return nil
	}

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
	return r.Update(ctx, crb)
}

func (r *HigressControllerReconciler) createCRDs(ctx context.Context, logger logr.Logger) error {
	apixClient, err := apixv1client.NewForConfig(r.Config)
	if err != nil {
		return err
	}

	crds, err := getCRDs()
	if err != nil {
		return err
	}

	cli := apixClient.CustomResourceDefinitions()
	for _, crd := range crds {
		if existing, err := cli.Get(ctx, crd.Name, metav1.GetOptions{TypeMeta: crd.TypeMeta}); err != nil {
			if !errors.IsNotFound(err) {
				logger.Error(err, fmt.Sprintf("failed to get CRD %v", crd.Name))
				return err
			}
			if _, err = cli.Create(ctx, crd, metav1.CreateOptions{TypeMeta: crd.TypeMeta}); err != nil {
				logger.Error(err, fmt.Sprintf("failed to create CRD %v", crd.Name))
				return err
			}
		} else if !equality.Semantic.DeepEqual(existing.Spec, crd.Spec) {
			// todo(lql): We should check if it has changed before updating it
			existing.Spec = crd.Spec
			if _, err = cli.Update(ctx, existing, metav1.UpdateOptions{TypeMeta: crd.TypeMeta}); err != nil {
				logger.Error(err, fmt.Sprintf("failed to update CRD %v", crd.Name))
				return err
			}
		}
	}
	return nil
}

func (r *HigressControllerReconciler) setDefaultValues(instance *operatorv1alpha1.HigressController) {
	if instance.Spec.RBAC == nil {
		instance.Spec.RBAC = &operatorv1alpha1.RBAC{Enable: true}
	}
	// serviceAccount
	if instance.Spec.ServiceAccount == nil {
		instance.Spec.ServiceAccount = &operatorv1alpha1.ServiceAccount{Enable: true, Name: "higress-controller"}
	}
	// SelectorLabels
	if len(instance.Spec.SelectorLabels) == 0 {
		instance.Spec.SelectorLabels = map[string]string{
			"app": "higress-controller",
		}
	}
}
