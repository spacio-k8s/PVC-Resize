package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GetK8sClientWithConfig() (*kubernetes.Clientset, *rest.Config, error) {
	var config *rest.Config
	var err error

	// in-cluster
	config, err = rest.InClusterConfig()
	if err != nil {
		// fallback to kubeconfig
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	return clientset, config, nil
}

// GetKubeConfig returns a Kubernetes REST config.
// It tries in-cluster config first, then falls back to ~/.kube/config
func GetKubeConfig() (*rest.Config, error) {
	// Try in-cluster config
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fallback to kubeconfig file
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		home := homedir.HomeDir()
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetK8sClient returns a Kubernetes clientset.
// It tries in-cluster config first; if not available, falls back to local kubeconfig.
func GetK8sClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			config, err = clientcmd.BuildConfigFromFlags("", os.ExpandEnv("$HOME/.kube/config"))
		}
	}
	if err != nil {

		return nil, fmt.Errorf("Failed to get kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Failed to create clientset: %v", err)
	}

	return clientset, nil
}
func GetClusterName() string {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err != nil {
		// fallback to kubeconfig from HOME
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return "UnknownCluster"
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "UnknownCluster"
	}

	versionInfo, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return "UnknownCluster"
	}

	return fmt.Sprintf("K8sCluster-%s.%s", versionInfo.Major, versionInfo.Minor)
}

// ListNamespaces returns all namespace names in the cluster
func ListNamespaces() ([]string, error) {
	clientset, err := GetK8sClient()
	if err != nil {
		return nil, err
	}
	nsList, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	namespaces := make([]string, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		namespaces = append(namespaces, ns.Name)
	}
	return namespaces, nil
}
