package api

import (
	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
)

func (c *Client) RecvAgentCommand() (*pb.AgentCommand, error) {
	return c.manageAgentStream.recv()
}

func (c *Client) CloseAgentCommandStream() {
	c.manageAgentStream.closeStream()
}
