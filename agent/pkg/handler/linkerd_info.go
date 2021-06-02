package handler

import (
	"time"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/api"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/prototext"
)

const (
	linkerdInfoInterval = time.Minute
)

// LinkerdInfo is responsible for obtaining Linkerd related
// data and sending it to the Bcloud API in the form of
// `LinkerdMessage` objects
type LinkerdInfo struct {
	api               *api.Client
	k8s               *k8s.Client
	log               *log.Entry
	proxyAddrOverride string
	stopCh            chan struct{}
}

// NewLinkerdInfo instantiates a new k8s event handler.
func NewLinkerdInfo(k8sClient *k8s.Client, apiClient *api.Client, proxyAddrOverride string) *LinkerdInfo {
	log := log.WithField("handler", "linkerd_info")
	log.Debug("initializing")

	return &LinkerdInfo{
		api:               apiClient,
		k8s:               k8sClient,
		log:               log,
		proxyAddrOverride: proxyAddrOverride,
		stopCh:            make(chan struct{}),
	}
}

// Start initiates linkerd info handler
func (h *LinkerdInfo) Start() {
	ticker := time.NewTicker(linkerdInfoInterval)
	for {
		select {
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
	certs, err := h.k8s.GetControlPlaneCerts(h.proxyAddrOverride)
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
