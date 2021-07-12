package handler

import (
	"time"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/api"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	log "github.com/sirupsen/logrus"
)

type ManageAgent struct {
	api    *api.Client
	k8s    *k8s.Client
	log    *log.Entry
	stopCh chan struct{}
}

// NewManageAgent instantiates a new agent commands handler
func NewManageAgent(k8sClient *k8s.Client, apiClient *api.Client) *ManageAgent {
	log := log.WithField("handler", "manage_agent")
	log.Debug("initializing")

	return &ManageAgent{
		api:    apiClient,
		k8s:    k8sClient,
		log:    log,
		stopCh: make(chan struct{}),
	}
}

// Start initiates ManageAgent handler
func (h *ManageAgent) Start() {
	for {
		select {
		case <-h.stopCh:
			return
		default:
			agentCommand, err := h.api.RecvAgentCommand()
			if err != nil {
				h.log.Errorf("Error from stream %s", err)
				continue
			}
			switch command := agentCommand.Command.(type) {
			case *pb.AgentCommand_GetProxyDiagnostics:
				proxyDiagnostic := command.GetProxyDiagnostics
				go h.handleProxyDiagnostics(proxyDiagnostic.PodName, proxyDiagnostic.PodNamespace, proxyDiagnostic.DiagnosticId)
			}
		}
	}
}

// Stop terminates the handler.
func (h *ManageAgent) Stop() {
	h.log.Info("shutting down")
	close(h.stopCh)
	h.api.CloseAgentCommandStream()
}

func (h *ManageAgent) handleProxyDiagnostics(podName string, namespace string, diagnosticId string) {
	initialLogLevel, err := h.k8s.GetProxyLogLevel(podName, namespace)
	if err != nil {
		h.log.Errorf("error getting proxy log level for pod %s/%s: %s", namespace, podName, err)
		return
	}

	err = h.k8s.SetProxyLogLevel(podName, namespace, "trace")
	if err != nil {
		h.log.Errorf("error setting proxy log level for pod %s/%s: %s", namespace, podName, err)
		return
	}

	time.AfterFunc(diagnosticCollectDuration, func() {
		logs, err := h.k8s.GetProxyLogs(podName, namespace)
		if err != nil {
			h.log.Errorf("cannot obtain proxy logs for pod %s/%s: %s", namespace, podName, err)
			return
		}

		metrics, err := h.k8s.GetPrometheusScrape(podName, namespace)
		if err != nil {
			h.log.Errorf("cannot obtain proxy metrics for pod %s/%s: %s", namespace, podName, err)
			return
		}
		err = h.api.ProxyDiagnostics(diagnosticId, []byte(logs), []byte(metrics))
		if err != nil {
			h.log.Errorf("error sending ProxyDiagnostics message: %s", err)
		}

		err = h.k8s.SetProxyLogLevel(podName, namespace, initialLogLevel)
		if err != nil {
			h.log.Errorf("error reverting proxy log level to %s for pod %s/%s: %s", initialLogLevel, namespace, podName, err)
			return
		}
	})
}
