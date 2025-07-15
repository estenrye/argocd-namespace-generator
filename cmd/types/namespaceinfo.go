package types

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NamespaceInfo holds information about a Kubernetes namespace.
type NamespaceInfo struct {
	Name        string            `json:"name"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
}

// ListNamespaces uses the in-cluster config to authenticate with Kubernetes and returns a list of NamespaceInfo.
func ListNamespaces() ([]NamespaceInfo, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}
	nsList, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}
	infos := make([]NamespaceInfo, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		infos = append(infos, NamespaceInfo{
			Name:        ns.Name,
			Annotations: ns.Annotations,
			Labels:      ns.Labels,
		})
	}
	return infos, nil
}

func ListRemoteClusterNamespaces() ([]NamespaceInfo, error) {

	// This function is a placeholder for future implementation.
	// It should connect to a remote cluster and list namespaces.
	return nil, fmt.Errorf("ListRemoteClusterNamespaces is not implemented yet")
}

func (ns NamespaceInfo) ToResult() map[string]string {
	result := make(map[string]string)
	result["name"] = ns.Name
	for k, v := range ns.Annotations {
		result[fmt.Sprintf("annotations-%s", k)] = v
	}
	for k, v := range ns.Labels {
		result[fmt.Sprintf("labels-%s", k)] = v
	}
	return result
}
