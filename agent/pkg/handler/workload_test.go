package handler

import (
	"context"
	"testing"
	"time"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/api"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestWorkloadStream(t *testing.T) {
	workloadName := "fake-workload"
	workloadNS := "fake-ns"

	ds1 := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadName + "-1",
			Namespace: workloadNS + "-1",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{},
		},
	}
	ds2 := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadName + "-2",
			Namespace: workloadNS + "-2",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{},
		},
	}
	deploy1 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadName + "-1",
			Namespace: workloadNS + "-1",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{},
		},
	}
	deploy2 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadName + "-2",
			Namespace: workloadNS + "-2",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{},
		},
	}
	sts1 := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadName + "-1",
			Namespace: workloadNS + "-1",
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{},
		},
	}
	sts2 := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadName + "-2",
			Namespace: workloadNS + "-2",
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{},
		},
	}

	fixtures := []*struct {
		testName          string
		existingWorkloads map[string][]runtime.Object
		createWorkloads   map[string][]runtime.Object
		updateWorkloads   map[string][]runtime.Object
		deleteWorkloads   map[string][]runtime.Object
		expMessages       int
	}{
		{
			"no workloads",
			nil,
			nil,
			nil,
			nil,
			0,
		},
		{
			"three existing workloads",
			map[string][]runtime.Object{
				k8s.DaemonSet:   {ds1},
				k8s.Deployment:  {deploy1},
				k8s.StatefulSet: {sts1},
			},
			nil,
			nil,
			nil,
			4,
		},
		{
			"three create workloads",
			nil,
			map[string][]runtime.Object{
				k8s.DaemonSet:   {ds1},
				k8s.Deployment:  {deploy1},
				k8s.StatefulSet: {sts1},
			},
			nil,
			nil,
			4,
		},
		{
			"three existing and three create workloads",
			map[string][]runtime.Object{
				k8s.DaemonSet:   {ds1},
				k8s.Deployment:  {deploy1},
				k8s.StatefulSet: {sts1},
			},
			map[string][]runtime.Object{
				k8s.DaemonSet:   {ds2},
				k8s.Deployment:  {deploy2},
				k8s.StatefulSet: {sts2},
			},
			nil,
			nil,
			7,
		},
		{
			"three existing and three update workloads",
			map[string][]runtime.Object{
				k8s.DaemonSet:   {ds1},
				k8s.Deployment:  {deploy1},
				k8s.StatefulSet: {sts1},
			},
			nil,
			map[string][]runtime.Object{
				k8s.DaemonSet:   {ds1},
				k8s.Deployment:  {deploy1},
				k8s.StatefulSet: {sts1},
			},
			nil,
			7,
		},
		{
			"three existing and three delete workloads",
			map[string][]runtime.Object{
				k8s.DaemonSet:   {ds1},
				k8s.Deployment:  {deploy1},
				k8s.StatefulSet: {sts1},
			},
			nil,
			nil,
			map[string][]runtime.Object{
				k8s.DaemonSet:   {ds1},
				k8s.Deployment:  {deploy1},
				k8s.StatefulSet: {sts1},
			},
			7,
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			objs := []runtime.Object{}
			for _, wrks := range tc.existingWorkloads {
				objs = append(objs, wrks...)
			}
			cs := fake.NewSimpleClientset(objs...)
			sharedInformers := informers.NewSharedInformerFactory(cs, 10*time.Minute)
			k8sClient := k8s.NewClient(cs, sharedInformers, nil)

			m := &api.MockBcloudClient{}
			apiClient := api.NewClient("", "", m)

			wh := NewWorkload(k8sClient, apiClient)
			if len(m.Messages()) != 0 {
				t.Errorf("Expected no messages sent, got %d", len(m.Messages()))
			}
			go wh.Start(sharedInformers)
			err := k8sClient.Sync(nil, time.Second)
			if err != nil {
				t.Error(err)
			}

			for k, objs := range tc.createWorkloads {
				for _, o := range objs {
					switch k {
					case k8s.DaemonSet:
						ds, ok := o.(*appsv1.DaemonSet)
						if !ok {
							t.Errorf("failed type assertion to appsv1.DaemonSet: %+v", o)
						}

						_, err = cs.AppsV1().DaemonSets(ds.Namespace).Create(context.TODO(), ds, metav1.CreateOptions{})
						if err != nil {
							t.Errorf("Error injecting appsv1.DaemonSet: %v", err)
						}
					case k8s.Deployment:
						deploy, ok := o.(*appsv1.Deployment)
						if !ok {
							t.Errorf("failed type assertion to appsv1.Deployment: %+v", o)
						}

						_, err = cs.AppsV1().Deployments(deploy.Namespace).Create(context.TODO(), deploy, metav1.CreateOptions{})
						if err != nil {
							t.Errorf("Error injecting appsv1.Deployment: %v", err)
						}
					case k8s.StatefulSet:
						sts, ok := o.(*appsv1.StatefulSet)
						if !ok {
							t.Errorf("failed type assertion to appsv1.StatefulSet: %+v", o)
						}

						_, err = cs.AppsV1().StatefulSets(sts.Namespace).Create(context.TODO(), sts, metav1.CreateOptions{})
						if err != nil {
							t.Errorf("Error injecting appsv1.StatefulSet: %v", err)
						}
					}
				}
			}

			for k, objs := range tc.updateWorkloads {
				for _, o := range objs {
					switch k {
					case k8s.DaemonSet:
						ds, ok := o.(*appsv1.DaemonSet)
						if !ok {
							t.Errorf("failed type assertion to appsv1.DaemonSet: %+v", o)
						}

						_, err = cs.AppsV1().DaemonSets(ds.Namespace).Update(context.TODO(), ds, metav1.UpdateOptions{})
						if err != nil {
							t.Errorf("Error injecting appsv1.DaemonSet: %v", err)
						}
					case k8s.Deployment:
						deploy, ok := o.(*appsv1.Deployment)
						if !ok {
							t.Errorf("failed type assertion to appsv1.Deployment: %+v", o)
						}

						_, err = cs.AppsV1().Deployments(deploy.Namespace).Update(context.TODO(), deploy, metav1.UpdateOptions{})
						if err != nil {
							t.Errorf("Error injecting appsv1.Deployment: %v", err)
						}
					case k8s.StatefulSet:
						sts, ok := o.(*appsv1.StatefulSet)
						if !ok {
							t.Errorf("failed type assertion to appsv1.StatefulSet: %+v", o)
						}

						_, err = cs.AppsV1().StatefulSets(sts.Namespace).Update(context.TODO(), sts, metav1.UpdateOptions{})
						if err != nil {
							t.Errorf("Error injecting appsv1.StatefulSet: %v", err)
						}
					}
				}
			}

			for k, objs := range tc.deleteWorkloads {
				for _, o := range objs {
					switch k {
					case k8s.DaemonSet:
						ds, ok := o.(*appsv1.DaemonSet)
						if !ok {
							t.Errorf("failed type assertion to appsv1.DaemonSet: %+v", o)
						}

						err = cs.AppsV1().DaemonSets(ds.Namespace).Delete(context.TODO(), ds.GetName(), metav1.DeleteOptions{})
						if err != nil {
							t.Errorf("Error injecting appsv1.DaemonSet: %v", err)
						}
					case k8s.Deployment:
						deploy, ok := o.(*appsv1.Deployment)
						if !ok {
							t.Errorf("failed type assertion to appsv1.Deployment: %+v", o)
						}

						err = cs.AppsV1().Deployments(deploy.Namespace).Delete(context.TODO(), deploy.GetName(), metav1.DeleteOptions{})
						if err != nil {
							t.Errorf("Error injecting appsv1.Deployment: %v", err)
						}
					case k8s.StatefulSet:
						sts, ok := o.(*appsv1.StatefulSet)
						if !ok {
							t.Errorf("failed type assertion to appsv1.StatefulSet: %+v", o)
						}

						err = cs.AppsV1().StatefulSets(sts.Namespace).Delete(context.TODO(), sts.GetName(), metav1.DeleteOptions{})
						if err != nil {
							t.Errorf("Error injecting appsv1.StatefulSet: %v", err)
						}
					}
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			for {
				if len(m.Messages()) == tc.expMessages {
					break
				}

				select {
				case <-time.After(10 * time.Millisecond):
				case <-ctx.Done():
					t.Errorf("Expected %d messages(s) sent to API, got %d", tc.expMessages, len(m.Messages()))
					return
				}
			}
		})
	}
}
