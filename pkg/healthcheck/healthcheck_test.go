package healthcheck

import (
	"bytes"
	"testing"

	"github.com/buoyantio/linkerd-buoyant/pkg/k8s"
	"github.com/linkerd/linkerd2/pkg/healthcheck"
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHealthChecker(t *testing.T) {
	fixtures := []*struct {
		testName string
		hc       func() *HealthChecker
		success  bool
		stdout   string
	}{
		{
			"Empty",
			func() *HealthChecker { return NewHealthChecker(nil, &healthcheck.Options{}) },
			true,
			"\nStatus check results are √\n",
		},
		{
			"Bad namespace",
			func() *HealthChecker {
				hc := NewHealthChecker(
					&k8s.MockClient{
						MockNamespace: &v1.Namespace{},
					},
					&healthcheck.Options{},
				)
				hc.AppendCategories(hc.L5dBuoyantCategory())
				return hc
			},
			false,
			`buoyant-cloud
-------------
√ linkerd-buoyant can determine the latest version
√ linkerd-buoyant cli is up-to-date
√ buoyant-cloud Namespace exists
× buoyant-cloud Namespace has correct labels
    missing linkerd.io/extension label

Status check results are ×
`,
		},
		{
			"Success",
			func() *HealthChecker {
				objMeta := metav1.ObjectMeta{
					Name: k8s.Namespace,
					Labels: map[string]string{
						l5dk8s.LinkerdExtensionLabel: "buoyant",
						k8s.PartOfKey:                k8s.PartOfVal,
					},
				}
				objMetaDeploy := metav1.ObjectMeta{
					Name: k8s.Namespace,
					Labels: map[string]string{
						l5dk8s.LinkerdExtensionLabel: "buoyant",
						k8s.PartOfKey:                k8s.PartOfVal,
						k8s.VersionLabel:             "v0.0.28",
					},
				}
				hc := NewHealthChecker(
					&k8s.MockClient{
						MockNamespace: &v1.Namespace{
							ObjectMeta: objMeta,
						},
						MockClusterRole: &rbacv1.ClusterRole{
							ObjectMeta: objMeta,
						},
						MockClusterRoleBinding: &rbacv1.ClusterRoleBinding{
							ObjectMeta: objMeta,
						},
						MockSecret: &v1.Secret{
							ObjectMeta: objMeta,
						},
						MockServiceAccount: &v1.ServiceAccount{
							ObjectMeta: objMeta,
						},
						MockDeployment: &appsv1.Deployment{
							ObjectMeta: objMetaDeploy,
						},
						MockPods: &v1.PodList{
							Items: []v1.Pod{
								{
									Status: v1.PodStatus{
										Phase: "Running",
										ContainerStatuses: []v1.ContainerStatus{
											{Ready: true},
										},
									},
									Spec: v1.PodSpec{
										Containers: []v1.Container{
											{Name: l5dk8s.ProxyContainerName},
										},
									},
								},
							},
						},
					},
					&healthcheck.Options{},
				)
				hc.AppendCategories(hc.L5dBuoyantCategory())
				return hc
			},
			true,
			`buoyant-cloud
-------------
√ linkerd-buoyant can determine the latest version
√ linkerd-buoyant cli is up-to-date
√ buoyant-cloud Namespace exists
√ buoyant-cloud Namespace has correct labels
√ buoyant-cloud-agent ClusterRole exists
√ buoyant-cloud-agent ClusterRoleBinding exists
√ buoyant-cloud-agent ServiceAccount exists
√ buoyant-cloud-id Secret exists
√ buoyant-cloud-agent Deployment exists
√ buoyant-cloud-agent Deployment is running
√ buoyant-cloud-agent Deployment is injected
√ buoyant-cloud-agent is up-to-date
√ buoyant-cloud-metrics Deployment exists
√ buoyant-cloud-metrics Deployment is running
√ buoyant-cloud-metrics Deployment is injected
√ buoyant-cloud-metrics is up-to-date

Status check results are √
`,
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			hc := tc.hc()

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			success := healthcheck.RunChecks(stdout, stderr, hc, healthcheck.TableOutput)
			if tc.success != success {
				t.Errorf("Expected success status: [%v], Got: [%v]", tc.success, success)
			}

			if stdout.String() != tc.stdout {
				t.Errorf("Expected stdout: [%s], Got: [%s]", tc.stdout, stdout.String())
			}
			expStderr := ""
			if stderr.String() != expStderr {
				t.Errorf("Expected stderr: [%s], Got: [%s]", expStderr, stderr.String())
			}
		})
	}
}
