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

	var reqBody types.RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil && err != io.EOF {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "invalid request body: %v"}`, err)
		logger.Error("invalid request body", zap.Error(err))
		return
	}

	logger.Debug("received request to get parameters", zap.Any("requestBody", reqBody))
	namespaces, err := types.ListNamespaces(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%v"}`, err)
		logger.Error("failed to list namespaces", zap.Error(err))
		return
	}

	logger.Debug("listed local namespaces", zap.Int("count", len(namespaces)), zap.Any("namespaces", namespaces))

	remoteNamespaces, err := types.ListRemoteClusterNamespaces(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%v"}`, err)
		logger.Error("failed to list remote cluster namespaces", zap.Error(err))
		return
	}

	logger.Debug("listed remote cluster namespaces", zap.Int("count", len(remoteNamespaces)), zap.Any("remoteNamespaces", remoteNamespaces))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var responseBody types.ResponseBody = types.ResponseBody{
		Output: types.Output{
			Parameters: make([]map[string]string, 0, len(namespaces)+len(remoteNamespaces)),
		},
	}

	for _, ns := range namespaces {
		if reqBody.Input.Parameters.Matches(ns) {
			logger.Debug("local namespace matches request", zap.String("namespace", ns.Name), zap.Any("request", reqBody.Input.Parameters))
			responseBody.Output.Parameters = append(responseBody.Output.Parameters, ns.ToResult())
		}
	}

	for _, ns := range remoteNamespaces {
		if reqBody.Input.Parameters.Matches(ns) {
			logger.Debug("remote namespace matches request", zap.String("namespace", ns.Name), zap.String("serverHost", ns.ServerHost), zap.Any("request", reqBody.Input.Parameters))
			responseBody.Output.Parameters = append(responseBody.Output.Parameters, ns.ToResult())
		}
	}

	logger.Info("returning response", zap.Any("requestBody", reqBody), zap.Any("responseBody", responseBody))
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
