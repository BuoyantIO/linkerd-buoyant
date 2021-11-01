package k8s

import (
	"context"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	sp "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) GetServiceProfiles(ctx context.Context) ([]*pb.ServiceProfile, error) {
	supported, err := c.resourceSupported(sp.SchemeGroupVersion.WithResource("serviceprofiles"))
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, nil
	}

	spses, err := c.l5dClient.LinkerdV1alpha2().ServiceProfiles(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	results := make([]*pb.ServiceProfile, len(spses.Items))
	for i, s := range spses.Items {
		s := s
		results[i] = &pb.ServiceProfile{
			ServiceProfile: c.serialize(&s, sp.SchemeGroupVersion),
		}
	}

	return results, nil
}
