package k8s

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func fakeClient(objects ...runtime.Object) *Client {
	cs := fake.NewSimpleClientset(objects...)
	sharedInformers := informers.NewSharedInformerFactory(cs, 10*time.Minute)
	return NewClient(sharedInformers, "")
}
