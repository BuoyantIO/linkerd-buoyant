package api

import (
	"context"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	"google.golang.org/protobuf/encoding/prototext"
)

// ProxyDiagnostics wraps the Buoyant Cloud API ProxyDiagnostics gRPC unary endpoint.
func (c *Client) ProxyDiagnostics(
	diagnosticID string,
	logs []byte,
	metrics [][]byte,
	podManifest *pb.Pod,
	linkerdConfigMap *pb.ConfigMap,
	nodes []*pb.Node,
	k8sServiceManifest *pb.Service) error {
	diagnostic := &pb.ProxyDiagnostic{
		Auth:               c.auth,
		DiagnosticId:       diagnosticID,
		Logs:               logs,
		Metrics:            metrics,
		PodManifest:        podManifest,
		LinkerdConfigMap:   linkerdConfigMap,
		Nodes:              nodes,
		K8SServiceManifest: k8sServiceManifest,
	}
	c.log.Tracef("ProxyDiagnostics: %s", prototext.Format(diagnostic))

	_, err := c.client.ProxyDiagnostics(context.Background(), diagnostic)
	return err
}
