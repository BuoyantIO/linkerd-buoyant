package k8s

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	"github.com/linkerd/linkerd2/pkg/identity"
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	ldTls "github.com/linkerd/linkerd2/pkg/tls"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	identityComponentName        = "identity"
	linkerdNsEnvVarName          = "_l5d_ns"
	linkerdTrustDomainEnvVarName = "_l5d_trustdomain"
	trustRootsConfigMapName      = "linkerd-identity-trust-roots"
	trustRootsConfigMapKeyName   = "ca-bundle.crt"
)

func (c *Client) GetControlPlaneCerts(ctx context.Context) (*pb.ControlPlaneCerts, error) {
	identityPod, err := c.getControlPlaneComponentPod(identityComponentName)
	if err != nil {
		return nil, err
	}

	container, err := getProxyContainer(identityPod)
	if err != nil {
		return nil, err
	}

	rootCerts, err := c.extractRootsCerts(ctx, container, identityPod.Namespace)
	if err != nil {
		return nil, err
	}

	issuerCerts, err := c.extractIssuerCertChain(identityPod, container)
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
		l5dk8s.ControllerComponentLabel: component,
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

func (c *Client) extractRootsCerts(ctx context.Context, container *v1.Container, namespace string) ([]*pb.CertData, error) {
	for _, ev := range container.Env {
		if ev.Name != identity.EnvTrustAnchors {
			continue
		}

		var certificates []*x509.Certificate
		var err error
		if ev.Value != "" {
			certificates, err = ldTls.DecodePEMCertificates(ev.Value)
			if err != nil {
				return nil, err
			}
		} else {
			// in this case we need to check for a reference to a config map
			if ev.ValueFrom == nil || ev.ValueFrom.ConfigMapKeyRef == nil {
				return nil, fmt.Errorf("neither a Value nor a ConfigMapKeyRef for the %s env var arepresent on proxy container [%s]", identity.EnvTrustAnchors, container.Name)
			}
			cmName := ev.ValueFrom.ConfigMapKeyRef.Name
			cm, err := c.k8sClient.CoreV1().ConfigMaps(namespace).Get(ctx, cmName, metav1.GetOptions{})
			if err != nil {
				return nil, fmt.Errorf("cannot obtain config map %s/%s", namespace, cmName)
			}

			rootsValue, ok := cm.Data[trustRootsConfigMapKeyName]
			if !ok {
				return nil, fmt.Errorf("config map %s/%s does not have %s key", namespace, cmName, trustRootsConfigMapKeyName)
			}

			certificates, err = ldTls.DecodePEMCertificates(rootsValue)
			if err != nil {
				return nil, err
			}
		}

		certsData := make([]*pb.CertData, len(certificates))
		for i, crt := range certificates {
			encoded := ldTls.EncodeCertificatesPEM(crt)
			certsData[i] = &pb.CertData{Raw: []byte(encoded)}
		}

		return certsData, nil
	}

	return nil, fmt.Errorf("could not find env var with name %s on proxy container [%s]", identity.EnvTrustAnchors, container.Name)
}

func (c *Client) extractIssuerCertChain(pod *v1.Pod, container *v1.Container) ([]*pb.CertData, error) {
	sn, err := getServerName(pod.Spec.ServiceAccountName, pod.ObjectMeta.Namespace, container)
	if err != nil {
		return nil, err
	}

	proxyConnection, err := c.getContainerConnection(pod, container, l5dk8s.ProxyAdminPortName)
	if err != nil {
		return nil, err
	}
	defer proxyConnection.cleanup()

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 5 * time.Second},
		"tcp",
		proxyConnection.host,
		&tls.Config{
			// we want to subvert TLS verification as we do not need
			// to verify that we actually trust these certs. We just
			// want the certificates and are not sending any data here.
			// Therefore `InsecureSkipVerify` is just fine. An added
			// benefit is that we save on some CPU cycles that would be
			// wasted doing TLS cert verification
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

	certificates := certs[1:]
	certsData := make([]*pb.CertData, len(certificates))
	for i, crt := range certificates {
		encoded := ldTls.EncodeCertificatesPEM(crt)
		certsData[i] = &pb.CertData{Raw: []byte(encoded)}
	}

	return certsData, nil
}
