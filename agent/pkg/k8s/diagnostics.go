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
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const linkerdNamespace = "linkerd"
const k8sServiceName = "kubernetes"

// GetProxyLogs retrieves the proxy logs of a pod
func (c *Client) GetProxyLogs(ctx context.Context, podName, namespace string, includeTimestamps bool, tailLines *int64) ([]byte, error) {
	podLogOptions := &v1.PodLogOptions{Container: l5dk8s.ProxyContainerName, Timestamps: includeTimestamps, TailLines: tailLines}
	req := c.l5dApi.CoreV1().Pods(namespace).GetLogs(podName, podLogOptions)
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
	pod, err := c.l5dApi.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if pod.Status.Phase != v1.PodRunning {
		return nil, fmt.Errorf("pod not running: %s/%s", namespace, podName)
	}

	if pod.Status.PodIP == "" {
		return nil, fmt.Errorf("pod IP not allocated: %s/%s", namespace, podName)
	}

	proxyContainer, err := getProxyContainer(pod)
	if err != nil {
		return nil, err
	}

	proxyConnection, err := c.getContainerConnection(pod, proxyContainer, l5dk8s.ProxyAdminPortName)
	if err != nil {
		return nil, err
	}
	defer proxyConnection.cleanup()

	metricsUrl := fmt.Sprintf("http://%s/metrics", proxyConnection.host)
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

// GetPodSpec retrieves pod manifest
func (c *Client) GetPodSpec(ctx context.Context, podName, namespace string) (*pb.Pod, error) {
	pod, err := c.l5dApi.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &pb.Pod{Pod: c.serialize(pod, v1.SchemeGroupVersion)}, nil
}

// GetLinkerdConfigMap retrieves Linkerd config map
func (c *Client) GetLinkerdConfigMap(ctx context.Context) (*pb.ConfigMap, error) {
	cm, err := c.l5dApi.CoreV1().ConfigMaps(linkerdNamespace).Get(ctx, l5dk8s.ConfigConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &pb.ConfigMap{ConfigMap: c.serialize(cm, v1.SchemeGroupVersion)}, nil
}

// GetNodeManifests retrieves all nodes in the cluster
func (c *Client) GetNodeManifests(ctx context.Context) ([]*pb.Node, error) {
	nodes, err := c.l5dApi.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
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
	svc, err := c.l5dApi.CoreV1().Services(v1.NamespaceDefault).Get(ctx, k8sServiceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &pb.Service{Service: c.serialize(svc, v1.SchemeGroupVersion)}, nil
}
