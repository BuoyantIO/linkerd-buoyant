package api

import (
	"context"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	"google.golang.org/protobuf/encoding/prototext"
)

func (c *Client) TrafficSplitInfo(info *pb.TrafficSplitInfo) error {
	msg := &pb.LinkerdMessage{
		Auth: c.auth,
		Message: &pb.LinkerdMessage_TrafficSplitInfo{
			TrafficSplitInfo: info,
		},
	}

	return c.sendLinkerdMsg(msg)
}

func (c *Client) SPInfo(info *pb.ServiceProfileInfo) error {
	msg := &pb.LinkerdMessage{
		Auth: c.auth,
		Message: &pb.LinkerdMessage_ServiceProfileInfo{
			ServiceProfileInfo: info,
		},
	}

	return c.sendLinkerdMsg(msg)
}

func (c *Client) MCInfo(info *pb.MulticlusterInfo) error {
	msg := &pb.LinkerdMessage{
		Auth: c.auth,
		Message: &pb.LinkerdMessage_MulticlusterInfo{
			MulticlusterInfo: info,
		},
	}

	return c.sendLinkerdMsg(msg)
}

func (c *Client) AuthPolicyInfo(info *pb.AuthPolicyInfo) error {
	msg := &pb.LinkerdMessage{
		Auth: c.auth,
		Message: &pb.LinkerdMessage_AuthPolicyInfo{
			AuthPolicyInfo: info,
		},
	}

	return c.sendLinkerdMsg(msg)
}

func (c *Client) CrtInfo(info *pb.CertificateInfo) error {
	msg := &pb.LinkerdMessage{
		Auth: c.auth,
		Message: &pb.LinkerdMessage_CrtInfo{
			CrtInfo: info,
		},
	}

	return c.sendLinkerdMsg(msg)
}

func (c *Client) sendLinkerdMsg(msg *pb.LinkerdMessage) error {
	c.log.Tracef("LinkerdInfo: %s", prototext.Format(msg))
	_, err := c.client.LinkerdInfo(context.Background(), msg)
	return err
}
