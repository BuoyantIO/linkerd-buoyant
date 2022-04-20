package k8s

import (
	"bytes"
	"context"
	"testing"

	policy "github.com/linkerd/linkerd2/controller/gen/apis/policy/v1alpha1"
	server "github.com/linkerd/linkerd2/controller/gen/apis/server/v1beta1"
	serverauthorization "github.com/linkerd/linkerd2/controller/gen/apis/serverauthorization/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kubectl/pkg/scheme"
	gatewayapiv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
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

var ns = gatewayapiv1alpha2.Namespace("ns")
var refs = []gatewayapiv1alpha2.PolicyTargetReference{
	{
		Group:     "group-2",
		Kind:      "kind-2",
		Name:      "name-2",
		Namespace: &ns,
	},
	{
		Group:     "group-3",
		Kind:      "kind-3",
		Name:      "name-3",
		Namespace: &ns,
	},
	{
		Group:     "group-4",
		Kind:      "kind-4",
		Name:      "name-4",
		Namespace: &ns,
	},
}

func TestGetAuthorizationPolicies(t *testing.T) {
	ap := &policy.AuthorizationPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: policy.SchemeGroupVersion.Identifier(),
			Kind:       "AuthorizationPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "authpol",
			Namespace: "authpolns",
		},
		Spec: policy.AuthorizationPolicySpec{
			TargetRef: gatewayapiv1alpha2.PolicyTargetReference{
				Group:     "group-1",
				Kind:      "kind-1",
				Name:      "name-1",
				Namespace: &ns,
			},
			RequiredAuthenticationRefs: refs,
		},
	}

	client := fakeClient(ap)

	result, err := client.GetAuthorizationPolicies(context.Background())
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	jsonSerializer := scheme.DefaultJSONEncoder()
	if err := jsonSerializer.Encode(ap, &buf); err != nil {
		t.Error(err)
	}

	expected := buf.String()
	actual := string(result[0].AuthorizationPolicy)

	if expected != actual {
		t.Fatalf("exepected %s, got %s", expected, actual)
	}
}

func TestGetMeshTLSAuthentications(t *testing.T) {
	tls := &policy.MeshTLSAuthentication{
		TypeMeta: metav1.TypeMeta{
			APIVersion: policy.SchemeGroupVersion.Identifier(),
			Kind:       "MeshTLSAuthentication",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tls",
			Namespace: "tlsns",
		},
		Spec: policy.MeshTLSAuthenticationSpec{
			Identities:   []string{"id1", "id2"},
			IdentityRefs: refs,
		},
	}

	client := fakeClient(tls)

	result, err := client.GetMeshTLSAuthentications(context.Background())
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	jsonSerializer := scheme.DefaultJSONEncoder()
	if err := jsonSerializer.Encode(tls, &buf); err != nil {
		t.Error(err)
	}

	expected := buf.String()
	actual := string(result[0].MeshTlsAuthentication)

	if expected != actual {
		t.Fatalf("exepected %s, got %s", expected, actual)
	}
}

func TestGetNetworkAuthentications(t *testing.T) {
	net := &policy.NetworkAuthentication{
		TypeMeta: metav1.TypeMeta{
			APIVersion: policy.SchemeGroupVersion.Identifier(),
			Kind:       "NetworkAuthentication",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "net",
			Namespace: "netns",
		},
		Spec: policy.NetworkAuthenticationSpec{
			Networks: []*policy.Network{
				{
					Cidr:   "cird1",
					Except: []string{"ex-1-1", "ex-1-2"},
				},
				{
					Cidr:   "cird2",
					Except: []string{"ex-2-1", "ex-2-2"},
				},
			},
		},
	}

	client := fakeClient(net)

	results, err := client.GetNetworkAuthentications(context.Background())
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	jsonSerializer := scheme.DefaultJSONEncoder()
	if err := jsonSerializer.Encode(net, &buf); err != nil {
		t.Error(err)
	}

	expected := buf.String()
	actual := string(results[0].NetworkAuthentication)

	if expected != actual {
		t.Fatalf("exepected %s, got %s", expected, actual)
	}
}
