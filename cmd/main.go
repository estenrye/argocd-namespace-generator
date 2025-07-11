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

type HasKeyValue struct {
	Key   string  `json:"key"`
	Value *string `json:"value,omitempty"`
}

func (h HasKeyValue) HasKey(m map[string]string) bool {
	_, exists := m[h.Key]
	return exists
}

func (h HasKeyValue) HasValue(m map[string]string) bool {
	if !h.HasKey(m) {
		return false
	}
	if h.Value == nil {
		return false
	}
	value, exists := m[h.Key]
	return exists && value == *h.Value
}

func (h HasKeyValue) Matches(m map[string]string) bool {
	if h.Value == nil {
		return h.HasKey(m) // Key exists, no value to match
	}

	return h.HasValue(m)
}

type Parameters struct {
	MatchLabels        []HasKeyValue `json:"matchLabels,omitempty"`
	MatchAnnotations   []HasKeyValue `json:"matchAnnotations,omitempty"`
	ExcludeLabels      []HasKeyValue `json:"excludeLabels,omitempty"`
	ExcludeAnnotations []HasKeyValue `json:"excludeAnnotations,omitempty"`
}

func (p Parameters) Matches(ns NamespaceInfo) bool {
	for _, label := range p.MatchLabels {
		if !label.Matches(ns.Labels) {
			return false
		}
	}

	for _, annotation := range p.MatchAnnotations {
		if !annotation.Matches(ns.Annotations) {
			return false
		}
	}

	for _, label := range p.ExcludeLabels {
		if label.Matches(ns.Labels) {
			return false
		}
	}

	for _, annotation := range p.ExcludeAnnotations {
		if annotation.Matches(ns.Annotations) {
			return false
		}
	}

	return true
}

type RequestBody struct {
	Parameters Parameters `json:"parameters"`
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
	var reqBody RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil && err != io.EOF {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "invalid request body: %v"}`, err)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%v"}`, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var result []NamespaceInfo = make([]NamespaceInfo, len(namespaces))
	for _, ns := range namespaces {
		if reqBody.Parameters.Matches(ns) {
			result = append(result, ns)
		}
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%v"}`, err)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/health", HealthCheckHandler)
	r.HandleFunc("/api/v1/getparams.execute", GetParamsExecuteHandler).Methods("GET")
	r.HandleFunc("/api/v1/getparams.execute", GetParamsExecuteHandler).Methods("POST")
	http.Handle("/", r)
	http.ListenAndServe(":8080", r)
}
