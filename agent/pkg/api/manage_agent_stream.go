package api

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	log "github.com/sirupsen/logrus"
)

var clientClosed = errors.New("Client closed")

// manageAgentStream wraps the Buoyant Cloud API ManageAgent gRPC endpoint, and
// manages the stream.
type manageAgentStream struct {
	auth   *pb.Auth
	client pb.ApiClient
	stream pb.Api_ManageAgentClient
	log    *log.Entry

	// protects stream
	sync.Mutex
	client_closed bool
}

func newManageAgentStream(auth *pb.Auth, client pb.ApiClient) *manageAgentStream {
	return &manageAgentStream{
		auth:   auth,
		client: client,
		log:    log.WithField("stream", "ManageAgentStream"),
	}
}

func (s *manageAgentStream) recv() (*pb.AgentCommand, error) {
	for {
		event, err := s.recv_locked()
		if err == io.EOF {
			s.log.Info("server closed stream, reseting")
			s.resetStream()
			continue
		} else if err != nil {
			if err != clientClosed {
				s.resetStream()
			}
			return nil, err
		}

		return event, nil
	}
}

func (s *manageAgentStream) recv_locked() (*pb.AgentCommand, error) {
	s.Lock()
	defer s.Unlock()
	if s.client_closed {
		return nil, clientClosed
	}
	if s.stream == nil {
		s.stream = s.newStream()
	}

	return s.stream.Recv()
}

func (s *manageAgentStream) newStream() pb.Api_ManageAgentClient {
	var stream pb.Api_ManageAgentClient

	// loop until the request to initiate a stream succeeds
	for {
		var err error
		stream, err = s.client.ManageAgent(context.Background(), s.auth)
		if err != nil {
			s.log.Errorf("failed to initiate stream: %s", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	s.log.Debug("stream opened")
	return stream
}

func (s *manageAgentStream) resetStream() {
	s.Lock()
	defer s.Unlock()
	if s.stream != nil {
		s.stream.CloseSend()
		s.stream = nil
	}
}

func (s *manageAgentStream) closeStream() {
	s.Lock()
	defer s.Unlock()
	s.stream.CloseSend()
	s.stream = nil
	s.client_closed = true
}
