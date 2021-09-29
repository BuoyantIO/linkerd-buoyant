package k8s

import (
	"time"

	spclient "github.com/linkerd/linkerd2/controller/gen/client/clientset/versioned"
	spfake "github.com/linkerd/linkerd2/controller/gen/client/clientset/versioned/fake"
	spscheme "github.com/linkerd/linkerd2/controller/gen/client/clientset/versioned/scheme"
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	tsclient "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/clientset/versioned"
	tsfake "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/clientset/versioned/fake"
	tsscheme "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/clientset/versioned/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicfakeclient "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubectl/pkg/scheme"
)

func fakeClient(objects ...runtime.Object) *Client {
	cs, sp, ts, dyn := fakeClientSets(objects...)

	sharedInformers := informers.NewSharedInformerFactory(cs, 10*time.Minute)

	k8sApi := &l5dk8s.KubernetesAPI{
		Interface:     cs,
		TsClient:      ts,
		DynamicClient: dyn,
	}

	client := NewClient(sharedInformers, k8sApi, sp, false)
	client.ignoreCRDSupportCheck = true
	return client
}

func fakeClientSets(objects ...runtime.Object) (kubernetes.Interface, spclient.Interface, tsclient.Interface, dynamic.Interface) {
	spscheme.AddToScheme(scheme.Scheme)
	tsscheme.AddToScheme(scheme.Scheme)

	objs := []runtime.Object{}
	spObjs := []runtime.Object{}
	tsObjs := []runtime.Object{}
	dynamicObjs := []runtime.Object{}

	for _, obj := range objects {
		switch obj.GetObjectKind().GroupVersionKind().Kind {
		case "ServiceProfile":
			spObjs = append(spObjs, obj)
		case "TrafficSplit":
			tsObjs = append(tsObjs, obj)
		case "Link":
			dynamicObjs = append(tsObjs, obj)
		case "ServerAuthorization":
			dynamicObjs = append(tsObjs, obj)
		case "Server":
			dynamicObjs = append(tsObjs, obj)
		default:
			objs = append(objs, obj)
		}
	}

	cs := fake.NewSimpleClientset(objs...)

	return cs,
		spfake.NewSimpleClientset(spObjs...),
		tsfake.NewSimpleClientset(tsObjs...),
		dynamicfakeclient.NewSimpleDynamicClient(scheme.Scheme, dynamicObjs...)
}
