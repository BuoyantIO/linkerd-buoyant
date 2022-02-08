package registrator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	k8stest "k8s.io/client-go/testing"

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

type token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func TestRegistrator(t *testing.T) {
	fixtures := []*struct {
		testName           string
		configMap          *corev1.ConfigMap
		expInfo            *AgentInfo
		expErr             error
		expectRegistration bool
	}{
		{
			"manifest fully hydrated (does not perform registration)",
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      agentMetadataConfigMap,
					Namespace: namespace,
				},
				Data: map[string]string{
					agentIDKey:   "agent-id",
					agentNameKey: "agent-name",
				},
			},
			&AgentInfo{
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
					Namespace: namespace,
				},
				Data: map[string]string{
					agentNameKey: "agent-name",
				},
			},
			&AgentInfo{
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
					Namespace: namespace,
				},
				Data: map[string]string{
					agentIDKey: "agent-id",
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

			tokenRequests := 0
			registrationRequests := 0
			apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.String() == "/token" {
					id, secret, _ := r.BasicAuth()
					if id != clientID {
						t.Fatalf("wrong client id: %s", id)
					}

					if secret != clientSecret {
						t.Fatalf("wrong client secret: %s", secret)
					}

					data, _ := json.Marshal(&token{AccessToken: accessToken, ExpiresIn: 3600})
					tokenRequests += 1
					w.Header().Set("Content-Type", "application/json")
					w.Write(data)
					return
				} else if strings.Contains(r.URL.String(), "register-agent") {
					auth := r.Header.Get("Authorization")
					expectedAuth := fmt.Sprintf("Bearer %s", accessToken)
					if auth != expectedAuth {
						t.Errorf("expected Authorization header: %s but got: %s", expectedAuth, auth)
					}

					expectedName := tc.configMap.Data[agentNameKey]
					actualName := r.URL.Query().Get(agentNameKey)

					if expectedName != actualName {
						t.Errorf("expected to register agent with name: %s but got name: %s", expectedName, actualName)
					}

					data, _ := json.Marshal(&AgentInfo{AgentName: actualName, AgentID: newAgentID, IsNewAgent: true})
					registrationRequests += 1
					w.Header().Set("Content-Type", "application/json")
					w.Write(data)
					return
				} else {
					t.Errorf("Unexpected request: %s", r.URL.String())
				}
			}))
			defer apiSrv.Close()

			registrator := NewAgentRegistrator(clientID, clientSecret, apiSrv.Listener.Addr().String(), false, k8sApi)
			info, err := registrator.EnsureRegistered(context.Background())
			if err != nil {
				if tc.expErr == nil {
					t.Fatalf("Got unexpected error: %s", err)
				}

				if tc.expErr.Error() != err.Error() {
					t.Fatalf("Expected error: %s, got %s", tc.expErr, err)
				}
			}

			if !reflect.DeepEqual(tc.expInfo, info) {
				t.Errorf("Expected info: %+v, got %+v", tc.expInfo, info)
			}

			if tc.expectRegistration {
				if tokenRequests != 1 {
					t.Errorf("Expected 1 token request, got %d", tokenRequests)
				}

				if registrationRequests != 1 {
					t.Errorf("Expected 1 registrationRequests request, got %d", registrationRequests)
				}

				found := false
				for _, action := range cs.Fake.Actions() {
					upd, ok := action.(k8stest.UpdateAction)
					if ok {
						cm, ok := upd.GetObject().(*corev1.ConfigMap)
						if ok {
							found = cm.Name == agentMetadataConfigMap && cm.Namespace == namespace
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
