package k8s

import v1 "k8s.io/api/core/v1"

// GetProxyLogs retrieves the proxy logs of a pod
func (c *Client) GetProxyLogs(podName, namespace string) (string, error) {
	return "here are some proxy logs", nil
}

// GetPrometheusScrape retrieves the raw prom scrape from the proxy of a pod
func (c *Client) GetPrometheusScrape(podName, namespace string) (string, error) {
	return "here are some prom metrics", nil
}

// GetPodManifest retrieves pod manifest
func (c *Client) GetPodManifest(podName, namespace string) (*v1.Pod, error) {
	return &v1.Pod{}, nil
}
