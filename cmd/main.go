package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/estenrye/argocd-namespace-generator/cmd/endpoints"
	"github.com/estenrye/argocd-namespace-generator/cmd/types"
	"github.com/gorilla/mux"
	"github.com/purini-to/zapmw"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NamespaceInfo holds information about a Kubernetes namespace.

func GetParamsExecuteHandler(w http.ResponseWriter, r *http.Request) {
	logger := zapmw.GetZap(r)

	logger.Sugar().Debug("Request Headers:", r.Header, "Query Args:", r.URL.Query(), "Body:", r.Body)

	var reqBody types.RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil && err != io.EOF {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "invalid request body: %v"}`, err)
		logger.Error("invalid request body", zap.Error(err))
		return
	}

	namespaces, err := types.ListNamespaces()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%v"}`, err)
		logger.Error("failed to list namespaces", zap.Error(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var responseBody types.ResponseBody = types.ResponseBody{
		Parameters: make([]types.NamespaceInfo, len(namespaces)),
	}

	for _, ns := range namespaces {
		if reqBody.Parameters.Matches(ns) {
			responseBody.Parameters = append(responseBody.Parameters, ns)
		}
	}

	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%v"}`, err)
	}
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	r.Use(
		zapmw.WithZap(logger),
		zapmw.Request(zapcore.InfoLevel, "request"),
		zapmw.Recoverer(zapcore.ErrorLevel, "recover", zapmw.RecovererDefault),
	)

	r.HandleFunc("/healthz", endpoints.HealthCheckHandler).Methods("GET")
	r.HandleFunc("/readyz", endpoints.HealthCheckHandler).Methods("GET")
	r.HandleFunc("/api/v1/getparams.execute", GetParamsExecuteHandler).Methods("POST")
	http.Handle("/", r)
	http.ListenAndServe(":8080", r)
}
