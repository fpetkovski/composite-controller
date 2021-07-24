package operator

import (
	"context"

	"github.com/fpetkovski/composite-controller/pkg/apis/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type reconciler struct {
	logger    logr.Logger
	k8sClient client.Client
	mapper    CompositeMapper
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	var composite v1alpha1.Composite
	err := r.k8sClient.Get(ctx, request.NamespacedName, &composite)
	if errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}
	if err != nil {
		return reconcile.Result{}, err
	}

	if composite.DeletionTimestamp != nil {
		return reconcile.Result{}, err
	}

	r.logger.Info("Reconciling composite", "Name", composite.GetName())
	components := r.mapper.GetComponents(composite)
	if err := r.reconcileComponents(composite, components); err != nil {
		return reconcile.Result{}, err
	}

	r.logger.Info("Reconciling status", "Name", composite.GetName())
	composite.Status = r.mapper.GetStatus(composite)
	if err := r.k8sClient.Status().Update(context.Background(), &composite); err != nil {
		return reconcile.Result{}, err
	}

	r.logger.Info("Done reconciling composite", "Name", composite.GetName())
	return reconcile.Result{}, nil
}

func (r *reconciler) reconcileComponents(composite v1alpha1.Composite, components []client.Object) error {
	fieldOwner := client.FieldOwner("composite-controller")
	ownerReferences := getOwnerReferences(composite)
	for _, c := range components {
		c.SetOwnerReferences(ownerReferences)
		ctx := context.Background()
		if err := r.k8sClient.Patch(ctx, c, client.Apply, fieldOwner); err != nil {
			return err
		}
	}
	return nil
}

func getOwnerReferences(composite v1alpha1.Composite) []v1.OwnerReference {
	trueVal := true
	return []v1.OwnerReference{
		{
			APIVersion:         composite.APIVersion,
			Kind:               composite.Kind,
			Name:               composite.GetName(),
			UID:                composite.GetUID(),
			BlockOwnerDeletion: &trueVal,
			Controller:         &trueVal,
		},
	}
}
