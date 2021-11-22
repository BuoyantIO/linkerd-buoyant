package k8s

import (
	"bytes"
	"context"
	"testing"

	link "github.com/linkerd/linkerd2/controller/gen/apis/link/v1alpha1"
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/kubectl/pkg/scheme"
)

func TestGetMcLinks(t *testing.T) {
	mcLink := link.Link{
		TypeMeta: metav1.TypeMeta{
			APIVersion: link.SchemeGroupVersion.Identifier(),
			Kind:       l5dk8s.LinkKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "linkname",
			Namespace: "linkns",
		},

		Spec: link.LinkSpec{
			TargetClusterName:             "tcn",
			TargetClusterDomain:           "tcd",
			TargetClusterLinkerdNamespace: "tcln",
			ClusterCredentialsSecret:      "ccs",
			GatewayAddress:                "ga",
			GatewayPort:                   "555",
			GatewayIdentity:               "identity",
			ProbeSpec: link.ProbeSpec{
				Path:   "pth",
				Port:   "80",
				Period: "8s",
			},
			Selector: *metav1.SetAsLabelSelector(labels.Set(map[string]string{"l": "v"})),
		},
	}

	client := fakeClient(&mcLink)

	result, err := client.GetMulticlusterLinks(context.Background())
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	jsonSerializer := scheme.DefaultJSONEncoder()
	if err := jsonSerializer.Encode(&mcLink, &buf); err != nil {
		t.Error(err)
	}

	expected := buf.String()
	actual := string(result[0].MulticlusterLink)

	if expected != actual {
		t.Fatalf("exepected %s, got %s", expected, actual)
	}
}
