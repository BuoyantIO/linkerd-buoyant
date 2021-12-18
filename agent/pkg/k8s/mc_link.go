package k8s

import (
	"context"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	link "github.com/linkerd/linkerd2/controller/gen/apis/link/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) GetMulticlusterLinks(ctx context.Context) ([]*pb.MulticlusterLink, error) {
	supported, err := c.resourceSupported(link.SchemeGroupVersion.WithResource("links"))
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, nil
	}

	links, err := c.l5dClient.LinkV1alpha1().Links(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	results := make([]*pb.MulticlusterLink, len(links.Items))
	for i, l := range links.Items {
		l := l
		results[i] = &pb.MulticlusterLink{
			MulticlusterLink: c.serialize(&l, link.SchemeGroupVersion),
		}
	}

	return results, nil
}
