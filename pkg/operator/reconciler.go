package operator

import (
	"context"
	"fmt"

	"github.com/fpetkovski/composite-controller/pkg/apis/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
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

func (r *reconciler) reconcileComponents(composite v1alpha1.Composite, desiredComponents []client.Object) error {
	observedComponents, err := r.getObservedComponents(composite)
	if err != nil {
		return err
	}

	if err := r.deleteObsoleteComponents(observedComponents, desiredComponents); err != nil {
		return err
	}

	for _, c := range desiredComponents {
		if err := r.reconcileComponent(composite, c); err != nil {
			return err
		}
	}
	return nil
}

func (r *reconciler) deleteObsoleteComponents(observedComponents []client.Object, desiredComponents []client.Object) error {
	for _, observed := range observedComponents {
		found := false
		found = findComponent(observed, desiredComponents)
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

func (r *reconciler) getObservedComponents(composite v1alpha1.Composite) ([]client.Object, error) {
	var components []client.Object
	owner := getOwnerReference(composite)
	for _, t := range r.mapper.GetTypes() {
		currentList := unstructured.UnstructuredList{}
		currentList.SetGroupVersionKind(t.GetObjectKind().GroupVersionKind())
		if err := r.k8sClient.List(context.Background(), &currentList); err != nil {
			return nil, err
		}
		for _, item := range currentList.Items {
			item := item
			if isManagedBy(&item, owner) {
				components = append(components, &item)
			}
		}
	}
	return components, nil
}

func (r *reconciler) reconcileComponent(composite v1alpha1.Composite, c client.Object) error {
	ownerReference := getOwnerReference(composite)
	current, err := r.getCurrentObject(c)
	if err != nil {
		return err
	}

	if current != nil && !isManagedBy(current, ownerReference) {
		return fmt.Errorf("object %s/%s is not managed by %s/%s",
			current.GetObjectKind().GroupVersionKind(),
			current.GetName(),
			composite.GroupVersionKind(),
			composite.GetName(),
		)
	}

	c.SetOwnerReferences([]v1.OwnerReference{ownerReference})
	fieldOwner := client.FieldOwner("composite-controller")
	if err := r.k8sClient.Patch(context.Background(), c, client.Apply, fieldOwner); err != nil {
		return err
	}
	return nil
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

func getOwnerReference(composite v1alpha1.Composite) v1.OwnerReference {
	trueVal := true
	return v1.OwnerReference{
		APIVersion:         composite.APIVersion,
		Kind:               composite.Kind,
		Name:               composite.GetName(),
		UID:                composite.GetUID(),
		BlockOwnerDeletion: &trueVal,
		Controller:         &trueVal,
	}
}

func isManagedBy(object client.Object, owner v1.OwnerReference) bool {
	for _, o := range object.GetOwnerReferences() {
		ownerEquals := o.APIVersion == owner.APIVersion &&
			o.Kind == owner.Kind &&
			o.Name == owner.Name &&
			o.UID == owner.UID

		if ownerEquals {
			return true
		}
	}

	return false
}
