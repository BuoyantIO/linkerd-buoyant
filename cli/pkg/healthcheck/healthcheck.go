package healthcheck

import (
	"context"
	"fmt"
	"net/http"

	agentk8s "github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	"github.com/buoyantio/linkerd-buoyant/cli/pkg/k8s"
	"github.com/buoyantio/linkerd-buoyant/cli/pkg/version"
	"github.com/linkerd/linkerd2/pkg/healthcheck"
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	v1 "k8s.io/api/core/v1"
)

const (
	// categoryID identifies this extension to linkerd check.
	categoryID healthcheck.CategoryID = version.LinkerdBuoyant
)

// HealthChecker wraps Linkerd's main healthchecker, adding extra fields for
// linkerd-buoyant.
type HealthChecker struct {
	*healthcheck.HealthChecker
	k8s          k8s.Client
	http         *http.Client
	bcloudServer string

	// these fields are used as caches between checks
	version string
	ns      *v1.Namespace
}

type getWorkloadLabelsFn func(ctx context.Context) (map[string]string, error)

// NewHealthChecker returns an initialized HealthChecker for linkerd-buoyant.
// The returned instance does not contain any linkerd-buoyant Categories.
// Categories are to be explicitly added by using hc.AppendCategories
func NewHealthChecker(
	options *healthcheck.Options,
	k8s k8s.Client,
	http *http.Client,
	bcloudServer string,
) *HealthChecker {
	return &HealthChecker{
		HealthChecker: healthcheck.NewHealthChecker(nil, options),
		k8s:           k8s,
		http:          http,
		bcloudServer:  bcloudServer,
	}
}

// L5dBuoyantCategory returns a healthcheck.Category containing checkers to
// verify the health of linkerd-buoyant components.
func (hc *HealthChecker) L5dBuoyantCategory() *healthcheck.Category {
	checks := append(
		hc.globalChecks(),
		append(
			hc.deploymentChecks(k8s.AgentName),
			hc.daemonSetChecks(k8s.MetricsName)...,
		)...,
	)
	return healthcheck.NewCategory(categoryID, checks, true).WithHintBaseURL("https://linkerd.io/checks#l5d-buoyant")
}

func (hc *HealthChecker) globalChecks() []healthcheck.Checker {
	return []healthcheck.Checker{
		*healthcheck.NewChecker("linkerd-buoyant can determine the latest version").
			Warning().
			WithCheck(func(ctx context.Context) error {
				url := fmt.Sprintf("%s/version.json", hc.bcloudServer)
				version, err := version.Get(ctx, hc.http, url)
				if err != nil {
					return err
				}
				hc.version = version
				return nil
			}),
		*healthcheck.NewChecker("linkerd-buoyant cli is up-to-date").
			Warning().
			WithCheck(func(ctx context.Context) error {
				if version.Version != hc.version {
					return fmt.Errorf("CLI version is %s but the latest is %s", version.Version, hc.version)
				}
				return nil
			}),
		*healthcheck.NewChecker("buoyant-cloud Namespace exists").
			Fatal().
			WithCheck(func(ctx context.Context) error {
				ns, err := hc.k8s.Namespace(ctx)
				if err != nil {
					return err
				}
				hc.ns = ns
				return nil
			}),
		*healthcheck.NewChecker("buoyant-cloud Namespace has correct labels").
			Fatal().
			WithCheck(func(ctx context.Context) error {
				err := checkLabel(hc.ns.GetLabels(), l5dk8s.LinkerdExtensionLabel, "buoyant")
				if err != nil {
					return err
				}
				return checkLabel(hc.ns.GetLabels(), k8s.PartOfKey, k8s.PartOfVal)
			}),
		*healthcheck.NewChecker("buoyant-cloud-agent ClusterRole exists").
			Fatal().
			WithCheck(func(ctx context.Context) error {
				cr, err := hc.k8s.ClusterRole(ctx)
				if err != nil {
					return err
				}
				return checkLabel(cr.GetLabels(), k8s.PartOfKey, k8s.PartOfVal)
			}),
		*healthcheck.NewChecker("buoyant-cloud-agent ClusterRoleBinding exists").
			Fatal().
			WithCheck(func(ctx context.Context) error {
				crb, err := hc.k8s.ClusterRoleBinding(ctx)
				if err != nil {
					return err
				}
				return checkLabel(crb.GetLabels(), k8s.PartOfKey, k8s.PartOfVal)
			}),
		*healthcheck.NewChecker("buoyant-cloud-agent ServiceAccount exists").
			Fatal().
			WithCheck(func(ctx context.Context) error {
				sa, err := hc.k8s.ServiceAccount(ctx)
				if err != nil {
					return err
				}
				return checkLabel(sa.GetLabels(), k8s.PartOfKey, k8s.PartOfVal)
			}),
		*healthcheck.NewChecker("buoyant-cloud-id Secret exists").
			Fatal().
			WithCheck(func(ctx context.Context) error {
				secret, err := hc.k8s.Secret(ctx)
				if err != nil {
					return err
				}
				return checkLabel(secret.GetLabels(), k8s.PartOfKey, k8s.PartOfVal)
			}),
	}
}

func (hc *HealthChecker) deploymentChecks(name string) []healthcheck.Checker {
	getWorkloadLabels := func(ctx context.Context) (map[string]string, error) {
		deploy, err := hc.k8s.Deployment(ctx, name)
		if err != nil {
			return nil, err
		}
		return deploy.GetLabels(), nil
	}

	return hc.workloadChecks(name, agentk8s.Deployment, getWorkloadLabels, true)
}

func (hc *HealthChecker) daemonSetChecks(name string) []healthcheck.Checker {
	getWorkloadLabels := func(ctx context.Context) (map[string]string, error) {
		ds, err := hc.k8s.DaemonSet(ctx, name)
		if err != nil {
			return nil, err
		}
		return ds.GetLabels(), nil
	}

	return hc.workloadChecks(name, agentk8s.DaemonSet, getWorkloadLabels, false)
}

func (hc *HealthChecker) workloadChecks(
	name, kind string, getWorkloadLabels getWorkloadLabelsFn, singleton bool,
) []healthcheck.Checker {
	var labels map[string]string
	var pods []v1.Pod

	checks := []healthcheck.Checker{
		*healthcheck.NewChecker(fmt.Sprintf("%s %s exists", name, kind)).
			Fatal().
			WithCheck(func(ctx context.Context) error {
				var err error
				labels, err = getWorkloadLabels(ctx)
				if err != nil {
					return err
				}
				return checkLabel(labels, k8s.PartOfKey, k8s.PartOfVal)
			}),
		*healthcheck.NewChecker(fmt.Sprintf("%s %s is running", name, kind)).
			WithCheck(func(ctx context.Context) error {
				labelSelector := fmt.Sprintf("app=%s", name)
				podList, err := hc.k8s.Pods(ctx, labelSelector)
				if err != nil {
					return err
				}

				pods = podList.Items

				if len(pods) == 0 {
					return fmt.Errorf("no running pods for %s %s", name, kind)
				}

				return healthcheck.CheckPodsRunning(pods, "")
			}),
		*healthcheck.NewChecker(fmt.Sprintf("%s %s is injected", name, kind)).
			WithCheck(func(ctx context.Context) error {
				return healthcheck.CheckIfDataPlanePodsExist(pods)
			}),
		*healthcheck.NewChecker(fmt.Sprintf("%s %s is up-to-date", name, kind)).
			Warning().
			WithCheck(func(ctx context.Context) error {
				return checkLabel(labels, k8s.VersionLabel, hc.version)
			}),
	}

	if singleton {
		checks = append(checks,
			*healthcheck.NewChecker(fmt.Sprintf("%s %s is running a single pod", name, kind)).
				WithCheck(func(ctx context.Context) error {
					if len(pods) != 1 {
						return fmt.Errorf("expected 1 %s pod, found %d", name, len(pods))
					}
					return nil
				}),
		)
	}

	return checks
}

func checkLabel(labels map[string]string, key, val string) error {
	if l, ok := labels[key]; !ok {
		return fmt.Errorf("missing %s label", key)
	} else if l != val {
		return fmt.Errorf("incorrect %s label: %s, expected: %s", key, l, val)
	}
	return nil
}
