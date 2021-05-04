package handler

import (
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/api"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// Event listens to the k8s API for events, and forwards them to the Buoyant
// Cloud API.
type Event struct {
	api *api.Client
	k8s *k8s.Client
	log *log.Entry
}

// NewEvent instantiates a new k8s event handler.
func NewEvent(k8sClient *k8s.Client, apiClient *api.Client) *Event {
	log := log.WithField("handler", "event")
	log.Debug("initializing")

	return &Event{
		api: apiClient,
		k8s: k8sClient,
		log: log,
	}
}

// Start initiates listening to a k8s event handler.
func (h *Event) Start(sharedInformers informers.SharedInformerFactory) {
	sharedInformers.Core().V1().Events().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				event := obj.(*v1.Event)
				h.handleEvent(event)
			},
		},
	)
}

func (h *Event) handleEvent(event *v1.Event) error {
	h.log.Tracef("handleEvent: %+v\n", event)

	e, err := h.k8s.EventToPB(event)
	if err != nil {
		h.log.Errorf("Failed to create event: %s", err)
		return err
	}

	if e == nil {
		h.log.Tracef("non-workload event, skipping: [%+v]", event)
		return nil
	}

	err = h.api.AddEvent(e)
	if err != nil {
		h.log.Errorf("Failed to send event: %s", err)
		return err
	}

	return nil
}
