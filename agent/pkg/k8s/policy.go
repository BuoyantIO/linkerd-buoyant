package k8s

import (
	"context"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) GetServers(ctx context.Context) ([]*pb.Server, error) {
	supported, err := c.resourceSupported(l5dk8s.ServerGVR)
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, nil
	}

	servers, err := c.l5dApi.DynamicClient.Resource(l5dk8s.ServerGVR).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	results := make([]*pb.Server, len(servers.Items))
	for i, s := range servers.Items {
		raw, err := s.MarshalJSON()
		if err != nil {
			return nil, err
		}

		results[i] = &pb.Server{
			Server: raw,
		}
	}

	return results, nil
}

func (c *Client) GetServerAuths(ctx context.Context) ([]*pb.ServerAuthorization, error) {
	supported, err := c.resourceSupported(l5dk8s.SazGVR)
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, nil
	}

	servers, err := c.l5dApi.DynamicClient.Resource(l5dk8s.SazGVR).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	results := make([]*pb.ServerAuthorization, len(servers.Items))
	for i, s := range servers.Items {
		raw, err := s.MarshalJSON()
		if err != nil {
			return nil, err
		}

		results[i] = &pb.ServerAuthorization{
			ServerAuthorization: raw,
		}
	}

	return results, nil
}
