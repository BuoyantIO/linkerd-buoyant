package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/yaml"
)

const (
	labelSelector = "app.kubernetes.io/part-of=buoyant-cloud"
)

// Resources returns all linkerd-buoyant resources required for deletion.
// Specifically, these three resource types matching the label
// app.kubernetes.io/part-of=buoyant-cloud:
// - ClusterRole
// - ClusterRoleBinding
// - Namespace
func Resources(ctx context.Context, client kubernetes.Interface) ([]string, error) {
	opts := metav1.ListOptions{LabelSelector: labelSelector}
	resources := []string{}

	crList, err := client.RbacV1().ClusterRoles().List(ctx, opts)
	if err != nil {
		return nil, err
	}
	for _, i := range crList.Items {
		i := i // pin
		y, err := toYaml(runtime.Object(&i), i.ObjectMeta)
		if err != nil {
			return nil, err
		}
		resources = append(resources, string(y))
	}

	crbList, err := client.RbacV1().ClusterRoleBindings().List(ctx, opts)
	if err != nil {
		return nil, err
	}
	for _, i := range crbList.Items {
		i := i // pin
		y, err := toYaml(runtime.Object(&i), i.ObjectMeta)
		if err != nil {
			return nil, err
		}
		resources = append(resources, string(y))
	}

	nsList, err := client.CoreV1().Namespaces().List(ctx, opts)
	if err != nil {
		return nil, err
	}
	for _, i := range nsList.Items {
		i := i // pin
		y, err := toYaml(runtime.Object(&i), i.ObjectMeta)
		if err != nil {
			return nil, err
		}
		resources = append(resources, string(y))
	}

	return resources, nil
}

func toYaml(runobj runtime.Object, objmeta metav1.ObjectMeta) ([]byte, error) {
	gvks, _, err := scheme.Scheme.ObjectKinds(runobj)
	if err != nil {
		return nil, err
	}
	if len(gvks) == 0 {
		return nil, fmt.Errorf("no GroupVersionKind found for %+v", runobj)
	}

	s := struct {
		runtime.TypeMeta
		metav1.ObjectMeta `json:"metadata"`
	}{
		runtime.TypeMeta{
			APIVersion: gvks[0].GroupVersion().String(),
			Kind:       gvks[0].Kind,
		},
		metav1.ObjectMeta{
			Name:      objmeta.GetName(),
			Namespace: objmeta.GetNamespace(),
		},
	}

	return yaml.Marshal(s)
}
