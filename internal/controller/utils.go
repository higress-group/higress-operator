package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func CreateOrUpdate(ctx context.Context, cli client.Client, kind string, object client.Object, f controllerutil.MutateFn, logger logr.Logger) error {
	key := client.ObjectKeyFromObject(object)
	status, err := controllerutil.CreateOrUpdate(ctx, cli, object, f)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to createOrUpdate object {%s:%s}", kind, key))
		return err
	}

	logger.Info(fmt.Sprintf("createOrUpdate object {%s:%s} : %s", kind, key, status))
	return nil
}

// CreateIfNotExits create obj if not exists
// true, nil : the obj exists
func CreateIfNotExits(ctx context.Context, cli client.Client, object client.Object) (bool, error) {
	var err error
	if err = cli.Create(ctx, object); err != nil && errors.IsAlreadyExists(err) {
		return true, nil
	}

	return false, err
}

func createOrUpdate(ctx context.Context, c client.Client, obj client.Object, f controllerutil.MutateFn, logger logr.Logger) (controllerutil.OperationResult, error) {
	key := client.ObjectKeyFromObject(obj)
	if err := c.Get(ctx, key, obj); err != nil {
		if !errors.IsNotFound(err) {
			return controllerutil.OperationResultNone, err
		}
		if err := mutate(f, key, obj); err != nil {
			return controllerutil.OperationResultNone, err
		}
		if err := c.Create(ctx, obj); err != nil {
			return controllerutil.OperationResultNone, err
		}
		return controllerutil.OperationResultCreated, nil
	}

	existing := obj.DeepCopyObject()
	if err := mutate(f, key, obj); err != nil {
		return controllerutil.OperationResultNone, err
	}

	if equality.Semantic.DeepEqual(existing, obj) {
		return controllerutil.OperationResultNone, nil
	}

	logger.Info(fmt.Sprintf("the diff of %v is %v", key, cmp.Diff(obj, existing)))

	if err := c.Update(ctx, obj); err != nil {
		return controllerutil.OperationResultNone, err
	}
	return controllerutil.OperationResultUpdated, nil
}

func mutate(f controllerutil.MutateFn, key client.ObjectKey, obj client.Object) error {
	if err := f(); err != nil {
		return err
	}
	if newKey := client.ObjectKeyFromObject(obj); key != newKey {
		return fmt.Errorf("MutateFn cannot mutate object name and/or object namespace")
	}
	return nil
}

func UpdateObjectMeta(obj *metav1.ObjectMeta, instance metav1.Object, labels map[string]string) {
	obj.Name = instance.GetName()
	obj.Namespace = instance.GetNamespace()
	obj.Labels = labels
}
