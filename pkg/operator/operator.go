package operator

import (
	"context"
	"github.com/fpetkovski/composite-controller/pkg/apis/v1alpha1"
	"github.com/fpetkovski/composite-controller/pkg/hook"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func New(cfg *rest.Config, logger logr.Logger) (manager.Manager, error) {
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		return nil, err
	}

	if err := v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	ctrlLogger := logger.WithName("composite-controller")
	ctrl, err := controller.New("composite-controller", mgr, controller.Options{
		MaxConcurrentReconciles: 1,
		Reconciler: &reconciler{
			k8sClient: mgr.GetClient(),
			logger:    ctrlLogger,
		},
		Log: ctrlLogger,
	})
	if err != nil {
		return nil, err
	}

	if err := ctrl.Watch(&source.Kind{Type: &v1alpha1.Composite{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return nil, err
	}

	return mgr, nil
}

type reconciler struct {
	logger    logr.Logger
	k8sClient client.Client
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

	children := hook.GetChildren(composite)
	if composite.DeletionTimestamp != nil {
		r.logger.Info("Deleting composite", "Name", composite.GetName())
		return r.deleteChildren(children)
	}

	r.logger.Info("Reconciling composite", "Name", composite.GetName())

	return r.reconcileChildren(composite, children)
}

func (r *reconciler) reconcileChildren(composite v1alpha1.Composite, children []client.Object) (reconcile.Result, error) {
	true := true
	for _, c := range children {
		c.SetOwnerReferences([]v1.OwnerReference{
			{
				APIVersion:         composite.APIVersion,
				Kind:               composite.Kind,
				Name:               composite.GetName(),
				UID:                composite.GetUID(),
				BlockOwnerDeletion: &true,
				Controller:         &true,
			},
		})
		if err := r.k8sClient.Patch(
			context.Background(),
			c,
			client.Apply,
			client.FieldOwner("composite-controller"),
		); err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *reconciler) deleteChildren(children []client.Object) (reconcile.Result, error) {
	for _, c := range children {
		if err := r.k8sClient.Delete(context.Background(), c); err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}
