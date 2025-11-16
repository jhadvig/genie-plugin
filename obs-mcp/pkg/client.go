package pkg

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClientConfig returns a Kubernetes REST config using kubeconfig
func GetClientConfig() (*rest.Config, error) {
	// Try to load from kubeconfig first
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	return config, nil
}

// GetKubeClient returns a Kubernetes client
func GetKubeClient() (*kubernetes.Clientset, error) {
	config, err := GetClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return clientset, nil
}

// GetThanosQuerierURL discovers the Thanos Querier service URL in OpenShift
func GetThanosQuerierURL() (string, error) {
	ctx := context.Background()

	kubeClient, err := GetKubeClient()
	if err != nil {
		return "", fmt.Errorf("failed to get kubernetes client: %w", err)
	}

	restClient := kubeClient.CoreV1().RESTClient()
	result := restClient.Get().
		AbsPath("/apis/route.openshift.io/v1").
		Namespace("openshift-monitoring").
		Resource("routes").
		Name("thanos-querier").
		Do(ctx)

	if result.Error() != nil {
		return "", fmt.Errorf("failed to load thanos-querier route: %w", result.Error())
	}

	body, err := result.Raw()
	if err != nil {
		return "", fmt.Errorf("failed to parse the route results: %w", err)
	}

	// Simple string parsing to extract the host
	bodyStr := string(body)
	if strings.Contains(bodyStr, `"host":`) {
		// Extract host field using string manipulation
		parts := strings.Split(bodyStr, `"host":"`)
		if len(parts) > 1 {
			hostPart := strings.Split(parts[1], `"`)[0]
			if hostPart != "" {
				return "https://" + hostPart, nil
			}
		}
	}

	return "", fmt.Errorf("no suitable route found for thanos-querier")
}

// GetPrometheusURL discovers the Prometheus service URL in a kind cluster
// For dev/testing, this returns a localhost URL that works with kubectl port-forward
func GetPrometheusURL() (string, error) {
	ctx := context.Background()

	kubeClient, err := GetKubeClient()
	if err != nil {
		return "", fmt.Errorf("failed to get kubernetes client: %w", err)
	}

	foundService, err := kubeClient.CoreV1().Services(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=prometheus,app.kubernetes.io/component=server",
	})
	if err != nil {
		return "", fmt.Errorf("failed to list services: %w", err)
	}
	if len(foundService.Items) != 1 {
		return "", fmt.Errorf("expected 1 service, got %d", len(foundService.Items))
	}

	svc := foundService.Items[0]
	var port string
	if len(svc.Spec.Ports) > 0 {
		port = svc.Spec.Ports[0].TargetPort.String()
	}

	// Check if port-forward is set up in prometheusNs
	conn, err := net.DialTimeout("tcp", ":"+port, time.Second*5)
	if err == nil {
		conn.Close()
		return "http://localhost:" + port, nil
	}

	return "", fmt.Errorf("no Prometheus service found: %w", err)
}
