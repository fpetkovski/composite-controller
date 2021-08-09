package simple_hook

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fpetkovski/composite-controller/pkg/apis/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

func TestSimpleHook(t *testing.T) {
	c := v1alpha1.Composite{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "composite-test-1",
			Namespace: "default",
		},
		Spec: v1alpha1.CompositeSpec{
			Image: "apache2",
		},
	}

	k8sClient := f.Operator.GetClient()
	if err := k8sClient.Create(context.Background(), &c); err != nil {
		t.Fatal(err)
	}
	assertDeploymentExists(t, k8sClient, c, c.Spec.Image)

	if err := k8sClient.Delete(context.Background(), &c); err != nil {
		t.Fatal(err)
	}
	assertDeploymentDoesNotExist(t, c, k8sClient)
}

func assertDeploymentExists(t *testing.T, k8sClient client.Client, c v1alpha1.Composite, expectedImage string) {
	if err := wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		d := appsv1.Deployment{}
		key := types.NamespacedName{Namespace: c.Namespace, Name: c.Name}
		if err := k8sClient.Get(context.Background(), key, &d); err != nil {
			return false, nil
		}

		if len(d.Spec.Template.Spec.Containers) != 1 {
			return false, nil
		}

		imageMatch := d.Spec.Template.Spec.Containers[0].Image == expectedImage
		return imageMatch, nil
	}); err != nil {
		t.Fatal(err)
	}
}

func assertDeploymentDoesNotExist(t *testing.T, c v1alpha1.Composite, k8sClient client.Client) {
	if err := wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		d := appsv1.Deployment{}
		key := types.NamespacedName{Namespace: c.Namespace, Name: c.Name}
		if err := k8sClient.Get(context.Background(), key, &d); errors.IsNotFound(err) {
			return true, nil
		}

		return false, nil
	}); err != nil {
		t.Fatal(err)
	}
}
