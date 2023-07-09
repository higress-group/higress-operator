package controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func CreateOrUpdate(ctx context.Context, cli client.Client, object client.Object, f controllerutil.MutateFn, logger logr.Logger) error {
	kind, key := object.GetObjectKind().GroupVersionKind().Kind, client.ObjectKeyFromObject(object)
	status, err := controllerutil.CreateOrUpdate(ctx, cli, object, f)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to createOrUpdate object {%s:%s}", kind, key))
		return err
	}

	logger.Info(fmt.Sprintf("createOrUpdate object {%s:%s} %s", kind, key, status))
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
