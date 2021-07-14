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
	workloadStream    *workloadStream
	manageAgentStream *manageAgentStream

	log *log.Entry

	// protects stream
	sync.Mutex
}

// NewClient instantiates a new Buoyant Cloud API client.
func NewClient(id string, key string, client pb.ApiClient) *Client {
	auth := &pb.Auth{
		AgentId:  id,
		AgentKey: key,
	}
	return &Client{
		auth:              auth,
		workloadStream:    newWorkloadStream(auth, client),
		manageAgentStream: newManageAgentStream(auth, client),
		client:            client,
		log:               log.WithField("api", id),
	}
}

func (c *Client) Init() {
	go c.manageAgentStream.startStream()
}
