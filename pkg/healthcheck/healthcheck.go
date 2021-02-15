package healthcheck

import (
	"context"
	"fmt"
	"strings"

	"github.com/linkerd/linkerd2/pkg/healthcheck"
	"github.com/linkerd/linkerd2/pkg/k8s"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// categoryID identifies this extension to linkerd check.
	categoryID healthcheck.CategoryID = "linkerd-buoyant"

	// namespace is where the linkerd-buoyant extension runs.
	namespace = "buoyant-cloud"

	partOfKey = "app.kubernetes.io/part-of"
	partOfVal = "buoyant-cloud"

	bcAgent   = "buoyant-cloud-agent"
	bcMetrics = "buoyant-cloud-metrics"
	bcSecret  = "buoyant-cloud-id"
)

// HealthChecker wraps Linkerd's main healthchecker, adding extra fields for
// linkerd-buoyant.
type HealthChecker struct {
	*healthcheck.HealthChecker
	client        kubernetes.Interface
	ns            *v1.Namespace
	metricsDeploy *appsv1.Deployment
	agentPod      v1.Pod
	metricsPod    v1.Pod
	version       string
}

// NewHealthChecker returns an initialized HealthChecker for linkerd-buoyant.
// The returned instance does not contain any linkerd-buoyant Categories.
// Categories are to be explicitly added by using hc.AppendCategories
func NewHealthChecker(
	client kubernetes.Interface,
	options *healthcheck.Options,
) *HealthChecker {
	return &HealthChecker{
		HealthChecker: healthcheck.NewHealthChecker(nil, options),
		client:        client,
	}
}

// TODO: remove HintAnchors ???

// L5dBuoyantCategory returns a healthcheck.Category containing checkers to
// verify the health of linkerd-buoyant components.
func (hc *HealthChecker) L5dBuoyantCategory() healthcheck.Category {
	return *healthcheck.NewCategory(categoryID, []healthcheck.Checker{
		*healthcheck.NewChecker("linkerd-buoyant can determine the latest version").
			WithHintAnchor("l5d-buoyant-version-latest").
			Warning().
			WithCheck(func(ctx context.Context) error {
				// TODO: retrieve from https://buoyant.cloud/version.json
				hc.version = "v0.0.28"
				return nil
			}),
		*healthcheck.NewChecker("linkerd-buoyant cli is up-to-date").
			WithHintAnchor("l5d-buoyant-version-cli").
			Warning().
			WithCheck(func(ctx context.Context) error {
				// TODO: have version number built into this Go binary
				return nil
			}),
		*healthcheck.NewChecker("buoyant-cloud Namespace exists").
			WithHintAnchor("l5d-buoyant-ns-exists").
			Fatal().
			WithCheck(func(ctx context.Context) error {
				ns, err := hc.client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
				if err != nil {
					return err
				}
				hc.ns = ns
				return nil
			}),
		*healthcheck.NewChecker("buoyant-cloud Namespace has correct labels").
			WithHintAnchor("l5d-buoyant-ns-labels").
			Fatal().
			WithCheck(func(ctx context.Context) error {
				err := checkLabel(hc.ns.GetLabels(), k8s.LinkerdExtensionLabel, "buoyant")
				if err != nil {
					return err
				}
				return checkLabel(hc.ns.GetLabels(), partOfKey, partOfVal)
			}),
		*healthcheck.NewChecker("buoyant-cloud-agent ClusterRole exists").
			WithHintAnchor("l5d-buoyant-cr-exists").
			Fatal().
			WithCheck(func(ctx context.Context) error {
				cr, err := hc.client.RbacV1().ClusterRoles().Get(ctx, bcAgent, metav1.GetOptions{})
				if err != nil {
					return err
				}
				return checkLabel(cr.GetLabels(), partOfKey, partOfVal)
			}),
		*healthcheck.NewChecker("buoyant-cloud-agent ClusterRoleBinding exists").
			WithHintAnchor("l5d-buoyant-crb-exists").
			Fatal().
			WithCheck(func(ctx context.Context) error {
				crb, err := hc.client.RbacV1().ClusterRoleBindings().Get(ctx, bcAgent, metav1.GetOptions{})
				if err != nil {
					return err
				}
				return checkLabel(crb.GetLabels(), partOfKey, partOfVal)
			}),
		*healthcheck.NewChecker("buoyant-cloud-agent ServiceAccount exists").
			WithHintAnchor("l5d-buoyant-sa-exists").
			Fatal().
			WithCheck(func(ctx context.Context) error {
				sa, err := hc.client.CoreV1().ServiceAccounts(namespace).Get(ctx, bcAgent, metav1.GetOptions{})
				if err != nil {
					return err
				}
				return checkLabel(sa.GetLabels(), partOfKey, partOfVal)
			}),
		*healthcheck.NewChecker("buoyant-cloud-id Secret exists").
			WithHintAnchor("l5d-buoyant-secret-exists").
			Fatal().
			WithCheck(func(ctx context.Context) error {
				secret, err := hc.client.CoreV1().Secrets(namespace).Get(ctx, bcSecret, metav1.GetOptions{})
				if err != nil {
					return err
				}
				return checkLabel(secret.GetLabels(), partOfKey, partOfVal)
			}),
		*healthcheck.NewChecker("buoyant-cloud-agent Deployment exists").
			WithHintAnchor("l5d-buoyant-agent-exists").
			Fatal().
			WithCheck(func(ctx context.Context) error {
				deploy, err := hc.client.AppsV1().Deployments(namespace).Get(ctx, bcAgent, metav1.GetOptions{})
				if err != nil {
					return err
				}
				return checkLabel(deploy.GetLabels(), partOfKey, partOfVal)
			}),
		*healthcheck.NewChecker("buoyant-cloud-agent Deployment is running").
			WithHintAnchor("l5d-buoyant-agent-running").
			WithCheck(func(ctx context.Context) error {
				pods, err := hc.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: "app=buoyant-cloud-agent"})
				if err != nil {
					return err
				}

				if len(pods.Items) != 1 {
					return fmt.Errorf("expected 1 buoyant-cloud-agent pod, found %d", len(pods.Items))
				}

				hc.agentPod = pods.Items[0]

				return healthcheck.CheckPodsRunning(pods.Items, "")
			}),
		*healthcheck.NewChecker("buoyant-cloud-agent Deployment is injected").
			WithHintAnchor("l5d-buoyant-agent-injected").
			WithCheck(func(ctx context.Context) error {
				return healthcheck.CheckIfDataPlanePodsExist([]v1.Pod{hc.agentPod})
			}),
		*healthcheck.NewChecker("buoyant-cloud-agent is up-to-date").
			WithHintAnchor("l5d-buoyant-version-control").
			Warning().
			WithCheck(func(ctx context.Context) error {
				found := false
				for _, c := range hc.agentPod.Spec.Containers {
					if c.Name != bcAgent {
						continue
					}
					found = true
					image := strings.Split(c.Image, ":")
					if len(image) != 2 {
						return fmt.Errorf("unexpected image: %s", c.Image)
					}
					tag := image[1]
					if tag != hc.version {
						return fmt.Errorf("is running version %s but the latest version is %s", tag, hc.version)
					}
				}
				if !found {
					return fmt.Errorf("%s container not found", bcAgent)
				}
				return nil
			}),
		*healthcheck.NewChecker("buoyant-cloud-metrics Deployment exists").
			WithHintAnchor("l5d-buoyant-metrics-exists").
			Fatal().
			WithCheck(func(ctx context.Context) error {
				deploy, err := hc.client.AppsV1().Deployments(namespace).Get(ctx, bcMetrics, metav1.GetOptions{})
				if err != nil {
					return err
				}
				err = checkLabel(deploy.GetLabels(), partOfKey, partOfVal)
				if err != nil {
					return err
				}
				hc.metricsDeploy = deploy
				return nil
			}),
		*healthcheck.NewChecker("buoyant-cloud-metrics Deployment is running").
			WithHintAnchor("l5d-buoyant-metrics-running").
			WithCheck(func(ctx context.Context) error {
				pods, err := hc.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: "app=buoyant-cloud-metrics"})
				if err != nil {
					return err
				}

				if len(pods.Items) != 1 {
					return fmt.Errorf("expected 1 buoyant-cloud-metrics pod, found %d", len(pods.Items))
				}

				hc.metricsPod = pods.Items[0]

				return healthcheck.CheckPodsRunning(pods.Items, "")
			}),
		*healthcheck.NewChecker("buoyant-cloud-metrics Deployment is injected").
			WithHintAnchor("l5d-buoyant-metrics-injected").
			WithCheck(func(ctx context.Context) error {
				return healthcheck.CheckIfDataPlanePodsExist([]v1.Pod{hc.metricsPod})
			}),
		*healthcheck.NewChecker("buoyant-cloud-metrics Deployment is up-to-date").
			WithHintAnchor("l5d-buoyant-version-control").
			Warning().
			WithCheck(func(ctx context.Context) error {
				return checkLabel(hc.metricsDeploy.GetLabels(), "app.kubernetes.io/version", hc.version)
			}),
	}, true)
}

func checkLabel(labels map[string]string, key, val string) error {
	if l, ok := labels[key]; !ok {
		return fmt.Errorf("missing %s label", key)
	} else if l != val {
		return fmt.Errorf("incorrect %s label: %s, expected: %s", key, l, val)
	}
	return nil
}
