package api

import (
	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
)

// SendWorkloadMessage sends a message via the Buoyant Cloud API WorkloadStream gRPC
// endpoint. It abstracts away the details around managing and protecting
// the client-side stream
func (c *Client) SendWorkloadMessage(msg *pb.WorkloadMessage) error {
	return c.workloadStream.send(msg)
}
