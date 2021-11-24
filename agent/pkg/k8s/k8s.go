package k8s

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	link "github.com/linkerd/linkerd2/controller/gen/apis/link/v1alpha1"
	server "github.com/linkerd/linkerd2/controller/gen/apis/server/v1beta1"
	serverAuthorization "github.com/linkerd/linkerd2/controller/gen/apis/serverauthorization/v1beta1"
	sp "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	l5dApi "github.com/linkerd/linkerd2/controller/gen/client/clientset/versioned"
	l5dscheme "github.com/linkerd/linkerd2/controller/gen/client/clientset/versioned/scheme"
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	ts "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha1"
	tsscheme "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/clientset/versioned/scheme"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/protobuf"
	"k8s.io/client-go/informers"
	corev1informers "k8s.io/client-go/informers/core/v1"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubectl/pkg/scheme"
)

type Client struct {
	k8sClient *l5dk8s.KubernetesAPI
	l5dClient l5dApi.Interface

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

	log   *log.Entry
	local bool

	// for testing
	ignoreCRDSupportCheck bool
}

type containerConnection struct {
	host    string
	cleanup func()
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

func NewClient(sharedInformers informers.SharedInformerFactory, k8sClient *l5dk8s.KubernetesAPI, l5dClient l5dApi.Interface, local bool) *Client {
	log := log.WithField("client", "k8s")
	log.Debug("initializing")

	l5dscheme.AddToScheme(scheme.Scheme)
	tsscheme.AddToScheme(scheme.Scheme)

	protoSerializer := protobuf.NewSerializer(scheme.Scheme, scheme.Scheme)
	jsonSerializer := scheme.DefaultJSONEncoder()

	// We handle the CRDs differently depending on whether we have a typed client provided
	// For types that we do not have a cleint for we use the `unstructured` package. As we
	// also do not have proto definitions for these CRDs, we serialize them to JSON.
	//
	// +-------------------------+----------------------+-----------+------------+
	// |          Group          |        Kind          |  Client   | Serializer |
	// +-------------------------+----------------------+-----------+------------+
	// | policy.linkerd.io       | serverAuthorizations | l5dClient | json       |
	// | policy.linkerd.io       | servers              | l5dClient | json       |
	// | multicluster.linkerd.io | links                | l5dClient | json       |
	// | linkerd.io              | serviceprofiles      | l5dClient | json       |
	// | split.smi-spec.io       | trafficsplits        | tsclient  | json       |
	// +-------------------------+----------------------+-----------+------------+

	encoders := map[runtime.GroupVersioner]runtime.Encoder{
		v1.SchemeGroupVersion:                  scheme.Codecs.EncoderForVersion(protoSerializer, v1.SchemeGroupVersion),
		appsv1.SchemeGroupVersion:              scheme.Codecs.EncoderForVersion(protoSerializer, appsv1.SchemeGroupVersion),
		ts.SchemeGroupVersion:                  scheme.Codecs.EncoderForVersion(jsonSerializer, ts.SchemeGroupVersion),
		sp.SchemeGroupVersion:                  scheme.Codecs.EncoderForVersion(jsonSerializer, sp.SchemeGroupVersion),
		link.SchemeGroupVersion:                scheme.Codecs.EncoderForVersion(jsonSerializer, link.SchemeGroupVersion),
		serverAuthorization.SchemeGroupVersion: scheme.Codecs.EncoderForVersion(jsonSerializer, serverAuthorization.SchemeGroupVersion),
		server.SchemeGroupVersion:              scheme.Codecs.EncoderForVersion(jsonSerializer, server.SchemeGroupVersion),
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

		k8sClient: k8sClient,
		l5dClient: l5dClient,
		log:       log,
		local:     local,
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

func (c *Client) localMode() bool {
	return c.local
}

// this method establishes a connection to a specific container in a pod
// and gives you the host addr. This logic is abstracted away in order to
// enable running this agent outside of a K8s cluster for the purpose of
// local development. The `containerConnection` struct returned contains
// a `cleanup()` function that must be called when this connection is not
// needed anymore
func (c *Client) getContainerConnection(pod *v1.Pod, container *v1.Container, portName string) (*containerConnection, error) {
	if c.localMode() {
		// running in local mode, we need a port forward
		pf, err := l5dk8s.NewContainerMetricsForward(c.k8sClient, *pod, *container, false, l5dk8s.ProxyAdminPortName)
		if err != nil {
			return nil, err
		}

		// not very elegant... We need a way to get the port and host from PortForward
		httpUrl, err := url.Parse(pf.URLFor(""))
		if err != nil {
			return nil, err
		}

		if err = pf.Init(); err != nil {
			return nil, err
		}

		return &containerConnection{
			host:    httpUrl.Host,
			cleanup: func() { pf.Stop() },
		}, nil
	} else {
		port, err := getContainerPort(container, portName)
		if err != nil {
			return nil, err
		}

		return &containerConnection{
			host:    fmt.Sprintf("%s:%d", pod.Status.PodIP, port),
			cleanup: func() {}, // noop
		}, nil
	}
}

func (c *Client) resourceSupported(gvr schema.GroupVersionResource) (bool, error) {
	if c.ignoreCRDSupportCheck {
		return true, nil
	}

	gv := gvr.GroupVersion().String()
	res, err := c.k8sClient.Discovery().ServerResourcesForGroupVersion(gv)
	if err != nil && !kerrors.IsNotFound(err) {
		return false, err
	}

	if res != nil && res.GroupVersion == gv {
		for _, apiRes := range res.APIResources {
			if apiRes.Name == gvr.Resource {
				return true, nil
			}
		}
	}

	c.log.Debugf("Resource %+v not supported", gvr)
	return false, nil
}

func getContainerPort(container *v1.Container, portName string) (int32, error) {
	for _, p := range container.Ports {
		if p.Name == portName {
			return p.ContainerPort, nil
		}
	}

	return 0, fmt.Errorf("could not find port %s on container [%s]", portName, container.Name)
}

func getProxyContainer(pod *v1.Pod) (*v1.Container, error) {
	for _, c := range pod.Spec.Containers {
		if c.Name == l5dk8s.ProxyContainerName {
			container := c
			return &container, nil
		}
	}

	return nil, fmt.Errorf("could not find proxy container in pod %s/%s", pod.Namespace, pod.Name)
}
