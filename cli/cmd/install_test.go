package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/bcloudapi"
	"github.com/buoyantio/linkerd-buoyant/cli/pkg/k8s"
	"google.golang.org/grpc/credentials"
)

type MockClient struct {
	Identifier       bcloudapi.AgentIdentifier
	ManifestToReturn string
	Err              error
}

func (mc *MockClient) GetAgentManifest(ctx context.Context, identifier bcloudapi.AgentIdentifier) (string, error) {
	mc.Identifier = identifier
	if mc.Err != nil {
		return "", mc.Err
	}
	return mc.ManifestToReturn, nil
}

func (mc *MockClient) RegisterAgent(ctx context.Context, agentName string) (*bcloudapi.AgentInfo, error) {
	return nil, nil
}

func (mc *MockClient) Credentials(ctx context.Context, agentID string) credentials.PerRPCCredentials {
	return nil
}

const fakeAgentName = "fake-agent-name"
const fakeAgentID = "fake-agent-id"
const fakeYaml = "fake-yaml"

func TestInstallNewAgent(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := &installCfg{
		config: &config{
			stdout: stdout,
			stderr: stderr,
		},
		agentName: fakeAgentName,
	}

	client := &k8s.MockClient{}
	apiClient := &MockClient{ManifestToReturn: fakeYaml}

	err := install(context.TODO(), cfg, client, apiClient)
	if err != nil {
		t.Error(err)
	}

	if stdout.String() != fmt.Sprintf("%s\n", fakeYaml) {
		t.Errorf("Expected: [%s], Got: [%s]", fakeYaml, stdout.String())
	}

	name, ok := apiClient.Identifier.(bcloudapi.AgentName)
	if !ok {
		t.Fatalf("Expected to call api with AgentName, called with: %+v", apiClient.Identifier)
	}
	if name.Value() != fakeAgentName {
		t.Fatalf("Expected name identifier to be %s, got: %s", fakeAgentName, name.Value())
	}
}

func TestInstallExistingAgent(t *testing.T) {
	t.Run("gets manifest with name", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		cfg := &installCfg{
			config: &config{
				stdout: stdout,
				stderr: stderr,
			},
		}

		client := &k8s.MockClient{
			MockAgent: &k8s.Agent{
				Name:    fakeAgentName,
				Version: "version",
			},
		}
		apiClient := &MockClient{ManifestToReturn: fakeYaml}

		err := install(context.TODO(), cfg, client, apiClient)
		if err != nil {
			t.Error(err)
		}
		if stdout.String() != fmt.Sprintf("%s\n", fakeYaml) {
			t.Errorf("Expected: [%s], Got: [%s]", fakeYaml, stdout.String())
		}

		name, ok := apiClient.Identifier.(bcloudapi.AgentName)
		if !ok {
			t.Fatalf("Expected to call api with AgentName, called with: %+v", apiClient.Identifier)
		}
		if name.Value() != fakeAgentName {
			t.Fatalf("Expected name identifier to be %s, got: %s", fakeAgentName, name.Value())
		}
	})

	t.Run("gets manifest with Id", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		cfg := &installCfg{
			config: &config{
				stdout: stdout,
				stderr: stderr,
			},
		}

		client := &k8s.MockClient{
			MockAgent: &k8s.Agent{
				Id:      fakeAgentID,
				Version: "version",
			},
		}
		apiClient := &MockClient{ManifestToReturn: fakeYaml}

		err := install(context.TODO(), cfg, client, apiClient)
		if err != nil {
			t.Error(err)
		}
		if stdout.String() != fmt.Sprintf("%s\n", fakeYaml) {
			t.Errorf("Expected: [%s], Got: [%s]", fakeYaml, stdout.String())
		}

		id, ok := apiClient.Identifier.(bcloudapi.AgentID)
		if !ok {
			t.Fatalf("Expected to call api with AgentID, called with: %+v", apiClient.Identifier)
		}
		if id.Value() != fakeAgentID {
			t.Fatalf("Expected name identifier to be %s, got: %s", fakeAgentID, id.Value())
		}
	})

	t.Run("invalid agent data", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		cfg := &installCfg{
			config: &config{
				stdout: stdout,
				stderr: stderr,
			},
		}

		client := &k8s.MockClient{
			MockAgent: &k8s.Agent{
				Version: "version",
			},
		}
		apiClient := &MockClient{ManifestToReturn: fakeYaml}

		expectedError := errors.New("could not find valid agent installation")
		err := install(context.TODO(), cfg, client, apiClient)

		if !reflect.DeepEqual(err, expectedError) {
			t.Errorf("Expected error: %s, Got: %s", expectedError, err)
		}

		expectedStdErr := `Could not find valid agent installation on cluster. To install agent run:
linkerd buoyant install --agent-name=my-new-agent | kubectl apply -f -
`
		if expectedStdErr != stderr.String() {
			t.Errorf("Expected:\n%s\nGot:\n%s", expectedStdErr, stderr.String())
		}
	})
}

func TestInstallError(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := &installCfg{
		config: &config{
			stdout: stdout,
			stderr: stderr,
		},
		agentName: "fake-agent-name",
	}

	expectedError := errors.New("failed to retrieve agent manifest from bcloud server for agent identifier bcloudapi.AgentName fake-agent-name")
	apiError := errors.New("problem with API")
	client := &k8s.MockClient{}
	apiClient := &MockClient{Err: apiError}
	err := install(context.TODO(), cfg, client, apiClient)
	if !reflect.DeepEqual(err, expectedError) {
		t.Errorf("Expected error: %s, Got: %s", expectedError, err)
	}
}
