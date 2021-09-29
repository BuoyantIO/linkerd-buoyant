package handler

import (
	"context"
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
	api    *api.Client
	k8s    *k8s.Client
	log    *log.Entry
	stopCh chan struct{}
}

// NewLinkerdInfo instantiates a new k8s event handler.
func NewLinkerdInfo(k8sClient *k8s.Client, apiClient *api.Client) *LinkerdInfo {
	log := log.WithField("handler", "linkerd_info")
	log.Debug("initializing")

	return &LinkerdInfo{
		api:    apiClient,
		k8s:    k8sClient,
		log:    log,
		stopCh: make(chan struct{}),
	}
}

// Start initiates linkerd info handler
func (h *LinkerdInfo) Start() {
	ticker := time.NewTicker(linkerdInfoInterval)
	for {
		select {
		case <-ticker.C:
			h.handleCertsInfo(context.Background())
			h.handleAuthPolicyInfo(context.Background())
			h.handleMulticluster(context.Background())
			h.handleServiceProfiles(context.Background())
			h.handleTrafficSplits(context.Background())
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

func (h *LinkerdInfo) handleTrafficSplits(ctx context.Context) {
	trafficSplits, err := h.k8s.GetTrafficSplits(ctx)
	if err != nil {
		h.log.Errorf("error getting traffic splits: %s", err)
		return
	}

	m := &pb.TrafficSplitInfo{
		TrafficSplits: trafficSplits,
	}
	h.log.Tracef("handleTrafficSplits: %s", prototext.Format(m))

	err = h.api.TrafficSplitInfo(m)
	if err != nil {
		h.log.Errorf("error sending TrafficSplitInfo message: %s", err)
	}
}

func (h *LinkerdInfo) handleServiceProfiles(ctx context.Context) {
	serviceProfiles, err := h.k8s.GetServiceProfiles(ctx)
	if err != nil {
		h.log.Errorf("error getting service profiles: %s", err)
		return
	}

	m := &pb.ServiceProfileInfo{
		ServiceProfiles: serviceProfiles,
	}
	h.log.Tracef("handleServiceProfiles: %s", prototext.Format(m))

	err = h.api.SPInfo(m)
	if err != nil {
		h.log.Errorf("error sending ServiceProfileInfo message: %s", err)
	}
}

func (h *LinkerdInfo) handleMulticluster(ctx context.Context) {
	links, err := h.k8s.GetMulticlusterLinks(ctx)
	if err != nil {
		h.log.Errorf("error getting MC links: %s", err)
		return
	}

	m := &pb.MulticlusterInfo{
		MulticlusterLinks: links,
	}
	h.log.Tracef("handleMulticluster: %s", prototext.Format(m))

	err = h.api.MCInfo(m)
	if err != nil {
		h.log.Errorf("error sending MulticlusterInfo message: %s", err)
	}
}

func (h *LinkerdInfo) handleAuthPolicyInfo(ctx context.Context) {
	servers, err := h.k8s.GetServers(ctx)
	if err != nil {
		h.log.Errorf("error getting servers: %s", err)
		return
	}

	serverAuths, err := h.k8s.GetServerAuths(ctx)
	if err != nil {
		h.log.Errorf("error getting server authorizations: %s", err)
		return
	}

	m := &pb.AuthPolicyInfo{
		Servers:              servers,
		ServerAuthorizations: serverAuths,
	}
	h.log.Tracef("handleAuthPolicyInfo: %s", prototext.Format(m))

	err = h.api.AuthPolicyInfo(m)
	if err != nil {
		h.log.Errorf("error sending AuthPolicyInfo message: %s", err)
	}
}

func (h *LinkerdInfo) handleCertsInfo(ctx context.Context) {
	certs, err := h.k8s.GetControlPlaneCerts(ctx)
	if err != nil {
		h.log.Errorf("error getting control plane certs: %s", err)
		return
	}

	m := &pb.CertificateInfo{
		Info: &pb.CertificateInfo_ControlPlane{
			ControlPlane: certs,
		},
	}
	h.log.Tracef("handleCertsInfo: %s", prototext.Format(m))

	err = h.api.CrtInfo(m)
	if err != nil {
		h.log.Errorf("error sending CertificateInfo message: %s", err)
	}
}
