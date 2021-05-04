package k8s

import (
	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// DSToWorkload serializes a k8s DaemonSet object and wraps it in a bcloud API
// proto message.
func (c *Client) DSToWorkload(ds *appsv1.DaemonSet) *pb.Workload {
	return &pb.Workload{
		Workload: &pb.Workload_Daemonset{
			Daemonset: &pb.DaemonSet{
				DaemonSet: c.serialize(ds, appsv1.SchemeGroupVersion),
				Pods:      c.getPodsFor(ds),
			},
		},
	}
}

// DeployToWorkload serializes a k8s Deployment object and wraps it in a bcloud
// API proto message.
func (c *Client) DeployToWorkload(deploy *appsv1.Deployment) *pb.Workload {
	return &pb.Workload{
		Workload: &pb.Workload_Deployment{
			Deployment: &pb.Deployment{
				Deployment:  c.serialize(deploy, appsv1.SchemeGroupVersion),
				ReplicaSets: c.deployToRS(deploy),
			},
		},
	}
}

// STSToWorkload serializes a k8s StatefulSet object and wraps it in a bcloud
// API proto message.
func (c *Client) STSToWorkload(sts *appsv1.StatefulSet) *pb.Workload {
	return &pb.Workload{
		Workload: &pb.Workload_Statefulset{
			Statefulset: &pb.StatefulSet{
				StatefulSet: c.serialize(sts, appsv1.SchemeGroupVersion),
				Pods:        c.getPodsFor(sts),
			},
		},
	}
}

func (c *Client) ListWorkloads() ([]*pb.Workload, error) {
	workloads := []*pb.Workload{}

	dsList, err := c.dsLister.List(labels.Everything())
	if err != nil {
		c.log.Errorf("error listing all DeamonSets: %s", err)
		return nil, err
	}
	for _, ds := range dsList {
		workloads = append(workloads, c.DSToWorkload(ds))
	}

	deployList, err := c.deployLister.List(labels.Everything())
	if err != nil {
		c.log.Errorf("error listing all deployments: %s", err)
		return nil, err
	}
	for _, deploy := range deployList {
		workloads = append(workloads, c.DeployToWorkload(deploy))
	}

	stsList, err := c.stsLister.List(labels.Everything())
	if err != nil {
		c.log.Errorf("error listing all StatefulSets: %s", err)
		return nil, err
	}
	for _, sts := range stsList {
		workloads = append(workloads, c.STSToWorkload(sts))
	}

	return workloads, nil
}

func (c *Client) deployToRS(deploy *appsv1.Deployment) []*pb.ReplicaSet {
	rsSelector := labels.Set(deploy.Spec.Selector.MatchLabels).AsSelector()
	replicaSets, err := c.rsLister.ReplicaSets(deploy.GetNamespace()).List(rsSelector)
	if err != nil {
		c.log.Errorf("failed to retrieve ReplicaSets for %s/%s", deploy.GetNamespace(), deploy.GetName())
		return nil
	}

	pbReplicaSets := make([]*pb.ReplicaSet, len(replicaSets))
	for i, rs := range replicaSets {
		pbReplicaSets[i] = &pb.ReplicaSet{
			ReplicaSet: c.serialize(rs, appsv1.SchemeGroupVersion),
			Pods:       c.getPodsFor(rs),
		}
	}

	return pbReplicaSets
}

// based on github.com/linkerd/linkerd2/controller/k8s/api.go
func (c *Client) getPodsFor(
	obj runtime.Object,
) []*pb.Pod {
	var namespace string
	var selector labels.Selector
	var ownerUID types.UID
	var err error

	switch typed := obj.(type) {
	case *appsv1.DaemonSet:
		namespace = typed.Namespace
		selector = labels.Set(typed.Spec.Selector.MatchLabels).AsSelector()
		ownerUID = typed.UID

	case *appsv1.ReplicaSet:
		namespace = typed.Namespace
		selector = labels.Set(typed.Spec.Selector.MatchLabels).AsSelector()
		ownerUID = typed.UID

	case *appsv1.StatefulSet:
		namespace = typed.Namespace
		selector = labels.Set(typed.Spec.Selector.MatchLabels).AsSelector()
		ownerUID = typed.UID

	default:
		c.log.Errorf("unrecognized runtime object: %v", obj)
		return nil
	}

	pods, err := c.podLister.Pods(namespace).List(selector)
	if err != nil {
		c.log.Errorf("failed to get pods for %s/%v", namespace, selector)
		return nil
	}

	pbPods := []*pb.Pod{}
	for _, pod := range pods {
		if isOwner(ownerUID, pod.GetOwnerReferences()) {
			pbPods = append(pbPods, &pb.Pod{
				Pod: c.serialize(pod, v1.SchemeGroupVersion),
			})
		}
	}

	return pbPods
}

func isOwner(u types.UID, ownerRefs []metav1.OwnerReference) bool {
	for _, or := range ownerRefs {
		if u == or.UID {
			return true
		}
	}
	return false
}
