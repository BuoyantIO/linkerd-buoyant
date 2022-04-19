package k8s

import (
	"context"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	ts "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) GetTrafficSplits(ctx context.Context) ([]*pb.TrafficSplit, error) {
	supported, err := c.resourceSupported(ts.SchemeGroupVersion.WithResource("trafficsplits"))
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, nil
	}

	splits, err := c.tsClient.SplitV1alpha1().TrafficSplits(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	results := make([]*pb.TrafficSplit, len(splits.Items))
	for i, t := range splits.Items {
		t := t
		results[i] = &pb.TrafficSplit{
			TrafficSplit: c.serialize(&t, ts.SchemeGroupVersion),
		}
	}

	return results, nil
}
