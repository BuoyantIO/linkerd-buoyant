package api

import (
	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
)

func (c *Client) AgentCommands() <-chan *pb.AgentCommand {
	return c.manageAgentStream.commands
}
