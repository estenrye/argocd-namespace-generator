package types

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/purini-to/zapmw"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NamespaceInfo holds information about a Kubernetes namespace.
type NamespaceInfo struct {
	Name         string            `json:"name"`
	Annotations  map[string]string `json:"annotations"`
	Labels       map[string]string `json:"labels"`
	ServerHost   string            `json:"serverHost"`
	ServerName   string            `json:"serverName"`
	ServerLabels map[string]string `json:"serverLabels,omitempty"`
}

// ListNamespaces uses the in-cluster config to authenticate with Kubernetes and returns a list of NamespaceInfo.
func ListNamespaces(r *http.Request) ([]NamespaceInfo, error) {
	logger := zapmw.GetZap(r)

	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Error("failed to load in-cluster config", zap.Error(err))
		return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error("failed to create clientset", zap.Error(err))
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}
	nsList, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Error("failed to list namespaces", zap.Error(err))
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}
	infos := make([]NamespaceInfo, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		infos = append(infos, NamespaceInfo{
			Name:        ns.Name,
			Annotations: ns.Annotations,
			Labels:      ns.Labels,
			ServerHost:  "https://kubernetes.default.svc",
			ServerName:  "in-cluster",
		})
	}
	return infos, nil
}

func ListRemoteClusterNamespaces(r *http.Request) ([]NamespaceInfo, error) {
	// This function will get all of the secrets in the argocd
	// namespace with the label argocd.argoproj.io/secret-type=cluster.
	// It will then use the kubeconfig in each secret to connect to the remote cluster
	// and list the namespaces in that cluster.
	logger := zapmw.GetZap(r)
	logger.Debug("listing remote cluster namespaces")
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Error("failed to load in-cluster config", zap.Error(err))
		return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error("failed to create clientset", zap.Error(err))
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	secrets, err := clientset.CoreV1().Secrets("argocd").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "argocd.argoproj.io/secret-type=cluster",
	})
	if err != nil {
		logger.Error("failed to list cluster secrets", zap.Error(err))
		return nil, fmt.Errorf("failed to list cluster secrets: %w", err)
	}

	logger.Info("found cluster secrets", zap.Int("count", len(secrets.Items)))
	var allNamespaces []NamespaceInfo
	for _, secret := range secrets.Items {
		// Decode the config data from the secret
		logger.Info("processing secret", zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))

		logger.Debug("retrieving config", zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))
		configData := secret.Data["config"]
		if configData == nil {
			continue
		}

		logger.Debug("retrieving server name", zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))
		serverName := secret.Data["name"]
		if serverName == nil {
			continue
		}

		logger.Debug("retrieving server host", zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))
		serverHost := secret.Data["server"]
		if serverHost == nil {
			continue
		}

		serverLabels := secret.Labels

		// Parse the JSON config data
		logger.Debug("unmarshalling cluster config", zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))
		var clusterData map[string]interface{}
		if err := json.Unmarshal(configData, &clusterData); err != nil {
			logger.Error("failed to unmarshal cluster config", zap.Error(err))
			continue
		}

		// Extract the bearer token from the cluster secret
		logger.Debug("retrieving bearer token", zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))
		bearerToken, ok := clusterData["bearerToken"].(string)
		if !ok {
			continue
		}

		// Extract the TLS client config
		logger.Debug("retrieving TLS client config", zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))
		tlsClientConfig := make(map[string]interface{})
		if caData, ok := clusterData["tlsClientConfig"].(map[string]interface{}); ok {
			if insecure, exists := caData["insecure"].(bool); exists {
				tlsClientConfig["insecure"] = insecure
			}
			if certData, exists := caData["caData"].(string); exists {
				tlsClientConfig["certificate-authority-data-64"] = certData
				decodedCAData, err := base64.StdEncoding.DecodeString(certData)
				if err != nil {
					logger.Error("failed to decode CA data", zap.Error(err), zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))
					continue
				}
				tlsClientConfig["caData"] = decodedCAData
			}
		}

		logger.Debug("certificate authority data", zap.Any("tlsClientConfig", tlsClientConfig), zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))

		logger.Debug("creating remote cluster config", zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))
		remoteClusterConfig := &rest.Config{
			Host:        string(serverHost),
			BearerToken: bearerToken,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: tlsClientConfig["insecure"].(bool),
				CAData:   tlsClientConfig["caData"].([]byte),
			},
		}

		logger.Debug("creating remote client", zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))
		remoteClient, err := kubernetes.NewForConfig(remoteClusterConfig)
		if err != nil {
			logger.Error("failed to create remote client", zap.Error(err), zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))
			continue
		}

		logger.Debug("listing namespaces in remote cluster", zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))
		nsList, err := remoteClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logger.Error("failed to list namespaces in remote cluster", zap.Error(err), zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))
			continue
		}

		logger.Debug("found namespaces in remote cluster", zap.Int("count", len(nsList.Items)), zap.String("name", secret.Name), zap.String("namespace", secret.Namespace))
		for _, ns := range nsList.Items {
			allNamespaces = append(allNamespaces, NamespaceInfo{
				Name:         ns.Name,
				Annotations:  ns.Annotations,
				Labels:       ns.Labels,
				ServerHost:   remoteClusterConfig.Host,
				ServerName:   string(serverName),
				ServerLabels: serverLabels,
			})
		}
	}
	return allNamespaces, nil
}

func (ns NamespaceInfo) ToResult() map[string]string {
	result := make(map[string]string)
	result["name"] = ns.Name
	result["serverHost"] = ns.ServerHost
	result["serverName"] = ns.ServerName
	for k, v := range ns.ServerLabels {
		result[fmt.Sprintf("serverLabels-%s", k)] = v
	}
	for k, v := range ns.Annotations {
		result[fmt.Sprintf("annotations-%s", k)] = v
	}
	for k, v := range ns.Labels {
		result[fmt.Sprintf("labels-%s", k)] = v
	}
	return result
}
