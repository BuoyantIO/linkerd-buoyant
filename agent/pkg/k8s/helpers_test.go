package k8s

import (
	"time"

	l5dClient "github.com/linkerd/linkerd2/controller/gen/client/clientset/versioned"
	l5dFake "github.com/linkerd/linkerd2/controller/gen/client/clientset/versioned/fake"
	l5dScheme "github.com/linkerd/linkerd2/controller/gen/client/clientset/versioned/scheme"
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	tsclient "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/clientset/versioned"
	tsfake "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/clientset/versioned/fake"
	tsscheme "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/clientset/versioned/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubectl/pkg/scheme"
)

func fakeClient(objects ...runtime.Object) *Client {
	cs, l5dApiClient, ts := fakeClientSets(objects...)

	sharedInformers := informers.NewSharedInformerFactory(cs, 10*time.Minute)

	k8sApi := &l5dk8s.KubernetesAPI{
		Interface: cs,
	}

	client := NewClient(sharedInformers, k8sApi, l5dApiClient, ts, false)
	client.ignoreCRDSupportCheck = true
	return client
}

func fakeClientSets(objects ...runtime.Object) (kubernetes.Interface, l5dClient.Interface, tsclient.Interface) {
	l5dScheme.AddToScheme(scheme.Scheme)
	tsscheme.AddToScheme(scheme.Scheme)

	objs := []runtime.Object{}
	l5dObjects := []runtime.Object{}
	tsObjs := []runtime.Object{}

	for _, obj := range objects {
		switch obj.GetObjectKind().GroupVersionKind().Kind {
		case "ServiceProfile":
			l5dObjects = append(l5dObjects, obj)
		case "ServerAuthorization":
			l5dObjects = append(l5dObjects, obj)
		case "Server":
			l5dObjects = append(l5dObjects, obj)
		case "AuthorizationPolicy":
			l5dObjects = append(l5dObjects, obj)
		case "MeshTLSAuthentication":
			l5dObjects = append(l5dObjects, obj)
		case "NetworkAuthentication":
			l5dObjects = append(l5dObjects, obj)
		case "Link":
			l5dObjects = append(l5dObjects, obj)
		case "TrafficSplit":
			tsObjs = append(tsObjs, obj)
		default:
			objs = append(objs, obj)
		}
	}

	cs := fake.NewSimpleClientset(objs...)

	return cs,
		l5dFake.NewSimpleClientset(l5dObjects...),
		tsfake.NewSimpleClientset(tsObjs...)
}
