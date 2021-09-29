package k8s

import (
	"bytes"
	"context"
	"testing"

	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kubectl/pkg/scheme"
)

func TestGetMcLinks(t *testing.T) {
	mcLink := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": l5dk8s.LinkAPIGroupVersion,
			"kind":       l5dk8s.LinkKind,
			"metadata": map[string]interface{}{
				"name":      "linkname",
				"namespace": "linkns",
			},
			"spec": map[string]interface{}{
				"targetClusterName":             "tcn",
				"targetClusterDomain":           "tcd",
				"targetClusterLinkerdNamespace": "tcln",
				"clusterCredentialsSecret":      "ccs",
				"gatewayAddress":                "ga",
				"gatewayPort":                   "555",
				"gatewayIdentity":               "identity",
				"probeSpec": map[string]interface{}{
					"path":   "pth",
					"port":   "80",
					"period": "8s",
				},
			},
		},
	}

	client := fakeClient(mcLink)

	result, err := client.GetMulticlusterLinks(context.Background())
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	jsonSerializer := scheme.DefaultJSONEncoder()
	if err := jsonSerializer.Encode(mcLink, &buf); err != nil {
		t.Error(err)
	}

	expected := buf.String()
	actual := string(result[0].MulticlusterLink)

	if expected != actual {
		t.Fatalf("exepected %s, got %s", expected, actual)
	}
}
