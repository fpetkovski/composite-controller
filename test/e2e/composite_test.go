package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fpetkovski/composite-controller/pkg/apis/v1alpha1"
	"github.com/fpetkovski/composite-controller/pkg/hook"
	"github.com/fpetkovski/composite-controller/pkg/operator"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2/klogr"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func TestComposite(t *testing.T) {
	c := v1alpha1.Composite{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "composite-test-1",
			Namespace: "default",
		},
		Spec: v1alpha1.CompositeSpec{
			Image: "nginx",
		},
	}
	mapper := hook.Mapper{}
	cfg := controllerruntime.GetConfigOrDie()
	logger := klogr.New()
	op, err := operator.New(cfg, logger, mapper)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		if err := op.Start(controllerruntime.SetupSignalHandler()); err != nil {
			t.Fatal(err)
		}
	}()

	k8sClient := op.GetClient()
	if err := k8sClient.Create(context.Background(), &c); err != nil {
		t.Fatal(err)
	}

	if err := wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		d := appsv1.Deployment{}
		key := types.NamespacedName{Namespace: c.Namespace, Name: c.Name}
		if err := k8sClient.Get(context.Background(), key, &d); err != nil {
			fmt.Println("deployment not found")
			return false, nil
		}

		fmt.Println(d.GetName())
		return true, nil
	}); err != nil {
		t.Fatal(err)
	}
}
