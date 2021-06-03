package k8s

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/protobuf"
	"k8s.io/client-go/informers"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type Client struct {
	encoders map[runtime.GroupVersioner]runtime.Encoder

	sharedInformers informers.SharedInformerFactory

	podLister    corev1listers.PodLister
	rsLister     appsv1listers.ReplicaSetLister
	dsLister     appsv1listers.DaemonSetLister
	deployLister appsv1listers.DeploymentLister
	stsLister    appsv1listers.StatefulSetLister

	podSynced    cache.InformerSynced
	rsSynced     cache.InformerSynced
	dsSynced     cache.InformerSynced
	deploySynced cache.InformerSynced
	stsSynced    cache.InformerSynced

	eventInformer corev1informers.EventInformer
	eventSynced   cache.InformerSynced

	proxyAddrOverride string

	log *log.Entry
}

const (
	DaemonSet   = "DaemonSet"
	Deployment  = "Deployment"
	Namespace   = "Namespace"
	Pod         = "Pod"
	ReplicaSet  = "ReplicaSet"
	StatefulSet = "StatefulSet"
)

var errSyncCache = errors.New("failed to sync caches")

func NewClient(sharedInformers informers.SharedInformerFactory, proxyAddrOverride string) *Client {
	log := log.WithField("client", "k8s")
	log.Debug("initializing")

	protoSerializer := protobuf.NewSerializer(scheme.Scheme, scheme.Scheme)
	encoders := map[runtime.GroupVersioner]runtime.Encoder{
		v1.SchemeGroupVersion:     scheme.Codecs.EncoderForVersion(protoSerializer, v1.SchemeGroupVersion),
		appsv1.SchemeGroupVersion: scheme.Codecs.EncoderForVersion(protoSerializer, appsv1.SchemeGroupVersion),
	}

	podInformer := sharedInformers.Core().V1().Pods()
	podInformerSynced := podInformer.Informer().HasSynced

	rsInformer := sharedInformers.Apps().V1().ReplicaSets()
	rsInformerSynced := rsInformer.Informer().HasSynced

	dsInformer := sharedInformers.Apps().V1().DaemonSets()
	dsInformerSynced := dsInformer.Informer().HasSynced

	deployInformer := sharedInformers.Apps().V1().Deployments()
	deployInformerSynced := deployInformer.Informer().HasSynced

	stsInformer := sharedInformers.Apps().V1().StatefulSets()
	stsInformerSynced := stsInformer.Informer().HasSynced

	eventInformer := sharedInformers.Core().V1().Events()
	eventInformerSynced := eventInformer.Informer().HasSynced

	return &Client{
		encoders: encoders,

		sharedInformers: sharedInformers,

		podLister:    podInformer.Lister(),
		rsLister:     rsInformer.Lister(),
		dsLister:     dsInformer.Lister(),
		deployLister: deployInformer.Lister(),
		stsLister:    stsInformer.Lister(),

		podSynced:    podInformerSynced,
		rsSynced:     rsInformerSynced,
		dsSynced:     dsInformerSynced,
		deploySynced: deployInformerSynced,
		stsSynced:    stsInformerSynced,

		eventInformer: eventInformer,
		eventSynced:   eventInformerSynced,

		proxyAddrOverride: proxyAddrOverride,

		log: log,
	}
}

func (c *Client) Sync(stopCh <-chan struct{}, timeout time.Duration) error {
	c.sharedInformers.Start(stopCh)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c.log.Infof("waiting for caches to sync")
	if !cache.WaitForCacheSync(
		ctx.Done(),
		c.rsSynced,
		c.podSynced,
		c.dsSynced,
		c.deploySynced,
		c.stsSynced,
		c.eventSynced,
	) {
		c.log.Error(errSyncCache)
		return errSyncCache
	}
	c.log.Infof("caches synced")

	return nil
}

// Serialize takes a k8s object and serializes it into a byte slice.
// For more info on k8s serialization:
// https://github.com/kubernetes/api#recommended-use
func (c *Client) serialize(obj runtime.Object, gv runtime.GroupVersioner) []byte {
	encoder, ok := c.encoders[gv]
	if !ok {
		c.log.Errorf("Unsupported GroupVersioner: %v", gv)
		return nil
	}

	buf, err := runtime.Encode(encoder, obj.DeepCopyObject())
	if err != nil {
		c.log.Errorf("Encode failed: %s", err)
		return nil
	}
	return buf
}
