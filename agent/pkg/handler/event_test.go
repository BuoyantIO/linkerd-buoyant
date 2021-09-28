package handler

import (
	"context"
	"testing"
	"time"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/api"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEvent(t *testing.T) {
	dsName := "fake-ds"
	dsNamespace := "fake-ds-ns"
	deployName := "fake-deploy"
	deployNamespace := "fake-deploy-ns"

	fixtures := []*struct {
		testName  string
		events    []*corev1.Event
		objs      []runtime.Object
		expEvents int
	}{
		{
			"no events",
			nil,
			nil,
			0,
		},
		{
			"one event",
			[]*corev1.Event{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-event",
					},
					InvolvedObject: corev1.ObjectReference{
						Kind:      k8s.DaemonSet,
						Name:      dsName,
						Namespace: dsNamespace,
					},
				},
			},
			[]runtime.Object{
				&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      dsName,
						Namespace: dsNamespace,
					},
					Spec: appsv1.DaemonSetSpec{
						Selector: &metav1.LabelSelector{},
					},
				},
			},
			1,
		},
		{
			"one daemonset event and one deployment event",
			[]*corev1.Event{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ds-event",
					},
					InvolvedObject: corev1.ObjectReference{
						Kind:      k8s.DaemonSet,
						Name:      dsName,
						Namespace: dsNamespace,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deploy-event",
					},
					InvolvedObject: corev1.ObjectReference{
						Kind:      k8s.Deployment,
						Name:      deployName,
						Namespace: deployNamespace,
					},
				},
			},
			[]runtime.Object{
				&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      dsName,
						Namespace: dsNamespace,
					},
					Spec: appsv1.DaemonSetSpec{
						Selector: &metav1.LabelSelector{},
					},
				},
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deployName,
						Namespace: deployNamespace,
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{},
					},
				},
			},
			2,
		},
		{
			"one event without an owner",
			[]*corev1.Event{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-event",
					},
					InvolvedObject: corev1.ObjectReference{
						Kind:      k8s.DaemonSet,
						Name:      dsName,
						Namespace: dsNamespace,
					},
				},
			},
			nil,
			0,
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			cs := fake.NewSimpleClientset(tc.objs...)
			sharedInformers := informers.NewSharedInformerFactory(cs, 10*time.Minute)
			k8sApi := &l5dk8s.KubernetesAPI{
				Interface: cs,
			}
			k8sClient := k8s.NewClient(sharedInformers, k8sApi, nil, false)

			m := &api.MockBcloudClient{}
			apiClient := api.NewClient("", "", m)

			eh := NewEvent(k8sClient, apiClient)
			if len(m.Events()) != 0 {
				t.Errorf("Expected no events sent, got %d", len(m.Events()))
			}
			eh.Start(sharedInformers)
			err := k8sClient.Sync(nil, time.Second)
			if err != nil {
				t.Error(err)
			}

			for _, e := range tc.events {
				_, err = cs.CoreV1().Events(dsNamespace).Create(context.TODO(), e, metav1.CreateOptions{})
				if err != nil {
					t.Errorf("Error injecting event: %v", err)
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			for {
				if len(m.Events()) == tc.expEvents {
					break
				}

				select {
				case <-time.After(10 * time.Millisecond):
				case <-ctx.Done():
					t.Errorf("Expected %d event(s) sent to API, got %d", tc.expEvents, len(m.Events()))
					return
				}
			}
		})
	}
}
