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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	ownerUID  = types.UID("owner")
	ownerRefs = []metav1.OwnerReference{{UID: ownerUID}}

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
	lbls = map[string]string{"appname": "app"}
)

func objectMeta(name string, owners []metav1.OwnerReference) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:            name,
		Namespace:       "namespace",
		UID:             ownerUID,
		Labels:          lbls,
		OwnerReferences: owners,
	}
}

func TestDSToWorkload(t *testing.T) {
	fixtures := []*struct {
		testName string
		ds       *appsv1.DaemonSet
		pods     []runtime.Object
	}{
		{
			"empty object",
			&appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: appsv1.SchemeGroupVersion.Identifier(),
					Kind:       "DaemonSet",
				},
				Spec: appsv1.DaemonSetSpec{
					Selector: &metav1.LabelSelector{},
				},
			},
			nil,
		},
		{
			"populated object",
			&appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: appsv1.SchemeGroupVersion.Identifier(),
					Kind:       "DaemonSet",
				},
				ObjectMeta: objectMeta("ds", nil),
				Spec: appsv1.DaemonSetSpec{
					Template: template,
					Selector: &metav1.LabelSelector{MatchLabels: lbls},
				},
			},
			[]runtime.Object{
				&corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						APIVersion: corev1.SchemeGroupVersion.Identifier(),
						Kind:       "Pod",
					},
					ObjectMeta: objectMeta("pod", ownerRefs),
				},
			},
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			client := fakeClient(tc.pods...)
			err := client.Sync(nil, time.Second)
			if err != nil {
				t.Error(err)
			}
			workload := client.DSToWorkload(tc.ds)

			nonEmptyPods := 0
			for _, p := range workload.GetDaemonset().Pods {
				if p.Pod != nil {
					nonEmptyPods = nonEmptyPods + 1
				}
			}

			if nonEmptyPods != len(tc.pods) {
				t.Errorf("Expected: [%d] pods, got: [%d]", len(tc.pods), len(workload.GetDaemonset().Pods))
			}

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
		pods     []runtime.Object
		rs       []runtime.Object
	}{
		{
			"empty object",
			&appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					APIVersion: appsv1.SchemeGroupVersion.Identifier(),
					Kind:       "Deployment",
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{},
				},
			},
			nil,
			nil,
		},
		{
			"populated object",
			&appsv1.Deployment{
				ObjectMeta: objectMeta("deploy", nil),
				TypeMeta: metav1.TypeMeta{
					APIVersion: appsv1.SchemeGroupVersion.Identifier(),
					Kind:       "Deployment",
				},
				Spec: appsv1.DeploymentSpec{
					Template: template,
					Selector: &metav1.LabelSelector{MatchLabels: lbls},
				},
			},
			[]runtime.Object{
				&corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						APIVersion: corev1.SchemeGroupVersion.Identifier(),
						Kind:       "Pod",
					},
					ObjectMeta: objectMeta("pod", ownerRefs),
				},
			},
			[]runtime.Object{
				&appsv1.ReplicaSet{
					TypeMeta: metav1.TypeMeta{
						APIVersion: appsv1.SchemeGroupVersion.Identifier(),
						Kind:       "ReplicaSet",
					},
					ObjectMeta: objectMeta("rs", nil),
					Spec: appsv1.ReplicaSetSpec{
						Selector: &metav1.LabelSelector{MatchLabels: lbls},
					},
				},
			},
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			objs := []runtime.Object{}
			objs = append(objs, tc.rs...)
			objs = append(objs, tc.pods...)

			client := fakeClient(objs...)
			err := client.Sync(nil, time.Second)
			if err != nil {
				t.Error(err)
			}
			workload := client.DeployToWorkload(tc.deploy)

			nonEmptyPods := 0
			for _, rs := range workload.GetDeployment().ReplicaSets {
				for _, p := range rs.Pods {
					if p.Pod != nil {
						nonEmptyPods = nonEmptyPods + 1
					}
				}
			}

			if nonEmptyPods != len(tc.pods) {
				t.Errorf("Expected: [%d] pods, got: [%d]", len(tc.pods), nonEmptyPods)
			}

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
		pods     []runtime.Object
	}{
		{
			"empty object",
			&appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: appsv1.SchemeGroupVersion.Identifier(),
					Kind:       "StatefulSet",
				},
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{},
				},
			},
			nil,
		},
		{
			"populated object",
			&appsv1.StatefulSet{
				ObjectMeta: objectMeta("ss", nil),
				TypeMeta: metav1.TypeMeta{
					APIVersion: appsv1.SchemeGroupVersion.Identifier(),
					Kind:       "StatefulSet",
				},
				Spec: appsv1.StatefulSetSpec{
					Template: template,
					Selector: &metav1.LabelSelector{MatchLabels: lbls},
				},
			},
			[]runtime.Object{
				&corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						APIVersion: corev1.SchemeGroupVersion.Identifier(),
						Kind:       "Pod",
					},
					ObjectMeta: objectMeta("pod", ownerRefs),
				},
			},
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			client := fakeClient(tc.pods...)
			err := client.Sync(nil, time.Second)
			if err != nil {
				t.Error(err)
			}
			workload := client.STSToWorkload(tc.sts)

			nonEmptyPods := 0
			for _, p := range workload.GetStatefulset().Pods {
				if p.Pod != nil {
					nonEmptyPods = nonEmptyPods + 1
				}
			}
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
