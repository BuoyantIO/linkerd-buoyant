package bcloudapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
)

type token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// MockAPIServer is used to mock the Bcloud client API
type MockAPIServer struct {
	tokenRequests        int
	registrationRequests int
	sync.Mutex
	err                 error
	expectedClientID    string
	expectedSecret      string
	expectedAccessToken string
	expectedName        string
	expectedNewAgentID  string
	srv                 *httptest.Server
}

// NewMockApiSrv creates a new MockAPIServer.
func NewMockApiSrv(expectedClientID, expectedSecret, expectedAccessToken, expectedName, expectedNewAgentID string) *MockAPIServer {
	return &MockAPIServer{
		expectedClientID:    expectedClientID,
		expectedSecret:      expectedSecret,
		expectedAccessToken: expectedAccessToken,
		expectedName:        expectedName,
		expectedNewAgentID:  expectedNewAgentID,
	}
}

// Start starts the mock server and returns its address.
func (mas *MockAPIServer) Start() string {
	mas.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mas.Lock()
		defer mas.Unlock()
		if r.URL.String() == tokenEndpoint {
			id, secret, _ := r.BasicAuth()
			if id != mas.expectedClientID {
				mas.err = fmt.Errorf("wrong client id: %s", id)
				return
			}

			if secret != mas.expectedSecret {
				mas.err = fmt.Errorf("wrong client secret: %s", secret)
				return
			}

			data, _ := json.Marshal(&token{AccessToken: mas.expectedAccessToken, ExpiresIn: 3600})
			mas.tokenRequests += 1
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
			return
		} else if strings.Contains(r.URL.String(), registerAgentEndpoint) {
			auth := r.Header.Get("Authorization")
			expectedAuth := fmt.Sprintf("Bearer %s", mas.expectedAccessToken)
			if auth != expectedAuth {
				mas.err = fmt.Errorf("expected Authorization header: %s but got: %s", expectedAuth, auth)
				return
			}

			expectedName := mas.expectedName
			actualName := r.URL.Query().Get(k8s.AgentNameKey)

			if expectedName != actualName {
				mas.err = fmt.Errorf("expected to register agent with name: %s but got name: %s", expectedName, actualName)
				return
			}

			data, _ := json.Marshal(&AgentInfo{AgentName: actualName, AgentID: mas.expectedNewAgentID, IsNewAgent: true})
			mas.registrationRequests += 1
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
			return
		} else {
			mas.err = fmt.Errorf("unexpected request: %s", r.URL.String())
		}
	}))

	return mas.srv.Listener.Addr().String()
}

// Stop stops the server
func (mas *MockAPIServer) Stop() {
	mas.srv.Close()
}

// GetTokenRequests returns the number of token requests that the server has received.
func (mas *MockAPIServer) GetTokenRequests() int {
	mas.Lock()
	defer mas.Unlock()
	return mas.tokenRequests
}

// GetRegistrationRequests returns the number of agent registration requests that the server has received.
func (mas *MockAPIServer) GetRegistrationRequests() int {
	mas.Lock()
	defer mas.Unlock()
	return mas.registrationRequests
}

// GetError returns any error that the server has encountered while processing requests.
func (mas *MockAPIServer) GetError() error {
	mas.Lock()
	defer mas.Unlock()
	return mas.err
}
