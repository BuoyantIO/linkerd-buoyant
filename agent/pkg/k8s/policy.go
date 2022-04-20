package k8s

import (
	"context"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	policy "github.com/linkerd/linkerd2/controller/gen/apis/policy/v1alpha1"
	server "github.com/linkerd/linkerd2/controller/gen/apis/server/v1beta1"
	serverAuthorization "github.com/linkerd/linkerd2/controller/gen/apis/serverauthorization/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	sazGVR         = serverAuthorization.SchemeGroupVersion.WithResource("serverauthorizations")
	serverGVR      = server.SchemeGroupVersion.WithResource("servers")
	authPolicyGVR  = policy.SchemeGroupVersion.WithResource("authorizationpolicies")
	meshTLSAuthGVR = policy.SchemeGroupVersion.WithResource("meshtlsauthentications")
	netwrokAuthGVR = policy.SchemeGroupVersion.WithResource("networkauthentications")
)

func (c *Client) GetServers(ctx context.Context) ([]*pb.Server, error) {
	supported, err := c.resourceSupported(serverGVR)
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, nil
	}

	servers, err := c.l5dClient.ServerV1beta1().Servers(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	results := make([]*pb.Server, len(servers.Items))
	for i, s := range servers.Items {
		s := s
		results[i] = &pb.Server{
			Server: c.serialize(&s, server.SchemeGroupVersion),
		}
	}

	return results, nil
}

func (c *Client) GetServerAuths(ctx context.Context) ([]*pb.ServerAuthorization, error) {
	supported, err := c.resourceSupported(sazGVR)
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, nil
	}

	serverAuths, err := c.l5dClient.ServerauthorizationV1beta1().ServerAuthorizations(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	results := make([]*pb.ServerAuthorization, len(serverAuths.Items))
	for i, s := range serverAuths.Items {
		s := s
		results[i] = &pb.ServerAuthorization{
			ServerAuthorization: c.serialize(&s, serverAuthorization.SchemeGroupVersion),
		}
	}

	return results, nil
}

func (c *Client) GetAuthorizationPolicies(ctx context.Context) ([]*pb.AuthorizationPolicy, error) {
	supported, err := c.resourceSupported(authPolicyGVR)
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, nil
	}

	policies, err := c.l5dClient.PolicyV1alpha1().AuthorizationPolicies(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	results := make([]*pb.AuthorizationPolicy, len(policies.Items))
	for i, s := range policies.Items {
		s := s
		results[i] = &pb.AuthorizationPolicy{
			AuthorizationPolicy: c.serialize(&s, policy.SchemeGroupVersion),
		}
	}

	return results, nil
}

func (c *Client) GetMeshTLSAuthentications(ctx context.Context) ([]*pb.MeshTLSAuthentication, error) {
	supported, err := c.resourceSupported(meshTLSAuthGVR)
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, nil
	}

	meshAuths, err := c.l5dClient.PolicyV1alpha1().MeshTLSAuthentications(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	results := make([]*pb.MeshTLSAuthentication, len(meshAuths.Items))
	for i, s := range meshAuths.Items {
		s := s
		results[i] = &pb.MeshTLSAuthentication{
			MeshTlsAuthentication: c.serialize(&s, policy.SchemeGroupVersion),
		}
	}

	return results, nil
}

func (c *Client) GetNetworkAuthentications(ctx context.Context) ([]*pb.NetworkAuthentication, error) {
	supported, err := c.resourceSupported(netwrokAuthGVR)
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, nil
	}

	networkAuths, err := c.l5dClient.PolicyV1alpha1().NetworkAuthentications(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	results := make([]*pb.NetworkAuthentication, len(networkAuths.Items))
	for i, s := range networkAuths.Items {
		s := s
		results[i] = &pb.NetworkAuthentication{
			NetworkAuthentication: c.serialize(&s, policy.SchemeGroupVersion),
		}
	}

	return results, nil
}
