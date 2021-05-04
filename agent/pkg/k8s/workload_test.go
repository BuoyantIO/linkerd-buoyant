package k8s

import (
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	om = metav1.ObjectMeta{
		Name:      "name",
		Namespace: "namespace",
	}
	template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "template-name",
			Namespace: "template-namespace",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "container-name"},
			},
		},
	}
)

func TestDSToWorkload(t *testing.T) {
	fixtures := []*struct {
		testName string
		ds       *appsv1.DaemonSet
	}{
		{
			"empty object",
			&appsv1.DaemonSet{
				Spec: appsv1.DaemonSetSpec{
					Selector: &metav1.LabelSelector{},
				},
			},
		},
		{
			"populated object",
			&appsv1.DaemonSet{
				ObjectMeta: om,
				Spec: appsv1.DaemonSetSpec{
					Template: template,
					Selector: &metav1.LabelSelector{},
				},
			},
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			workload := fakeClient().DSToWorkload(tc.ds)

			obj, gvk, err := deserialize(
				workload.GetDaemonset().DaemonSet,
			)
			if err != nil {
				t.Error(err)
			}
			ds, ok := obj.(*appsv1.DaemonSet)
			if !ok {
				t.Errorf("failed type assertion to appsv1.DaemonSet: %+v", obj)
			}

			if !equality.Semantic.DeepEqual(ds.String(), tc.ds.String()) {
				t.Errorf("Expected: [%+v], got: [%+v]", tc.ds.String(), ds.String())
			}

			expectedGVK := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: DaemonSet}
			if gvk.String() != expectedGVK.String() {
				t.Errorf("Expected: [%s], got: [%s]", expectedGVK, gvk)
			}
		})
	}
}

func TestDeployToWorkload(t *testing.T) {
	fixtures := []*struct {
		testName string
		deploy   *appsv1.Deployment
	}{
		{
			"empty object",
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{},
				},
			},
		},
		{
			"populated object",
			&appsv1.Deployment{
				ObjectMeta: om,
				Spec: appsv1.DeploymentSpec{
					Template: template,
					Selector: &metav1.LabelSelector{},
				},
			},
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			workload := fakeClient().DeployToWorkload(tc.deploy)

			obj, gvk, err := deserialize(
				workload.GetDeployment().Deployment,
			)
			if err != nil {
				t.Error(err)
			}
			deploy, ok := obj.(*appsv1.Deployment)
			if !ok {
				t.Errorf("failed type assertion to appsv1.Deployment: %+v", obj)
			}

			if !equality.Semantic.DeepEqual(deploy.String(), tc.deploy.String()) {
				t.Errorf("Expected: [%+v], got: [%+v]", tc.deploy.String(), deploy.String())
			}

			expectedGVK := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: Deployment}
			if gvk.String() != expectedGVK.String() {
				t.Errorf("Expected: [%s], got: [%s]", expectedGVK, gvk)
			}
		})
	}
}

func TestSTSToWorkload(t *testing.T) {
	fixtures := []*struct {
		testName string
		sts      *appsv1.StatefulSet
	}{
		{
			"empty object",
			&appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{},
				},
			},
		},
		{
			"populated object",
			&appsv1.StatefulSet{
				ObjectMeta: om,
				Spec: appsv1.StatefulSetSpec{
					Template: template,
					Selector: &metav1.LabelSelector{},
				},
			},
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			workload := fakeClient().STSToWorkload(tc.sts)

			obj, gvk, err := deserialize(
				workload.GetStatefulset().StatefulSet,
			)
			if err != nil {
				t.Error(err)
			}
			sts, ok := obj.(*appsv1.StatefulSet)
			if !ok {
				t.Errorf("failed type assertion to appsv1.StatefulSet: %+v", obj)
			}

			if !equality.Semantic.DeepEqual(sts.String(), tc.sts.String()) {
				t.Errorf("Expected: [%+v], got: [%+v]", tc.sts.String(), sts.String())
			}

			expectedGVK := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: StatefulSet}
			if gvk.String() != expectedGVK.String() {
				t.Errorf("Expected: [%s], got: [%s]", expectedGVK, gvk)
			}
		})
	}
}

func TestListWorkloads(t *testing.T) {
	objects := []runtime.Object{
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ds-1",
			},
			Spec: appsv1.DaemonSetSpec{
				Selector: &metav1.LabelSelector{},
			},
		},
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ds-2",
			},
			Spec: appsv1.DaemonSetSpec{
				Selector: &metav1.LabelSelector{},
			},
		},
		&appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{},
			},
		},
		&appsv1.StatefulSet{
			Spec: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{},
			},
		},
		&corev1.Endpoints{},
		&corev1.Pod{},
	}
	c := fakeClient(objects...)
	c.Sync(nil, time.Second)

	workloads, err := c.ListWorkloads()
	if err != nil {
		t.Error(err)
	}

	if len(workloads) != 4 {
		t.Errorf("Expected [4] workloads, got [%d]: %+v", len(workloads), workloads)
	}
}

//
// test helpers
//

func deserialize(b []byte) (runtime.Object, *schema.GroupVersionKind, error) {
	return scheme.Codecs.UniversalDeserializer().Decode(b, nil, nil)
}
