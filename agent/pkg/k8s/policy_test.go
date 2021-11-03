package k8s

import (
	"bytes"
	"context"
	"testing"

	server "github.com/linkerd/linkerd2/controller/gen/apis/server/v1beta1"
	serverauthorization "github.com/linkerd/linkerd2/controller/gen/apis/serverauthorization/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kubectl/pkg/scheme"
)

func TestGetServers(t *testing.T) {
	srv := &server.Server{
		TypeMeta: metav1.TypeMeta{
			APIVersion: server.SchemeGroupVersion.Identifier(),
			Kind:       "Server",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "srv",
			Namespace: "srvns",
		},
		Spec: server.ServerSpec{
			Port:          intstr.FromString("http"),
			ProxyProtocol: "HTTP/1",
		},
	}

	client := fakeClient(srv)

	result, err := client.GetServers(context.Background())
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	jsonSerializer := scheme.DefaultJSONEncoder()
	if err := jsonSerializer.Encode(srv, &buf); err != nil {
		t.Error(err)
	}

	expected := buf.String()
	actual := string(result[0].Server)

	if expected != actual {
		t.Fatalf("exepected %s, got %s", expected, actual)
	}
}

func TestGetServerAuths(t *testing.T) {
	srvAuth := &serverauthorization.ServerAuthorization{
		TypeMeta: metav1.TypeMeta{
			APIVersion: serverauthorization.SchemeGroupVersion.Identifier(),
			Kind:       "ServerAuthorization",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "saz",
			Namespace: "sazns",
		},

		Spec: serverauthorization.ServerAuthorizationSpec{
			Server: serverauthorization.Server{
				Name: "web-http",
			},
			Client: serverauthorization.Client{
				Unauthenticated: true,
			},
		},
	}

	client := fakeClient(srvAuth)

	result, err := client.GetServerAuths(context.Background())
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	jsonSerializer := scheme.DefaultJSONEncoder()
	if err := jsonSerializer.Encode(srvAuth, &buf); err != nil {
		t.Error(err)
	}

	expected := buf.String()
	actual := string(result[0].ServerAuthorization)

	if expected != actual {
		t.Fatalf("exepected %s, got %s", expected, actual)
	}
}
