package k8s

import (
	"time"

	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func fakeClient(objects ...runtime.Object) *Client {
	cs := fake.NewSimpleClientset(objects...)
	sharedInformers := informers.NewSharedInformerFactory(cs, 10*time.Minute)

	k8sApi := &l5dk8s.KubernetesAPI{
		Interface: cs,
	}

	return NewClient(sharedInformers, k8sApi, nil, false)
}
