package k8s

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetProxyLogsAndLevel retrieves the the proxy logs and level of a pod
func (c *Client) GetProxyLogsAndLevel(ctx context.Context, podName, namespace string, lines int64) ([]byte, string, error) {
	logs, err := c.GetProxyLogs(ctx, podName, namespace, true, &lines)
	if err != nil {
		return nil, "", err
	}

	pod, err := c.k8sClient.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, "", err
	}

	if pod.Status.Phase != v1.PodRunning {
		return nil, "", fmt.Errorf("pod not running: %s/%s", namespace, podName)
	}

	if pod.Status.PodIP == "" {
		return nil, "", fmt.Errorf("pod IP not allocated: %s/%s", namespace, podName)
	}

	proxyContainer, err := getProxyContainer(pod)
	if err != nil {
		return nil, "", err
	}

	proxyConnection, err := c.getContainerConnection(pod, proxyContainer, l5dk8s.ProxyAdminPortName)
	if err != nil {
		return nil, "", err
	}
	defer proxyConnection.cleanup()

	metricsUrl := fmt.Sprintf("http://%s/proxy-log-level", proxyConnection.host)
	resp, err := http.Get(metricsUrl)
	if err != nil {
		return nil, "", err
	}

	level, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	resp.Body.Close()

	return logs, string(level), nil
}
