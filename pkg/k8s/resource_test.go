package k8s

import (
	"context"
	"errors"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestResources(t *testing.T) {
	fixtures := []*struct {
		testName     string
		objs         []runtime.Object
		expErr       error
		expResources []string
	}{
		{
			"no objects found",
			nil,
			nil,
			[]string{},
		},
		{
			"namespace not found with incorrect label value",
			[]runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name:   Namespace,
						Labels: map[string]string{PartOfKey: PartOfVal + "bad"},
					},
				},
			},
			nil,
			[]string{},
		},
		{
			"namespace found",
			[]runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name:   Namespace,
						Labels: map[string]string{PartOfKey: PartOfVal},
					},
				},
			},
			nil,
			[]string{
				`apiVersion: v1
kind: Namespace
metadata:
  creationTimestamp: null
  name: buoyant-cloud
`,
			},
		},
		{
			"namespace, cr, crb found",
			[]runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name:   Namespace,
						Labels: map[string]string{PartOfKey: PartOfVal},
					},
				},
				&rbacv1.ClusterRole{
					ObjectMeta: metav1.ObjectMeta{
						Name:   AgentName,
						Labels: map[string]string{PartOfKey: PartOfVal},
					},
				},
				&rbacv1.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:   AgentName,
						Labels: map[string]string{PartOfKey: PartOfVal},
					},
				},
			},
			nil,
			[]string{
				`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  name: buoyant-cloud-agent
`,
				`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  creationTimestamp: null
  name: buoyant-cloud-agent
`,
				`apiVersion: v1
kind: Namespace
metadata:
  creationTimestamp: null
  name: buoyant-cloud
`,
			},
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.TODO()
			fakeCS := fake.NewSimpleClientset(tc.objs...)
			client := client{fakeCS, ""}

			resources, err := client.Resources(ctx)
			if !errors.Is(err, tc.expErr) {
				t.Errorf("Expected: [%s], got: [%s]", tc.expErr, err)
			}
			if !reflect.DeepEqual(resources, tc.expResources) {
				t.Errorf("Expected: [%+v], got: [%+v]", tc.expResources, resources)
			}
		})
	}
}
