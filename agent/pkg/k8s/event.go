package k8s

import (
	"fmt"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	v1 "k8s.io/api/core/v1"
)

// EventToPB converts a Kubernetes event to a bcloud protobuf event. If the
// owner of the event is not a DeamonSet, Deployment, or StatefulSet, return a
// nil event without error.
func (c *Client) EventToPB(event *v1.Event) (*pb.Event, error) {
	involvedObj := event.InvolvedObject
	name := involvedObj.Name
	kind := involvedObj.Kind

	switch involvedObj.Kind {
	case ReplicaSet:
		var err error
		name, err = c.rsToDeploy(involvedObj.Name, involvedObj.Namespace)
		if err != nil {
			c.log.Debugf("Failed to get owner for ReplicaSet event [%s/%s]: %s", involvedObj.Namespace, involvedObj.Name, err)
			return nil, err
		}
		kind = Deployment
	case Pod:
		var err error
		name, kind, err = c.podToWorkload(involvedObj.Name, involvedObj.Namespace)
		if err != nil {
			c.log.Debugf("Failed to get owner for Pod event [%s/%s]: %s", involvedObj.Namespace, involvedObj.Name, err)
			return nil, err
		}
	}

	if !isValidWorkloadKind(kind) {
		c.log.Tracef("invalid workload kind: %s", kind)
		return nil, nil
	}

	workload, err := c.createWorkload(name, involvedObj.Namespace, kind)
	if err != nil {
		c.log.Errorf("Failed to create workload for [%s/%s/%s]: %s", kind, involvedObj.Namespace, name, err)
		return nil, err
	}

	return &pb.Event{
		Event: c.serialize(event, v1.SchemeGroupVersion),
		Owner: workload,
	}, nil
}

func (c *Client) rsToDeploy(replicaSetName, namespace string) (string, error) {
	deployName := ""
	rs, err := c.rsLister.ReplicaSets(namespace).Get(replicaSetName)
	if err != nil {
		return deployName, err
	}
	rsOwners := rs.GetOwnerReferences()
	if len(rsOwners) > 0 {
		deployName = rsOwners[0].Name
	}
	return deployName, nil
}

func (c *Client) podToWorkload(podName, namespace string) (string, string, error) {
	p, err := c.podLister.Pods(namespace).Get(podName)
	if err != nil {
		return "", "", err
	}

	var workloadName, workloadKind string

	podOwners := p.GetOwnerReferences()
	if len(podOwners) > 0 {
		ownerRs := podOwners[0]
		if ownerRs.Kind == ReplicaSet {
			workloadKind = Deployment
			workloadName, err = c.rsToDeploy(ownerRs.Name, namespace)
			if err != nil {
				return "", "", err
			}
		} else {
			workloadKind = ownerRs.Kind
			workloadName = ownerRs.Name
		}
	}
	return workloadName, workloadKind, nil
}

func (c *Client) createWorkload(name, namespace, kind string) (*pb.Workload, error) {
	if !isValidWorkloadKind(kind) {
		return nil, fmt.Errorf("can't handle events for unsupported resource kind %s", kind)
	}

	var workload *pb.Workload
	switch kind {
	case DaemonSet:
		ds, err := c.dsLister.DaemonSets(namespace).Get(name)
		if err != nil {
			return nil, err
		}
		workload = c.DSToWorkload(ds)

	case Deployment:
		deploy, err := c.deployLister.Deployments(namespace).Get(name)
		if err != nil {
			return nil, err
		}
		workload = c.DeployToWorkload(deploy)

	case StatefulSet:
		sts, err := c.stsLister.StatefulSets(namespace).Get(name)
		if err != nil {
			return nil, err
		}
		workload = c.STSToWorkload(sts)
	}

	return workload, nil
}

// isValidWorkloadKind returns true if the given kind is one of daemonSet,
// deployment, or statefulset. Otherwise, it returns false.
func isValidWorkloadKind(kind string) bool {
	return kind == DaemonSet || kind == Deployment || kind == StatefulSet
}
