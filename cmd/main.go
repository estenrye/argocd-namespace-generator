package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// A very simple health check.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	io.WriteString(w, `{"alive": true}`)
}

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

func GetParamsExecuteHandler(w http.ResponseWriter, r *http.Request) {
	namespaces, err := ListNamespaces()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%v"}`, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(namespaces); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%v"}`, err)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/health", HealthCheckHandler)
	r.HandleFunc("/api/v1/getparams.execute", GetParamsExecuteHandler).Methods("GET")
	http.Handle("/", r)
	http.ListenAndServe(":8080", r)
}
