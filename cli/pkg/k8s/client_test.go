package k8s

import (
	"context"
	"reflect"
	"testing"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNamespace(t *testing.T) {
	fixtures := []*struct {
		testName string
		objs     []runtime.Object
		expErr   error
		expNS    string
	}{
		{
			"not found",
			nil,
			apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "namespaces"}, k8s.AgentNamespace),
			"",
		},
		{
			"found",
			[]runtime.Object{&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: k8s.AgentNamespace}}},
			nil,
			k8s.AgentNamespace,
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.TODO()
			fakeCS := fake.NewSimpleClientset(tc.objs...)
			client := client{fakeCS}

			ns, err := client.Namespace(ctx)
			if !reflect.DeepEqual(err, tc.expErr) {
				t.Errorf("Expected: [%s], got: [%s]", tc.expErr, err)
			}
			nsName := ""
			if ns != nil {
				nsName = ns.Name
			}
			if nsName != tc.expNS {
				t.Errorf("Expected: [%s], got: [%s]", tc.expNS, nsName)
			}
		})
	}
}
