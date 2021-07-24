package operator

import (
	"github.com/fpetkovski/composite-controller/pkg/apis/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type CompositeMapper interface {
	GetTypes() []client.Object
	GetComponents(composite v1alpha1.Composite) []client.Object
	GetStatus(composite v1alpha1.Composite) v1alpha1.CompositeStatus
}

func New(cfg *rest.Config, logger logr.Logger, mapper CompositeMapper) (manager.Manager, error) {
	mgr, err := manager.New(cfg, manager.Options{
		MetricsBindAddress: "0",
	})
	if err != nil {
		return nil, err
	}

	if err := v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	ctrlLogger := logger.WithName("reconciler")
	ctrl, err := controller.New("composite-controller", mgr, controller.Options{
		MaxConcurrentReconciles: 1,
		Reconciler: &reconciler{
			k8sClient: mgr.GetClient(),
			logger:    ctrlLogger,
			mapper:    mapper,
		},
		Log: logger,
	})
	if err != nil {
		return nil, err
	}

	if err := ctrl.Watch(
		&source.Kind{Type: &v1alpha1.Composite{}},
		&handler.EnqueueRequestForObject{},
		predicate.GenerationChangedPredicate{},
	); err != nil {
		return nil, err
	}

	enqueueForOwnerHandler := handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType: &v1alpha1.Composite{},
	}
	for _, t := range mapper.GetTypes() {
		if err := ctrl.Watch(&source.Kind{Type: t}, &enqueueForOwnerHandler); err != nil {
			return nil, err
		}
	}

	return mgr, nil
}
