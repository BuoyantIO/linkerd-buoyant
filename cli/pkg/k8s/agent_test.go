package k8s

import (
	"context"
	"errors"
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAgent(t *testing.T) {
	fixtures := []*struct {
		testName string
		objs     []runtime.Object
		expErr   error
		expAgent *Agent
	}{
		{
			"no objects found",
			nil,
			nil,
			nil,
		},
		{
			"secret found",
			[]runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{Name: agentSecret, Namespace: Namespace},
					Data: map[string][]byte{
						"name":        []byte("fake-name"),
						"id":          []byte("fake-id"),
						"downloadKey": []byte("fake-downloadKey"),
					},
				},
			},
			nil,
			&Agent{
				Name: "fake-name",
				URL:  "/agent/buoyant-cloud-k8s-fake-name-fake-id-fake-downloadKey.yml",
			},
		},
		{
			"secret and deployment found",
			[]runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{Name: agentSecret, Namespace: Namespace},
					Data: map[string][]byte{
						"name":        []byte("fake-name"),
						"id":          []byte("fake-id"),
						"downloadKey": []byte("fake-downloadKey"),
					},
				},
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      AgentName,
						Namespace: Namespace,
						Labels:    map[string]string{VersionLabel: "fake-version"},
					},
				},
			},
			nil,
			&Agent{
				Name:    "fake-name",
				URL:     "/agent/buoyant-cloud-k8s-fake-name-fake-id-fake-downloadKey.yml",
				Version: "fake-version",
			},
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.TODO()
			fakeCS := fake.NewSimpleClientset(tc.objs...)
			client := client{fakeCS, ""}

			agent, err := client.Agent(ctx)
			if !errors.Is(err, tc.expErr) {
				t.Errorf("Expected: [%s], got: [%s]", tc.expErr, err)
			}
			if !reflect.DeepEqual(agent, tc.expAgent) {
				t.Errorf("Expected: [%+v], got: [%+v]", tc.expAgent, agent)
			}
		})
	}
}
