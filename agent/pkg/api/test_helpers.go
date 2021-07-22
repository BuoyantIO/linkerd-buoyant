package api

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	fakeID  = "fake-id"
	fakeKey = "fake-key"
)

type MockAPI_ManageAgentClient struct {
	agentCommandMessages []*pb.AgentCommand
	grpc.ClientStream

	// protects state
	sync.Mutex
}

func (c *MockAPI_ManageAgentClient) Recv() (*pb.AgentCommand, error) {
	c.Lock()
	defer c.Unlock()

	if len(c.agentCommandMessages) == 0 {
		return nil, fmt.Errorf("no more messages")
	}

	command := c.agentCommandMessages[0]
	if len(c.agentCommandMessages) > 1 {
		c.agentCommandMessages = c.agentCommandMessages[1:]
	} else {
		c.agentCommandMessages = make([]*pb.AgentCommand, 0)
	}

	return command, nil
}

func (c *MockAPI_ManageAgentClient) CloseSend() error {
	return nil // noop
}

// MockBcloudClient satisfies the bcloud.ApiClient and
// bcloud.Api_WorkloadStreamClient interfaces, and saves all params and messages
// passed to it.
type MockBcloudClient struct {
	// input
	err error

	// output
	id                      string
	key                     string
	messages                []*pb.WorkloadMessage
	events                  []*pb.Event
	linkerdMessages         []*pb.LinkerdMessage
	proxyDiagnosticMessages []*pb.ProxyDiagnostic
	proxyDiagnosticAuth     *pb.Auth

	// simulates commands received from bcloud-api
	agentCommandMessages []*pb.AgentCommand

	// protects messages and events
	sync.Mutex
}

//
// MockBcloudClient methods
//

func (m *MockBcloudClient) Events() []*pb.Event {
	m.Lock()
	defer m.Unlock()
	return m.events
}

func (m *MockBcloudClient) Messages() []*pb.WorkloadMessage {
	m.Lock()
	defer m.Unlock()
	return m.messages
}

func (m *MockBcloudClient) LinkerdMessages() []*pb.LinkerdMessage {
	m.Lock()
	defer m.Unlock()
	return m.linkerdMessages
}

func (m *MockBcloudClient) ProxyDiagnosticMessages() []*pb.ProxyDiagnostic {
	m.Lock()
	defer m.Unlock()
	return m.proxyDiagnosticMessages
}

func (m *MockBcloudClient) ProxyDiagnosticsAuth() *pb.Auth {
	m.Lock()
	defer m.Unlock()
	return m.proxyDiagnosticAuth
}

//
// bcloud.ApiClient methods
//

func (m *MockBcloudClient) WorkloadStream(
	ctx context.Context, _ ...grpc.CallOption,
) (pb.Api_WorkloadStreamClient, error) {
	return m, m.err
}

func (m *MockBcloudClient) AddEvent(
	ctx context.Context, event *pb.Event, _ ...grpc.CallOption,
) (*pb.Empty, error) {
	m.Lock()
	defer m.Unlock()

	m.id = event.GetAuth().GetAgentId()
	m.key = event.GetAuth().GetAgentKey()
	m.events = append(m.events, event)
	return nil, m.err
}

func (m *MockBcloudClient) LinkerdInfo(
	ctx context.Context, message *pb.LinkerdMessage, _ ...grpc.CallOption,
) (*pb.Empty, error) {
	m.Lock()
	defer m.Unlock()

	m.id = message.GetAuth().GetAgentId()
	m.key = message.GetAuth().GetAgentKey()
	m.linkerdMessages = append(m.linkerdMessages, message)
	return nil, m.err
}

func (m *MockBcloudClient) ManageAgent(
	ctx context.Context, auth *pb.Auth, opts ...grpc.CallOption) (pb.Api_ManageAgentClient, error) {
	m.Lock()
	defer m.Unlock()
	m.proxyDiagnosticAuth = auth
	stream := &MockAPI_ManageAgentClient{
		agentCommandMessages: m.agentCommandMessages,
	}
	return stream, nil
}

func (m *MockBcloudClient) ProxyDiagnostics(ctx context.Context, message *pb.ProxyDiagnostic, opts ...grpc.CallOption) (*pb.Empty, error) {
	m.Lock()
	defer m.Unlock()

	m.id = message.GetAuth().GetAgentId()
	m.key = message.GetAuth().GetAgentKey()
	m.proxyDiagnosticMessages = append(m.proxyDiagnosticMessages, message)
	return nil, m.err
}

//
// bcloud.Api_WorkloadStreamClient methods
//

func (m *MockBcloudClient) Send(msg *pb.WorkloadMessage) error {
	m.Lock()
	defer m.Unlock()

	_, ok := msg.Message.(*pb.WorkloadMessage_Auth)
	if ok {
		m.id = msg.GetAuth().GetAgentId()
		m.key = msg.GetAuth().GetAgentKey()
	}

	m.messages = append(m.messages, msg)
	return nil
}

func (m *MockBcloudClient) CloseAndRecv() (*pb.Empty, error) {
	return nil, nil
}

//
// grpc.ClientStream methods
//

func (m *MockBcloudClient) Header() (metadata.MD, error) {
	return nil, nil
}
func (m *MockBcloudClient) Trailer() metadata.MD {
	return nil
}
func (m *MockBcloudClient) CloseSend() error {
	return nil
}
func (m *MockBcloudClient) Context() context.Context {
	return nil
}
func (m *MockBcloudClient) SendMsg(_ interface{}) error {
	return nil
}
func (m *MockBcloudClient) RecvMsg(_ interface{}) error {
	return nil
}
