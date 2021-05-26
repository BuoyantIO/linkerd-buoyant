package k8s

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	"github.com/linkerd/linkerd2/pkg/identity"
	ldConsts "github.com/linkerd/linkerd2/pkg/k8s"
	ldTls "github.com/linkerd/linkerd2/pkg/tls"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	identityComponentName        = "identity"
	linkerdNsEnvVarName          = "_l5d_ns"
	linkerdTrustDomainEnvVarName = "_l5d_trustdomain"
)

func (c *Client) GetControlPlaneCerts() (*pb.ControlPlaneCerts, error) {
	identityPod, err := c.getControlPlaneComponentPod(identityComponentName)
	if err != nil {
		return nil, err
	}

	container, err := getProxyContainer(identityPod)
	if err != nil {
		return nil, err
	}

	rootCerts, err := extractRootsCerts(container)
	if err != nil {
		return nil, err
	}

	issuerCerts, err := extractIssuerCertChain(identityPod, container)
	if err != nil {
		return nil, err
	}

	cpCerts := &pb.ControlPlaneCerts{
		IssuerCrtChain: issuerCerts,
		Roots:          rootCerts,
	}

	return cpCerts, nil
}

func (c *Client) getControlPlaneComponentPod(component string) (*v1.Pod, error) {
	selector := labels.Set(map[string]string{
		ldConsts.ControllerComponentLabel: component,
	}).AsSelector()

	pods, err := c.podLister.List(selector)
	if err != nil {
		c.log.Errorf("error listing pod: %s", err)
		return nil, err
	}

	if len(pods) == 0 {
		return nil, fmt.Errorf("could not find linkerd-%s pod", component)
	}

	for _, p := range pods {
		if p.Status.Phase == v1.PodRunning {
			return p, nil
		}
	}

	return nil, fmt.Errorf("could not find running pod for linkerd-%s", component)
}

func getProxyContainer(pod *v1.Pod) (*v1.Container, error) {
	for _, c := range pod.Spec.Containers {
		if c.Name == ldConsts.ProxyContainerName {
			container := c
			return &container, nil
		}
	}

	return nil, fmt.Errorf("could not find proxy container in pod %s/%s", pod.Namespace, pod.Name)
}

func getProxyAdminPort(container *v1.Container) (int32, error) {
	for _, p := range container.Ports {
		if p.Name == ldConsts.ProxyAdminPortName {
			return p.ContainerPort, nil
		}
	}

	return 0, fmt.Errorf("could not find port %s on proxy container [%s]", ldConsts.ProxyAdminPortName, container.Name)
}

func getServerName(podsa string, podns string, container *v1.Container) (string, error) {
	var l5dns string
	var l5dtrustdomain string
	for _, env := range container.Env {
		if env.Name == linkerdNsEnvVarName {
			l5dns = env.Value
		}
		if env.Name == linkerdTrustDomainEnvVarName {
			l5dtrustdomain = env.Value
		}
	}

	if l5dns == "" {
		return "", fmt.Errorf("could not find %s env var on proxy container [%s]", linkerdNsEnvVarName, container.Name)
	}

	if l5dtrustdomain == "" {
		return "", fmt.Errorf("could not find %s env var on proxy container [%s]", linkerdTrustDomainEnvVarName, container.Name)
	}
	return fmt.Sprintf("%s.%s.serviceaccount.identity.%s.%s", podsa, podns, l5dns, l5dtrustdomain), nil
}

func extractRootsCerts(container *v1.Container) (*pb.CertData, error) {
	for _, ev := range container.Env {
		if ev.Name == identity.EnvTrustAnchors {
			return &pb.CertData{
				Raw: []byte(ev.Value),
			}, nil
		}
	}

	return nil, fmt.Errorf("could not find env var with name %s on proxy container [%s]", identity.EnvTrustAnchors, container.Name)
}

func extractIssuerCertChain(pod *v1.Pod, container *v1.Container) (*pb.CertData, error) {
	port, err := getProxyAdminPort(container)
	if err != nil {
		return nil, err
	}

	sn, err := getServerName(pod.Spec.ServiceAccountName, pod.ObjectMeta.Namespace, container)
	if err != nil {
		return nil, err
	}

	dialer := new(net.Dialer)
	dialer.Timeout = 5 * time.Second

	conn, err := tls.DialWithDialer(
		dialer,
		"tcp",
		fmt.Sprintf("%s:%d", pod.Status.PodIP, port), &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         sn,
		})
	if err != nil {
		return nil, err
	}

	// skip the end cert
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) < 2 {
		return nil, fmt.Errorf("expected to get at least 2 peer certs, got %d", len(certs))
	}

	encodedCerts := ldTls.EncodeCertificatesPEM(certs[1:]...)
	certsData := []byte(encodedCerts)

	return &pb.CertData{
		Raw: certsData,
	}, nil
}
