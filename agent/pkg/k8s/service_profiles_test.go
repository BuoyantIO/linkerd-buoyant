package k8s

import (
	"bytes"
	"context"
	"testing"

	sp "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
)

func TestGetServiceProfile(t *testing.T) {
	serviceProfile := sp.ServiceProfile{
		TypeMeta: metav1.TypeMeta{
			APIVersion: sp.SchemeGroupVersion.Identifier(),
			Kind:       l5dk8s.ServiceProfileKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "spname",
			Namespace: "spns",
		},
		Spec: sp.ServiceProfileSpec{
			Routes: []*sp.RouteSpec{
				{
					Name: "GET /my/path/hi",
					Condition: &sp.RequestMatch{
						PathRegex: `/my/path/hi`,
						Method:    "GET",
					},
				},
				{
					Name: "POST /emojivoto.v1.VotingService/VoteFire",
					Condition: &sp.RequestMatch{
						PathRegex: `/emojivoto\.v1\.VotingService/VoteFire`,
						Method:    "POST",
					},
				},
			},
		},
	}

	client := fakeClient(&serviceProfile)

	result, err := client.GetServiceProfiles(context.Background())
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	jsonSerializer := scheme.DefaultJSONEncoder()
	if err := jsonSerializer.Encode(&serviceProfile, &buf); err != nil {
		t.Error(err)
	}

	expected := buf.String()
	actual := string(result[0].ServiceProfile)

	if expected != actual {
		t.Fatalf("exepected %s, got %s", expected, actual)
	}
}
