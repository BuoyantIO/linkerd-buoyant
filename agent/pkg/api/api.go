package api

import (
	"sync"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	log "github.com/sirupsen/logrus"
)

// Client provides a simple Buoyant Cloud API client.
type Client struct {
	auth *pb.Auth

	client            pb.ApiClient
	stream            pb.Api_WorkloadStreamClient
	manageAgentClient pb.Api_ManageAgentClient

	log *log.Entry

	// protects stream
	sync.Mutex
}

// NewClient instantiates a new Buoyant Cloud API client.
func NewClient(id string, key string, client pb.ApiClient) *Client {
	return &Client{
		auth: &pb.Auth{
			AgentId:  id,
			AgentKey: key,
		},
		client: client,
		log:    log.WithField("api", id),
	}
}
