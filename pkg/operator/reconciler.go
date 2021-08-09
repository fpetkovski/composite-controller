package operator

import (
	"context"
	"fmt"

	"github.com/fpetkovski/composite-controller/pkg/apis/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type reconciler struct {
	logger          logr.Logger
	k8sClient       client.Client
	mapper          CompositeMapper
	resourceManager resourceManager
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

	observedComponents, err := r.resourceManager.getManagedComponents(composite, r.mapper.GetTypes())
	if err != nil {
		return reconcile.Result{}, err
	}

	if err := r.reconcileComponents(&composite, observedComponents); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.updateStatus(&composite, observedComponents); err != nil {
		return reconcile.Result{}, err
	}

	r.logger.Info("Done reconciling composite", "Name", composite.GetName())
	return reconcile.Result{}, nil
}

func (r *reconciler) reconcileComponents(object client.Object, observedComponents []client.Object) error {
	r.logger.Info("Reconciling object", "Name", object.GetName())

	desiredComponents, err := r.mapper.GetComponents(object, observedComponents)
	if err != nil {
		return err
	}

	if err := r.deleteObsoleteComponents(observedComponents, desiredComponents); err != nil {
		return err
	}

	for _, c := range desiredComponents {
		if err := r.reconcileComponent(object, c); err != nil {
			return err
		}
	}
	return nil
}

func (r *reconciler) deleteObsoleteComponents(observedComponents []client.Object, desiredComponents []client.Object) error {
	for _, observed := range observedComponents {
		found := findComponent(observed, desiredComponents)
		if found {
			continue
		}

		if err := r.k8sClient.Delete(context.Background(), observed); err != nil {
			return err
		}
	}
	return nil
}

func findComponent(observed client.Object, desiredComponents []client.Object) bool {
	for _, desired := range desiredComponents {
		equalGVK := desired.GetObjectKind().GroupVersionKind() == observed.GetObjectKind().GroupVersionKind()
		equalName := desired.GetName() == observed.GetName() && desired.GetNamespace() == observed.GetNamespace()
		if equalGVK && equalName {
			return true
		}
	}
	return false
}

func (r *reconciler) reconcileComponent(composite client.Object, c client.Object) error {
	current, err := r.getCurrentObject(c)
	if err != nil {
		return err
	}

	if current != nil && !manages(composite, current) {
		return fmt.Errorf("object %s/%s is not managed by %s/%s",
			current.GetObjectKind().GroupVersionKind(),
			current.GetName(),
			composite.GetObjectKind().GroupVersionKind(),
			composite.GetName(),
		)
	}

	manage(composite, c)
	return r.k8sClient.Patch(context.Background(), c, client.Apply, client.FieldOwner("composite-controller"))
}

func (r *reconciler) getCurrentObject(c client.Object) (client.Object, error) {
	var current unstructured.Unstructured
	current.SetGroupVersionKind(c.GetObjectKind().GroupVersionKind())

	key := types.NamespacedName{Name: c.GetName(), Namespace: c.GetNamespace()}
	err := r.k8sClient.Get(context.Background(), key, &current)
	if errors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &current, nil
}

func (r *reconciler) updateStatus(controller client.Object, observedComponents []client.Object) error {
	r.logger.Info("Reconciling status", "Name", controller.GetName())
	_, err := r.mapper.GetStatus(controller, observedComponents)
	if err != nil {
		return err
	}

	//controller.Status = status
	return r.k8sClient.Status().Update(context.Background(), controller)
}
