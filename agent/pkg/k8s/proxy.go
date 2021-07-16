package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	ld5k8s "github.com/linkerd/linkerd2/pkg/k8s"
	ldConsts "github.com/linkerd/linkerd2/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const linkerdNamespace = "linkerd"
const k8sServiceName = "kubernetes"

// GetProxyLogs retrieves the proxy logs of a pod
func (c *Client) GetProxyLogs(ctx context.Context, podName, namespace string) ([]byte, error) {
	req := c.k8sClient.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{Container: ldConsts.ProxyContainerName})
	logs, err := req.Stream(ctx)
	if err != nil {
		return nil, err
	}
	defer logs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logs)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GetPrometheusScrape retrieves the raw prom scrape from the proxy of a pod
func (c *Client) GetPrometheusScrape(ctx context.Context, podName, namespace string) ([][]byte, error) {
	// first get the pod
	pod, err := c.k8sClient.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if pod.Status.Phase != v1.PodRunning {
		return nil, fmt.Errorf("pod not running: %s/%s", namespace, podName)
	}

	if pod.Status.PodIP == "" {
		return nil, fmt.Errorf("pod IP not allocated: %s/%s", namespace, podName)
	}

	var proxyContainer *v1.Container
	for _, c := range pod.Spec.Containers {
		if c.Name == ldConsts.ProxyContainerName {
			c := c
			proxyContainer = &c
			break
		}
	}
	if proxyContainer == nil {
		return nil, fmt.Errorf("cannot find proxy container for pod: %s/%s", namespace, podName)
	}

	metricsUrl := ""
	if c.ld5API != nil {
		pf, err := ld5k8s.NewContainerMetricsForward(c.ld5API, *pod, *proxyContainer, false, ldConsts.ProxyAdminPortName)
		if err != nil {
			return nil, err
		}
		metricsUrl = pf.URLFor("/metrics")
		if err = pf.Init(); err != nil {
			return nil, err
		}

		defer pf.Stop()
	} else {
		// now find the port we need to hit
		var port *int32
		for _, p := range proxyContainer.Ports {
			if p.Name == ldConsts.ProxyAdminPortName {
				p := p
				port = &p.ContainerPort
				break
			}
		}
		if port == nil {
			return nil, fmt.Errorf("cannot find proxy admin port for pod: %s/%s", namespace, podName)
		}
		metricsUrl = fmt.Sprintf("http://%s:%d/metrics", pod.Status.PodIP, *port)
	}

	data := [][]byte{}
	for {
		resp, err := http.Get(metricsUrl)
		if err != nil {
			return nil, err
		}

		sample, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()
		data = append(data, sample)
		if len(data) == 6 {
			break
		}
		time.Sleep(time.Second * 10)
	}

	return data, nil
}

// GetPodManifest retrieves pod manifest
func (c *Client) GetPodSpec(ctx context.Context, podName, namespace string) (*pb.Pod, error) {
	pod, err := c.k8sClient.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &pb.Pod{Pod: c.serialize(pod, v1.SchemeGroupVersion)}, nil
}

// GetLinkerdConfigMap retrieves Linkerd config map
func (c *Client) GetLinkerdConfigMap(ctx context.Context) (*pb.ConfigMap, error) {
	cm, err := c.k8sClient.CoreV1().ConfigMaps(linkerdNamespace).Get(ctx, ldConsts.ConfigConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &pb.ConfigMap{ConfigMap: c.serialize(cm, v1.SchemeGroupVersion)}, nil
}

// GetNodeManifests retrieves all nodes in the cluster
func (c *Client) GetNodeManifests(ctx context.Context) ([]*pb.Node, error) {
	nodes, err := c.k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	data := make([]*pb.Node, len(nodes.Items))
	for i, node := range nodes.Items {
		data[i] = &pb.Node{Node: c.serialize(&node, v1.SchemeGroupVersion)}
	}

	return data, nil
}

// GetK8sServiceManifest the manifest of the kubernetes service residing in the default namespace
func (c *Client) GetK8sServiceManifest(ctx context.Context) (*pb.Service, error) {
	svc, err := c.k8sClient.CoreV1().Services(v1.NamespaceDefault).Get(ctx, k8sServiceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &pb.Service{Service: c.serialize(svc, v1.SchemeGroupVersion)}, nil
}
