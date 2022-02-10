package registrator

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	k8stest "k8s.io/client-go/testing"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/bcloudapi"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	clientID     = "client-id"
	clientSecret = "client-secret"
	accessToken  = "eyJhbGciOiJIUzI6IkpXVCJ9.eyJzdWIiOiIxMjMNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2Q6yJV_adQssw5c"
	newAgentID   = "new-agent-id"
)

func TestRegistrator(t *testing.T) {
	fixtures := []*struct {
		testName           string
		configMap          *corev1.ConfigMap
		expInfo            *bcloudapi.AgentInfo
		expErr             error
		expectRegistration bool
	}{
		{
			"manifest fully hydrated (does not perform registration)",
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      agentMetadataConfigMap,
					Namespace: k8s.AgentNamespace,
				},
				Data: map[string]string{
					k8s.AgentIDKey:   "agent-id",
					k8s.AgentNameKey: "agent-name",
				},
			},
			&bcloudapi.AgentInfo{
				AgentName:  "agent-name",
				AgentID:    "agent-id",
				IsNewAgent: false,
			},
			nil,
			false,
		},
		{
			"only name present (performs registration)",
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      agentMetadataConfigMap,
					Namespace: k8s.AgentNamespace,
				},
				Data: map[string]string{
					k8s.AgentNameKey: "agent-name",
				},
			},
			&bcloudapi.AgentInfo{
				AgentName:  "agent-name",
				AgentID:    newAgentID,
				IsNewAgent: true,
			},
			nil,
			true,
		},
		{
			"missing agent name",
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      agentMetadataConfigMap,
					Namespace: k8s.AgentNamespace,
				},
				Data: map[string]string{
					k8s.AgentIDKey: "agent-id",
				},
			},
			nil,
			fmt.Errorf("%s config map needs to have an agent_name key", agentMetadataConfigMap),
			false,
		},
		{
			"missing config map",
			nil,
			nil,
			fmt.Errorf("could not find %s config map", agentMetadataConfigMap),
			false,
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			objs := []runtime.Object{}
			if tc.configMap != nil {
				objs = append(objs, tc.configMap)
			}
			cs := fake.NewSimpleClientset(objs...)
			k8sApi := &l5dk8s.KubernetesAPI{
				Interface: cs,
			}

			expectedAgentName := ""
			if tc.configMap != nil {
				expectedAgentName = tc.configMap.Data[k8s.AgentNameKey]
			}

			mockApiSrv := bcloudapi.NewMockApiSrv(
				clientID,
				clientSecret,
				accessToken,
				expectedAgentName,
				newAgentID,
			)

			apiAddr := mockApiSrv.Start()
			defer mockApiSrv.Stop()

			apiClient := bcloudapi.New(clientID, clientSecret, apiAddr, false)
			registrator := New(apiClient, k8sApi)
			info, err := registrator.EnsureRegistered(context.Background())
			if err != nil {
				if tc.expErr == nil {
					t.Fatalf("Got unexpected error: %s", err)
				}

				if tc.expErr.Error() != err.Error() {
					t.Fatalf("Expected error: %s, got %s", tc.expErr, err)
				}
			}

			if err := mockApiSrv.GetError(); err != nil {
				t.Fatalf("Got unexpected error: %s", err)
			}

			if !reflect.DeepEqual(tc.expInfo, info) {
				t.Errorf("Expected info: %+v, got %+v", tc.expInfo, info)
			}

			if tc.expectRegistration {
				if mockApiSrv.GetTokenRequests() != 1 {
					t.Errorf("Expected 1 token request, got %d", mockApiSrv.GetTokenRequests())
				}

				if mockApiSrv.GetRegistrationRequests() != 1 {
					t.Errorf("Expected 1 registrationRequests request, got %d", mockApiSrv.GetRegistrationRequests())
				}

				found := false
				for _, action := range cs.Fake.Actions() {
					upd, ok := action.(k8stest.UpdateAction)
					if ok {
						cm, ok := upd.GetObject().(*corev1.ConfigMap)
						if ok {
							found = cm.Name == agentMetadataConfigMap && cm.Namespace == k8s.AgentNamespace
						}
					}
				}
				if !found {
					t.Error("Expected config map to be updated")
				}
			}
		})
	}
}
