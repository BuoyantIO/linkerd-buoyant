package handler

import (
	"time"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/api"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/prototext"
	"gopkg.in/yaml.v2"
)

const (
	linkerdInfoInterval       = time.Minute
	diagnosticCollectDuration = time.Minute * 5
)

type DiagnosticsData struct {
	Logs    string `yaml:"logs"`
	Metrics string `yaml:"metrics"`
}

type DiagnosticsDataRequest struct {
	PodName      string
	Namespace    string
	DiagnosticId string
}

// LinkerdInfo is responsible for obtaining Linkerd related
// data and sending it to the Bcloud API in the form of
// `LinkerdMessage` objects
type LinkerdInfo struct {
	api                      *api.Client
	k8s                      *k8s.Client
	log                      *log.Entry
	proxyDiagnosticsRequests <-chan DiagnosticsDataRequest
	stopCh                   chan struct{}
}

// NewLinkerdInfo instantiates a new k8s event handler.
func NewLinkerdInfo(k8sClient *k8s.Client, apiClient *api.Client, proxyDiagnosticsRequests <-chan DiagnosticsDataRequest) *LinkerdInfo {
	log := log.WithField("handler", "linkerd_info")
	log.Debug("initializing")

	return &LinkerdInfo{
		api:                      apiClient,
		k8s:                      k8sClient,
		log:                      log,
		stopCh:                   make(chan struct{}),
		proxyDiagnosticsRequests: proxyDiagnosticsRequests,
	}
}

// Start initiates linkerd info handler
func (h *LinkerdInfo) Start() {
	ticker := time.NewTicker(linkerdInfoInterval)
	for {
		select {
		case r := <-h.proxyDiagnosticsRequests:
			h.handleProxyDiagnostics(r.PodName, r.Namespace, r.DiagnosticId)
		case <-ticker.C:
			h.handleCertsInfo()
		case <-h.stopCh:
			return
		}
	}
}

// Stop terminates the handler.
func (h *LinkerdInfo) Stop() {
	h.log.Info("shutting down")
	close(h.stopCh)
}

func (h *LinkerdInfo) handleCertsInfo() {
	certs, err := h.k8s.GetControlPlaneCerts()
	if err != nil {
		h.log.Errorf("error getting control plane certs: %s", err)
		return
	}

	m := &pb.CertificateInfo{
		Info: &pb.CertificateInfo_ControlPlane{
			ControlPlane: certs,
		},
	}
	h.log.Tracef("handleLinkerdInfo: %s", prototext.Format(m))

	err = h.api.CrtInfo(m)
	if err != nil {
		h.log.Errorf("error sending CertificateInfo message: %s", err)
	}
}

func (h *LinkerdInfo) handleProxyDiagnostics(podName string, namespace string, diagnosticsId string) {
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

		data, err := yaml.Marshal(DiagnosticsData{Logs: logs, Metrics: metrics})
		if err != nil {
			h.log.Errorf("cannot marshal diagnostic data for pod %s/%s: %s", namespace, podName, err)
			return
		}
		m := &pb.ProxyDiagnostics{
			DiagnosticId: diagnosticsId,
			Data:         data,
		}
		h.log.Tracef("handleLinkerdInfo: %s", prototext.Format(m))

		err = h.api.ProxyDiagnostics(m)
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
