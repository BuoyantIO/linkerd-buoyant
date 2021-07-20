package handler

import (
	"context"

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
	h.log.Info("Starting manage agent handler")
	for {
		select {
		case <-h.stopCh:
			return
		case agentCommand := <-h.api.AgentCommands():
			switch command := agentCommand.Command.(type) {
			case *pb.AgentCommand_GetProxyDiagnostics:
				proxyDiagnostic := command.GetProxyDiagnostics
				go h.handleProxyDiagnostics(context.Background(), proxyDiagnostic.PodName, proxyDiagnostic.PodNamespace, proxyDiagnostic.DiagnosticId)
			}
		}
	}
}

// Stop terminates the handler.
func (h *ManageAgent) Stop() {
	h.log.Info("shutting down")
	close(h.stopCh)
}

func (h *ManageAgent) handleProxyDiagnostics(ctx context.Context, podName, namespace, diagnosticID string) {
	logs, err := h.k8s.GetProxyLogs(ctx, podName, namespace)
	if err != nil {
		h.log.Errorf("cannot obtain proxy logs for diagnosticID %s: %s", diagnosticID, err)
	} else {
		h.log.Debugf("Successfully obtained proxy logs for diagnosticID: %s", diagnosticID)
	}

	podSpec, err := h.k8s.GetPodSpec(ctx, podName, namespace)
	if err != nil {
		h.log.Errorf("cannot obtain pod manifest for diagnosticID %s: %s", diagnosticID, err)
	} else {
		h.log.Debugf("Successfully obtained pod manifest for diagnosticID: %s", diagnosticID)
	}

	metrics, err := h.k8s.GetPrometheusScrape(ctx, podName, namespace)
	if err != nil {
		h.log.Errorf("cannot obtain proxy metrics for diagnosticID %s: %s", diagnosticID, err)
	} else {
		h.log.Debugf("Successfully obtained proxy metrics for diagnosticID: %s", diagnosticID)
	}

	ld5ConfigMap, err := h.k8s.GetLinkerdConfigMap(ctx)
	if err != nil {
		h.log.Errorf("cannot Linkerd config map for diagnosticID %s: %s", diagnosticID, err)
	} else {
		h.log.Debugf("Successfully obtained Linkerd config map for for diagnosticID: %s", diagnosticID)
	}

	nodes, err := h.k8s.GetNodeManifests(ctx)
	if err != nil {
		h.log.Errorf("cannot obtain nodes for diagnosticID %s: %s", diagnosticID, err)
	} else {
		h.log.Debugf("Successfully obtained nodes for diagnosticID: %s", diagnosticID)
	}

	svcManifest, err := h.k8s.GetK8sServiceManifest(ctx)
	if err != nil {
		h.log.Errorf("cannot obtain K8s svc manifest for diagnosticID %s: %s", diagnosticID, err)
	} else {
		h.log.Debugf("Successfully obtained K8s svc manifest for diagnosticID: %s", diagnosticID)
	}

	err = h.api.ProxyDiagnostics(diagnosticID, logs, metrics, podSpec, ld5ConfigMap, nodes, svcManifest)
	if err != nil {
		h.log.Errorf("error sending ProxyDiagnostics message: %s", err)
	}
}
