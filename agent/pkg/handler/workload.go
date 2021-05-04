package handler

import (
	"time"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/api"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/types/known/timestamppb"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

const (
	workloadsRefreshInterval = 10 * time.Minute
)

// Workload listens to the k8s API for DaemonSet, Deployment, and StatefulSet
// changes, and forwards them to the Buoyant Cloud API.
type Workload struct {
	log    *log.Entry
	api    *api.Client
	k8s    *k8s.Client
	stopCh chan struct{}
}

// NewWorkload instantiates a new k8s workload handler.
func NewWorkload(k8sClient *k8s.Client, apiClient *api.Client) *Workload {
	log := log.WithField("handler", "workload")
	log.Debug("initializing")

	handler := &Workload{
		log:    log,
		api:    apiClient,
		k8s:    k8sClient,
		stopCh: make(chan struct{}),
	}

	return handler
}

//
// Workload
//

// Start begins a polling loop, periodically resyncing all k8s objects with the
// Buoyant Cloud API. This syncing operation runs independently from the k8s
// handlers, but messages are syncronized via the WorkloadStream.
func (h *Workload) Start(sharedInformers informers.SharedInformerFactory) {
	sharedInformers.Apps().V1().Deployments().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    h.handleDeployAdd,
			UpdateFunc: h.handleDeployUpdate,
			DeleteFunc: h.handleDeployDelete,
		},
	)
	sharedInformers.Apps().V1().DaemonSets().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    h.handleDSAdd,
			UpdateFunc: h.handleDSUpdate,
			DeleteFunc: h.handleDSDelete,
		},
	)
	sharedInformers.Apps().V1().StatefulSets().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    h.handleSTSAdd,
			UpdateFunc: h.handleSTSUpdate,
			DeleteFunc: h.handleSTSDelete,
		},
	)

	ticker := time.NewTicker(workloadsRefreshInterval)
	for {
		select {
		case <-ticker.C:
			h.handleWorkloadList()
		case <-h.stopCh:
			return
		}
	}
}

// Stop terminates the WorkloadStream resync loop.
func (h *Workload) Stop() {
	h.log.Info("shutting down")
	close(h.stopCh)
}

//
// Deployments
//

func (h *Workload) handleDeployAdd(obj interface{}) {
	deploy := obj.(*appsv1.Deployment)
	h.log.Debugf("adding Deployment %s/%s", deploy.Namespace, deploy.Name)
	h.handleAdd(h.k8s.DeployToWorkload(deploy))
}

func (h *Workload) handleDeployUpdate(oldObj, newObj interface{}) {
	oldDeploy := oldObj.(*appsv1.Deployment)
	newDeploy := newObj.(*appsv1.Deployment)
	h.log.Debugf("updating Deployment %s/%s", newDeploy.Namespace, newDeploy.Name)
	h.handleUpdate(h.k8s.DeployToWorkload(oldDeploy), h.k8s.DeployToWorkload(newDeploy))
}

func (h *Workload) handleDeployDelete(obj interface{}) {
	deploy, ok := obj.(*appsv1.Deployment)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			h.log.Errorf("couldn't get object from tombstone %+v", obj)
			return
		}
		deploy, ok = tombstone.Obj.(*appsv1.Deployment)
		if !ok {
			h.log.Errorf("tombstone contained object that is not a Deployment %+v", obj)
			return
		}
	}

	h.log.Debugf("deleting Deployment %s/%s", deploy.Namespace, deploy.Name)

	h.handleDelete(h.k8s.DeployToWorkload(deploy))
}

//
// DaemonSets
//

func (h *Workload) handleDSAdd(obj interface{}) {
	ds := obj.(*appsv1.DaemonSet)
	h.log.Debugf("adding DaemonSet %s/%s", ds.Namespace, ds.Name)
	h.handleAdd(h.k8s.DSToWorkload(ds))
}

func (h *Workload) handleDSUpdate(oldObj, newObj interface{}) {
	oldDS := oldObj.(*appsv1.DaemonSet)
	newDS := newObj.(*appsv1.DaemonSet)
	h.log.Debugf("updating DaemonSet %s/%s", newDS.Namespace, newDS.Name)
	h.handleUpdate(h.k8s.DSToWorkload(oldDS), h.k8s.DSToWorkload(newDS))
}

func (h *Workload) handleDSDelete(obj interface{}) {
	ds, ok := obj.(*appsv1.DaemonSet)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			h.log.Errorf("couldn't get object from tombstone %+v", obj)
			return
		}
		ds, ok = tombstone.Obj.(*appsv1.DaemonSet)
		if !ok {
			h.log.Errorf("tombstone contained object that is not a daemonset %+v", obj)
			return
		}
	}

	h.log.Debugf("deleting DaemonSet %s/%s", ds.Namespace, ds.Name)

	h.handleDelete(h.k8s.DSToWorkload(ds))
}

//
// StatefulSets
//

func (h *Workload) handleSTSAdd(obj interface{}) {
	ds := obj.(*appsv1.StatefulSet)
	h.log.Debugf("adding StatefulSet %s/%s", ds.Namespace, ds.Name)
	h.handleAdd(h.k8s.STSToWorkload(ds))
}

func (h *Workload) handleSTSUpdate(oldObj, newObj interface{}) {
	oldSTS := oldObj.(*appsv1.StatefulSet)
	newSTS := newObj.(*appsv1.StatefulSet)
	h.log.Debugf("updating StatefulSet %s/%s", newSTS.Namespace, newSTS.Name)
	h.handleUpdate(h.k8s.STSToWorkload(oldSTS), h.k8s.STSToWorkload(newSTS))
}

func (h *Workload) handleSTSDelete(obj interface{}) {
	sts, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			h.log.Errorf("couldn't get object from tombstone %+v", obj)
			return
		}
		sts, ok = tombstone.Obj.(*appsv1.StatefulSet)
		if !ok {
			h.log.Errorf("tombstone contained object that is not a StatefulSet %+v", obj)
			return
		}
	}

	h.log.Debugf("deleting StatefulSet %s/%s", sts.Namespace, sts.Name)

	h.handleDelete(h.k8s.STSToWorkload(sts))
}

//
// Generic workload handlers
//

func (h *Workload) handleAdd(workload *pb.Workload) {
	m := &pb.WorkloadMessage{
		Message: &pb.WorkloadMessage_Added{
			Added: &pb.AddWorkload{Workload: workload},
		},
	}
	h.log.Tracef("handleAdd: %s", prototext.Format(m))

	err := h.api.WorkloadStream(m)
	if err != nil {
		h.log.Errorf("error sending add message: %s", err)
	}
}

func (h *Workload) handleUpdate(oldWorkload *pb.Workload, newWorkload *pb.Workload) {
	now := time.Now()
	m := &pb.WorkloadMessage{
		Message: &pb.WorkloadMessage_Updated{
			Updated: &pb.UpdateWorkload{
				OldWorkload: oldWorkload,
				NewWorkload: newWorkload,
				Timestamp: &timestamppb.Timestamp{
					Seconds: now.Unix(),
					Nanos:   int32(now.Nanosecond()),
				},
			},
		},
	}
	h.log.Tracef("handleUpdate: %s", prototext.Format(m))

	err := h.api.WorkloadStream(m)
	if err != nil {
		h.log.Errorf("error sending update message: %s", err)
	}
}

func (h *Workload) handleDelete(workload *pb.Workload) {
	m := &pb.WorkloadMessage{
		Message: &pb.WorkloadMessage_Deleted{
			Deleted: &pb.DeleteWorkload{Workload: workload},
		},
	}

	h.log.Tracef("handleDelete: %s", prototext.Format(m))

	err := h.api.WorkloadStream(m)
	if err != nil {
		h.log.Errorf("error sending delete message: %s", err)
	}
}

func (h *Workload) handleWorkloadList() {
	workloads, err := h.k8s.ListWorkloads()
	if err != nil {
		h.log.Errorf("error listing all workloads: %s", err)
		return
	}

	m := &pb.WorkloadMessage{
		Message: &pb.WorkloadMessage_List{
			List: &pb.ListWorkloads{Workloads: workloads},
		},
	}
	h.log.Tracef("handleWorkloadList: %s", prototext.Format(m))

	err = h.api.WorkloadStream(m)
	if err != nil {
		h.log.Errorf("error sending list message: %s", err)
	}
}
