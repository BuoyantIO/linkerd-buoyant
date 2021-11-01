package k8s

import (
	"bytes"
	"context"
	"testing"

	ts "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
)

func TestGetTrafficSplits(t *testing.T) {
	weight := resource.MustParse("500m")
	trafficSplit := &ts.TrafficSplit{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ts.SchemeGroupVersion.Identifier(),
			Kind:       "TrafficSplit",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "split",
			Namespace: "ns",
		},
		Spec: ts.TrafficSplitSpec{
			Service: "foo",
			Backends: []ts.TrafficSplitBackend{
				{
					Service: "foo-v1",
					Weight:  &weight,
				},
				{
					Service: "foo-v2",
					Weight:  &weight,
				},
			},
		},
	}

	client := fakeClient(trafficSplit)

	result, err := client.GetTrafficSplits(context.Background())
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	jsonSerializer := scheme.DefaultJSONEncoder()
	if err := jsonSerializer.Encode(trafficSplit, &buf); err != nil {
		t.Error(err)
	}

	expected := buf.String()
	actual := string(result[0].TrafficSplit)

	if expected != actual {
		t.Fatalf("exepected %s, got %s", expected, actual)
	}
}
