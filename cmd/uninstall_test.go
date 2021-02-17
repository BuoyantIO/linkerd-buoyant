package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/buoyantio/linkerd-buoyant/pkg/k8s"
)

func TestUninstall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cfg := config{
		stdout:       stdout,
		stderr:       stderr,
		bcloudServer: "http://example.com",
	}

	client := &k8s.MockClient{
		MockAgent:     &k8s.Agent{URL: "/fake-url"},
		MockResources: []string{"resource-1", "resource-2", "resource-3"},
	}
	err := uninstall(context.TODO(), cfg, client)
	if err != nil {
		t.Error(err)
	}

	expStdout := "resource-1---\nresource-2---\nresource-3---\n"
	if stdout.String() != expStdout {
		t.Errorf("Expected %s, Got %s", expStdout, stdout.String())
	}
	if !strings.Contains(stderr.String(), "/fake-url") {
		t.Errorf("Expected stderr to contain /fake-url, Got %s", stderr.String())
	}
}
