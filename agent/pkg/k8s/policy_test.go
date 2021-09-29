package k8s

import (
	"bytes"
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kubectl/pkg/scheme"
)

func TestGetServers(t *testing.T) {
	server := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "policy.linkerd.io/v1alpha1",
			"kind":       "Server",
			"metadata": map[string]interface{}{
				"name":      "srv",
				"namespace": "srvns",
			},
			"spec": map[string]interface{}{
				"port":          "http",
				"proxyProtocol": "HTTP/1",
			},
		},
	}

	client := fakeClient(server)

	result, err := client.GetServers(context.Background())
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	jsonSerializer := scheme.DefaultJSONEncoder()
	if err := jsonSerializer.Encode(server, &buf); err != nil {
		t.Error(err)
	}

	expected := buf.String()
	actual := string(result[0].Server)

	if expected != actual {
		t.Fatalf("exepected %s, got %s", expected, actual)
	}
}

func TestGetServerAuths(t *testing.T) {
	server := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "policy.linkerd.io/v1alpha1",
			"kind":       "ServerAuthorization",
			"metadata": map[string]interface{}{
				"name":      "saz",
				"namespace": "sazns",
			},
			"spec": map[string]interface{}{
				"server": map[string]interface{}{
					"name": "web-http",
				},
				"client": map[string]interface{}{
					"unauthenticated": "true",
				},
			},
		},
	}

	client := fakeClient(server)

	result, err := client.GetServerAuths(context.Background())
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	jsonSerializer := scheme.DefaultJSONEncoder()
	if err := jsonSerializer.Encode(server, &buf); err != nil {
		t.Error(err)
	}

	expected := buf.String()
	actual := string(result[0].ServerAuthorization)

	if expected != actual {
		t.Fatalf("exepected %s, got %s", expected, actual)
	}
}
