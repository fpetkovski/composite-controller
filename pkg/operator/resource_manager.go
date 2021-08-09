package operator

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fpetkovski/composite-controller/pkg/apis/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type resourceManager struct {
	k8sClient client.Client
}

func (rm resourceManager) getManagedComponents(
	composite v1alpha1.Composite,
	resourceTypes []client.Object,
) ([]client.Object, error) {
	var components []client.Object

	for _, t := range resourceTypes {
		currentList := unstructured.UnstructuredList{}
		currentList.SetGroupVersionKind(t.GetObjectKind().GroupVersionKind())
		if err := rm.k8sClient.List(context.Background(), &currentList); err != nil {
			return nil, err
		}
		for _, item := range currentList.Items {
			item := item
			if manages(&composite, &item) {
				components = append(components, &item)
			}
		}
	}
	return components, nil
}


func manages(controller client.Object, object client.Object) bool {
	owner := makeOwnerReference(controller)
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

func manage(controller client.Object, object client.Object) {
	owner := makeOwnerReference(controller)
	object.SetOwnerReferences([]metav1.OwnerReference{owner})
}


func makeOwnerReference(controller client.Object) metav1.OwnerReference {
	gvk := controller.GetObjectKind().GroupVersionKind()

	trueVal := true
	return metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               controller.GetName(),
		UID:                controller.GetUID(),
		BlockOwnerDeletion: &trueVal,
		Controller:         &trueVal,
	}
}
