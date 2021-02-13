package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// New takes a kubeconfig and kubecontext and returns a new Kubernetes
// clientset, which satisfies kubernetes.Interface, along with the current
// context.
func New(kubeconfig string, kubecontext string) (kubernetes.Interface, string, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{CurrentContext: kubecontext})

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, "", err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, "", err
	}

	r, err := clientConfig.RawConfig()
	if err != nil {
		return nil, "", err
	}

	return clientset, r.CurrentContext, nil
}
