package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// TestConfig holds test configuration
type TestConfig struct {
	Namespace       string
	KubeConfig      string
	FrontendURL     string
	IdentityURL     string
	LedgerURL       string
	PaymentURL      string
	ProductURL      string
	CardURL         string
	IstioGatewayURL string
	Timeout         time.Duration
}

var (
	testConfig *TestConfig
	k8sClient  *kubernetes.Clientset
	httpClient *http.Client
)

func init() {
	testConfig = &TestConfig{
		Namespace:       getEnv("TEST_NAMESPACE", "neobank"),
		KubeConfig:      getEnv("KUBECONFIG", filepath.Join(os.Getenv("HOME"), ".kube", "config")),
		FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:3001"),
		IdentityURL:     getEnv("IDENTITY_URL", "http://localhost:8081"),
		LedgerURL:       getEnv("LEDGER_URL", "http://localhost:8082"),
		PaymentURL:      getEnv("PAYMENT_URL", "http://localhost:8083"),
		ProductURL:      getEnv("PRODUCT_URL", "http://localhost:8084"),
		CardURL:         getEnv("CARD_URL", "http://localhost:8085"),
		IstioGatewayURL: getEnv("ISTIO_GATEWAY_URL", "http://localhost:8080"),
		Timeout:         30 * time.Second,
	}

	httpClient = &http.Client{
		Timeout: testConfig.Timeout,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupK8sClient initializes the Kubernetes client
func SetupK8sClient(t *testing.T) *kubernetes.Clientset {
	if k8sClient != nil {
		return k8sClient
	}

	config, err := clientcmd.BuildConfigFromFlags("", testConfig.KubeConfig)
	if err != nil {
		t.Skipf("Cannot build kubeconfig: %v", err)
		return nil
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Skipf("Cannot create k8s client: %v", err)
		return nil
	}

	k8sClient = client
	return client
}

// WaitForPodReady waits for a pod to be ready
func WaitForPodReady(ctx context.Context, client *kubernetes.Clientset, namespace, labelSelector string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			return err
		}

		allReady := true
		for _, pod := range pods.Items {
			if pod.Status.Phase != "Running" {
				allReady = false
				break
			}
			for _, cond := range pod.Status.Conditions {
				if cond.Type == "Ready" && cond.Status != "True" {
					allReady = false
					break
				}
			}
		}

		if allReady && len(pods.Items) > 0 {
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for pods with selector %s", labelSelector)
}

// HealthCheck performs a health check on a service
func HealthCheck(url string) error {
	resp, err := httpClient.Get(url + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetPodLogs retrieves logs from a pod
func GetPodLogs(ctx context.Context, client *kubernetes.Clientset, namespace, podName string, lines int64) (string, error) {
	req := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		TailLines: &lines,
	})

	logs, err := req.Do(ctx).Raw()
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

// Basic test to verify test framework is working
func TestFrameworkSetup(t *testing.T) {
	t.Log("E2E Test Framework initialized")
	t.Logf("Test Configuration:")
	t.Logf("  Namespace: %s", testConfig.Namespace)
	t.Logf("  Frontend URL: %s", testConfig.FrontendURL)
	t.Logf("  Identity URL: %s", testConfig.IdentityURL)
	t.Logf("  Istio Gateway URL: %s", testConfig.IstioGatewayURL)
}
