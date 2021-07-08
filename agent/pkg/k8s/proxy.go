package k8s

// GetProxyLogLevel retrieves the proxy log level for a particular pod
func (c *Client) GetProxyLogLevel(podName, namespace string) (string, error) {
	return "info", nil
}

// SetProxyLogLevel sets the proxy log level for a particular pod
func (c *Client) SetProxyLogLevel(podName, namespace, logLevel string) error {
	return nil
}

// GetProxyLogs retrieves the proxy logs of a pod
func (c *Client) GetProxyLogs(podName, namespace string) (string, error) {
	return "here are some proxy logs", nil
}

// GetPrometheusScrape retrieves the raw prom scrape from the proxy of a pod
func (c *Client) GetPrometheusScrape(podName, namespace string) (string, error) {
	return "here are some prom metrics", nil
}
