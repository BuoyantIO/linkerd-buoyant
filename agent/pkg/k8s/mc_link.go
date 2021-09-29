package k8s

import (
	"context"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	"github.com/linkerd/linkerd2/pkg/multicluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) GetMulticlusterLinks(ctx context.Context) ([]*pb.MulticlusterLink, error) {
	supported, err := c.resourceSupported(multicluster.LinkGVR)
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, nil
	}

	links, err := c.l5dApi.DynamicClient.Resource(multicluster.LinkGVR).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	results := make([]*pb.MulticlusterLink, len(links.Items))
	for i, s := range links.Items {
		s := s
		results[i] = &pb.MulticlusterLink{
			MulticlusterLink: c.serialize(&s, linkSGV),
		}
	}

	return results, nil
}
