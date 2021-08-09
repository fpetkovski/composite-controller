package hook

import (
	"fmt"

	"github.com/fpetkovski/composite-controller/pkg/apis/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Mapper struct {
}

func (m Mapper) GetTypes() []client.Object {
	return []client.Object{
		&appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
		},
		&v1.Service{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
		},
		&networkingv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "networking.k8s.io/v1",
				Kind:       "Ingress",
			},
		},
	}
}

func (m Mapper) GetComponents(object client.Object, observed []client.Object) ([]client.Object, error) {
	composite, ok := object.(*v1alpha1.Composite)
	if !ok {
		return nil, fmt.Errorf("object is not of type v1alpha1.Composite")
	}

	ingressPathType := networkingv1.PathTypePrefix
	portName := "web"
	podLabels := map[string]string{
		"app": composite.GetName(),
	}

	return []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      composite.GetName(),
				Namespace: composite.GetNamespace(),
				Annotations: map[string]string{
					"kubectl.kubernetes.io/default-container": "main",
				},
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: podLabels,
				},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: podLabels,
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "main",
								Image: composite.Spec.Image,
							},
						},
					},
				},
			},
		},
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
				Selector: podLabels,
				Ports: []v1.ServicePort{
					{
						Name: portName,
						Port: 80,
						TargetPort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 80,
						},
					},
				},
			},
		},
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      composite.GetName(),
				Namespace: composite.GetNamespace(),
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: "networking.k8s.io/v1",
				Kind:       "Ingress",
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						Host: fmt.Sprintf("%s.kind.local", composite.GetName()),
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										Path:     "/",
										PathType: &ingressPathType,
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: composite.GetName(),
												Port: networkingv1.ServiceBackendPort{
													Name: portName,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func (m Mapper) GetStatus(object client.Object, observedComponents []client.Object) (v1alpha1.CompositeStatus, error) {
	objects, err := m.GetComponents(object, observedComponents)
	if err != nil {
		return v1alpha1.CompositeStatus{}, err
	}

	return v1alpha1.CompositeStatus{
		ManagedTypes:   len(m.GetTypes()),
		ManagedObjects: len(objects),
	}, nil
}
