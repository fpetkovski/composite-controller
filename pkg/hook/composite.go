package hook

import (
	"github.com/fpetkovski/composite-controller/pkg/apis/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetChildren(composite v1alpha1.Composite) []client.Object {
	appLabels := map[string]string{
		"app": composite.GetName(),
	}
	return []client.Object{
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      composite.GetName(),
				Namespace: composite.GetNamespace(),
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name: "web",
						Port: 80,
						TargetPort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 80,
						},
					},
				},
			},
		},
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      composite.GetName(),
				Namespace: composite.GetNamespace(),
				Labels:    appLabels,
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "nginx",
						Image: "nginx",
					},
				},
			},
		},
	}
}
