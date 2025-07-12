package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/estenrye/argocd-namespace-generator/cmd/endpoints"
	"github.com/estenrye/argocd-namespace-generator/cmd/types"
	"github.com/gorilla/mux"
)

// NamespaceInfo holds information about a Kubernetes namespace.

func GetParamsExecuteHandler(w http.ResponseWriter, r *http.Request) {
	namespaces, err := types.ListNamespaces()
	var reqBody types.RequestBody
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

	var result []types.NamespaceInfo = make([]types.NamespaceInfo, len(namespaces))
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
	r.HandleFunc("/healthz", endpoints.HealthCheckHandler).Methods("GET")
	r.HandleFunc("/readyz", endpoints.HealthCheckHandler).Methods("GET")
	r.HandleFunc("/api/v1/getparams.execute", GetParamsExecuteHandler).Methods("POST")
	http.Handle("/", r)
	http.ListenAndServe(":8080", r)
}
