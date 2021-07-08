package handler

import (
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/api"
	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	log "github.com/sirupsen/logrus"
)

type ManageAgent struct {
	api                      *api.Client
	log                      *log.Entry
	proxyDiagnosticsRequests chan<- DiagnosticsDataRequest
	stopCh                   chan struct{}
}

// NewManageAgent instantiates a new agent commands handler
func NewManageAgent(apiClient *api.Client, proxyDiagnosticsRequests chan<- DiagnosticsDataRequest) *ManageAgent {
	log := log.WithField("handler", "manage_agent")
	log.Debug("initializing")

	return &ManageAgent{
		api:                      apiClient,
		log:                      log,
		stopCh:                   make(chan struct{}),
		proxyDiagnosticsRequests: proxyDiagnosticsRequests,
	}
}

// Start initiates ManageAgent handler
func (h *ManageAgent) Start() {
	for {
		select {
		case <-h.stopCh:
			return
		default:
			agentCommand, err := h.api.RecvCommand()
			if err != nil {
				h.log.Errorf("error receiving command from bcloud: %s", err)
			} else {
				switch command := agentCommand.Command.(type) {
				case *pb.AgentCommand_GetProxyDiagnostics:
					r := command.GetProxyDiagnostics
					h.proxyDiagnosticsRequests <- DiagnosticsDataRequest{
						PodName:      r.PodName,
						Namespace:    r.PodNamespace,
						DiagnosticId: r.DiagnosticId,
					}
				}
			}
		}
	}
}

// Stop terminates the handler.
func (h *ManageAgent) Stop() {
	h.log.Info("shutting down")
	close(h.stopCh)
}
