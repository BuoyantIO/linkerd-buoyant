package k8s

import (
	"context"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SazGVR is the GroupVersionResource for the ServerAuthorization resource.
var sazGVR = schema.GroupVersionResource{
	Group:    "policy.linkerd.io",
	Version:  "v1alpha1",
	Resource: "serverauthorizations",
}

// ServerGVR is the GroupVersionResource for the Server resource.
var serverGVR = schema.GroupVersionResource{
	Group:    "policy.linkerd.io",
	Version:  "v1alpha1",
	Resource: "servers",
}

func (c *Client) GetServers(ctx context.Context) ([]*pb.Server, error) {
	servers, err := c.k8sDynClient.Resource(serverGVR).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
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
	servers, err := c.k8sDynClient.Resource(sazGVR).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
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
