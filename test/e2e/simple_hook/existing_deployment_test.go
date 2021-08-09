package simple_hook

import (
	"context"
	"github.com/fpetkovski/composite-controller/pkg/apis/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestHookWithExistingDeployment(t *testing.T) {
	c := v1alpha1.Composite{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing-deployment",
			Namespace: "default",
		},
		Spec: v1alpha1.CompositeSpec{
			Image: "apache2",
		},
	}
	k8sClient := f.Operator.GetClient()

	deployment := makeDeployment("existing-deployment")
	if err := k8sClient.Create(context.Background(), deployment); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = k8sClient.Delete(context.Background(), deployment)
	}()

	if err := k8sClient.Create(context.Background(), &c); err != nil {
		t.Fatal(err)
	}
	assertDeploymentExists(t, k8sClient, c, "nginx")

	if err := k8sClient.Delete(context.Background(), &c); err != nil {
		t.Fatal(err)
	}
	assertDeploymentDoesNotExist(t, c, k8sClient)
}

func makeDeployment(name string) *appsv1.Deployment {
	replicas := int32(1)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  name,
							Image: "nginx",
						},
					},
				},
			},
		},
	}
}
