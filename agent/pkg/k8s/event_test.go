package k8s

import (
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestEventToPB(t *testing.T) {
	fixtures := []*struct {
		testName string
		event    *corev1.Event
		objs     []runtime.Object
		err      error
	}{
		{
			"empty event",
			&corev1.Event{},
			nil,
			nil,
		},
		{
			"missing daemonset event",
			&corev1.Event{
				InvolvedObject: corev1.ObjectReference{
					Kind:      DaemonSet,
					Name:      "test-ds",
					Namespace: "test-ns",
				},
			},
			nil,
			k8serrors.NewNotFound(schema.GroupResource{Group: "apps", Resource: "daemonset"}, "test-ds"),
		},
		{
			"daemonset event",
			&corev1.Event{
				InvolvedObject: corev1.ObjectReference{
					Kind:      DaemonSet,
					Name:      "test-ds",
					Namespace: "test-ns",
				},
			},
			[]runtime.Object{
				&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ds",
						Namespace: "test-ns",
					},
					Spec: appsv1.DaemonSetSpec{
						Selector: &metav1.LabelSelector{},
					},
				},
			},
			nil,
		},
		{
			"deployment event",
			&corev1.Event{
				InvolvedObject: corev1.ObjectReference{
					Kind:      Deployment,
					Name:      "test-deploy",
					Namespace: "test-deploy-ns",
				},
			},
			[]runtime.Object{
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deploy",
						Namespace: "test-deploy-ns",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{},
					},
				},
			},
			nil,
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			c := fakeClient(tc.objs...)
			err := c.Sync(nil, time.Second)
			if err != nil {
				t.Error(err)
			}

			_, err = c.EventToPB(tc.event)
			errCmp(t, tc.err, err)
		})
	}
}

func errCmp(t *testing.T, expErr, err error) {
	if expErr == nil && err == nil {
		return
	}

	if expErr == nil && err != nil ||
		expErr != nil && err == nil ||
		expErr.Error() != err.Error() {
		t.Errorf("Unexpected error: [%s], Expected: [%s]", err, expErr)
	}
}
