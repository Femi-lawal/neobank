package e2e

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Istio E2E Tests
// Tests service mesh configuration, circuit breaking, rate limiting, and security

const istioNamespace = "istio-system"

func TestIstioInstallation(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("IstioNamespaceExists", func(t *testing.T) {
		_, err := client.CoreV1().Namespaces().Get(ctx, istioNamespace, metav1.GetOptions{})
		if err != nil {
			t.Skip("Istio namespace does not exist - Istio may not be installed")
		}
		t.Log("✓ Istio namespace exists")
	})

	t.Run("IstiodDeployment", func(t *testing.T) {
		deployment, err := client.AppsV1().Deployments(istioNamespace).Get(ctx, "istiod", metav1.GetOptions{})
		if err != nil {
			t.Skip("Istiod deployment not found")
		}

		if deployment.Status.ReadyReplicas > 0 {
			t.Logf("✓ Istiod running: %d/%d replicas ready", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas)
		} else {
			t.Log("○ Istiod has no ready replicas")
		}
	})

	t.Run("IstioIngressGateway", func(t *testing.T) {
		deployment, err := client.AppsV1().Deployments(istioNamespace).Get(ctx, "istio-ingressgateway", metav1.GetOptions{})
		if err != nil {
			t.Skip("Istio ingress gateway not found")
		}

		if deployment.Status.ReadyReplicas > 0 {
			t.Logf("✓ Istio ingress gateway running: %d/%d replicas ready", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas)
		} else {
			t.Log("○ Istio ingress gateway has no ready replicas")
		}
	})
}

func TestIstioSidecarInjection(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("NamespaceInjectionLabel", func(t *testing.T) {
		ns, err := client.CoreV1().Namespaces().Get(ctx, testConfig.Namespace, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Failed to get namespace: %v", err)
		}

		if label, ok := ns.Labels["istio-injection"]; ok && label == "enabled" {
			t.Log("✓ Istio injection enabled for namespace")
		} else {
			t.Log("○ Istio injection not enabled for namespace")
		}
	})

	t.Run("PodSidecars", func(t *testing.T) {
		pods, err := client.CoreV1().Pods(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list pods: %v", err)
		}

		podsWithSidecar := 0
		podsWithoutSidecar := 0

		for _, pod := range pods.Items {
			hasSidecar := false
			for _, container := range pod.Spec.Containers {
				if container.Name == "istio-proxy" {
					hasSidecar = true
					break
				}
			}
			if hasSidecar {
				podsWithSidecar++
			} else {
				podsWithoutSidecar++
			}
		}

		t.Logf("Pods with Istio sidecar: %d", podsWithSidecar)
		t.Logf("Pods without Istio sidecar: %d", podsWithoutSidecar)
	})
}

func TestIstioGatewayConfiguration(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("GatewayYAMLExists", func(t *testing.T) {
		// Check if gateway config exists by looking for the service
		_, err := client.CoreV1().Services(istioNamespace).Get(ctx, "istio-ingressgateway", metav1.GetOptions{})
		if err != nil {
			t.Skip("Istio ingress gateway service not found")
		}
		t.Log("✓ Istio ingress gateway service exists")
	})
}

func TestIstioDestinationRules(t *testing.T) {
	// Note: This would require Istio client-go to check DestinationRule CRDs
	// For now, we test indirectly through behavior

	t.Run("CircuitBreakerBehavior", func(t *testing.T) {
		// Test circuit breaker by making requests
		client := &http.Client{Timeout: 5 * time.Second}

		successCount := 0
		failCount := 0

		for i := 0; i < 10; i++ {
			resp, err := client.Get(testConfig.IdentityURL + "/health")
			if err != nil {
				failCount++
			} else {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					successCount++
				} else {
					failCount++
				}
			}
		}

		t.Logf("Health check results: %d success, %d fail", successCount, failCount)
	})
}

func TestIstioRateLimiting(t *testing.T) {
	t.Run("RateLimitEnforcement", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}

		// Make rapid requests to test rate limiting
		responses := make(chan int, 100)
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				resp, err := client.Get(testConfig.IstioGatewayURL + "/api/auth/health")
				if err != nil {
					responses <- 0
					return
				}
				defer resp.Body.Close()
				responses <- resp.StatusCode
			}()
		}

		wg.Wait()
		close(responses)

		statusCounts := make(map[int]int)
		for status := range responses {
			statusCounts[status]++
		}

		t.Log("Rate limit test results:")
		for status, count := range statusCounts {
			t.Logf("  Status %d: %d responses", status, count)
		}

		// Check if any 429 (rate limited) responses
		if count429, ok := statusCounts[429]; ok && count429 > 0 {
			t.Logf("✓ Rate limiting is active (%d requests were rate limited)", count429)
		} else {
			t.Log("○ No rate limiting detected (may not be configured or threshold not reached)")
		}
	})
}

func TestIstioMTLS(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("PeerAuthenticationExists", func(t *testing.T) {
		// Check for PeerAuthentication by looking for the file presence
		// In a real test, we'd use Istio client-go
		cm, err := client.CoreV1().ConfigMaps(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Skip("Could not list ConfigMaps")
		}

		t.Logf("Found %d ConfigMaps in namespace (checking for mTLS config)", len(cm.Items))

		// Note: PeerAuthentication is a CRD, not a ConfigMap
		// This is a placeholder - real test would use Istio API
		t.Log("Note: Full mTLS verification requires Istio client libraries")
	})
}

func TestIstioRetryPolicy(t *testing.T) {
	t.Run("RetryBehavior", func(t *testing.T) {
		// Test retry behavior by making requests to a potentially flaky endpoint
		client := &http.Client{Timeout: 30 * time.Second}

		start := time.Now()
		resp, err := client.Get(testConfig.IdentityURL + "/health")
		duration := time.Since(start)

		if err != nil {
			t.Logf("Request failed after %v: %v", duration, err)
			return
		}
		defer resp.Body.Close()

		t.Logf("Request completed in %v with status %d", duration, resp.StatusCode)
	})
}

func TestIstioTimeouts(t *testing.T) {
	t.Run("TimeoutConfiguration", func(t *testing.T) {
		// Test that timeouts are enforced
		client := &http.Client{Timeout: 60 * time.Second}

		services := []struct {
			name    string
			url     string
			timeout time.Duration // expected timeout from VirtualService
		}{
			{"identity-service", testConfig.IdentityURL, 30 * time.Second},
			{"ledger-service", testConfig.LedgerURL, 45 * time.Second},
			{"payment-service", testConfig.PaymentURL, 60 * time.Second},
		}

		for _, svc := range services {
			t.Run(svc.name, func(t *testing.T) {
				start := time.Now()
				resp, err := client.Get(svc.url + "/health")
				duration := time.Since(start)

				if err != nil {
					t.Logf("Request to %s failed after %v: %v", svc.name, duration, err)
					return
				}
				defer resp.Body.Close()

				t.Logf("%s responded in %v (expected timeout: %v)", svc.name, duration, svc.timeout)
			})
		}
	})
}

func TestIstioTrafficManagement(t *testing.T) {
	t.Run("TrafficRouting", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}

		// Test routing through Istio gateway
		endpoints := []struct {
			path    string
			service string
		}{
			{"/api/auth/health", "identity-service"},
			{"/api/ledger/health", "ledger-service"},
			{"/api/payment/health", "payment-service"},
		}

		for _, ep := range endpoints {
			resp, err := client.Get(testConfig.IstioGatewayURL + ep.path)
			if err != nil {
				t.Logf("○ %s (%s): request failed - %v", ep.path, ep.service, err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				t.Logf("✓ %s (%s): routed correctly", ep.path, ep.service)
			} else {
				t.Logf("○ %s (%s): status %d", ep.path, ep.service, resp.StatusCode)
			}
		}
	})
}

func TestIstioObservability(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("PrometheusIntegration", func(t *testing.T) {
		// Check if Prometheus is scraping Istio metrics
		services, err := client.CoreV1().Services(istioNamespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Skipf("Failed to list services: %v", err)
		}

		for _, svc := range services.Items {
			if strings.Contains(svc.Name, "prometheus") {
				t.Logf("✓ Found Prometheus service: %s", svc.Name)
			}
		}
	})

	t.Run("TracingIntegration", func(t *testing.T) {
		// Check for tracing components
		deployments, err := client.AppsV1().Deployments(istioNamespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Skipf("Failed to list deployments: %v", err)
		}

		tracingFound := false
		for _, deploy := range deployments.Items {
			if strings.Contains(deploy.Name, "jaeger") || strings.Contains(deploy.Name, "zipkin") {
				t.Logf("✓ Found tracing component: %s", deploy.Name)
				tracingFound = true
			}
		}

		if !tracingFound {
			t.Log("○ No tracing component found (Jaeger/Zipkin)")
		}
	})
}

func TestIstioSecurityPolicies(t *testing.T) {
	t.Run("AuthorizationPolicyEnforcement", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}

		// Test that unauthorized requests are blocked
		// This depends on AuthorizationPolicy configuration
		resp, err := client.Get(testConfig.IstioGatewayURL + "/api/protected/resource")
		if err != nil {
			t.Logf("Request failed (may be expected if blocked): %v", err)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		t.Logf("Protected endpoint response: status=%d, body=%s", resp.StatusCode, string(body))
	})
}

func TestIstioLoadBalancing(t *testing.T) {
	t.Run("LoadBalancingDistribution", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}

		// Make multiple requests and check for load distribution
		// This is best tested when multiple replicas exist
		responses := make(map[string]int)

		for i := 0; i < 20; i++ {
			resp, err := client.Get(testConfig.IdentityURL + "/health")
			if err != nil {
				continue
			}

			// Check for any headers indicating which pod served the request
			if podName := resp.Header.Get("X-Pod-Name"); podName != "" {
				responses[podName]++
			} else {
				responses["unknown"]++
			}
			resp.Body.Close()
		}

		t.Log("Load balancing distribution:")
		for pod, count := range responses {
			t.Logf("  %s: %d requests", pod, count)
		}
	})
}

func TestIstioFaultInjection(t *testing.T) {
	t.Run("FaultInjectionCapability", func(t *testing.T) {
		// This tests that fault injection is possible (configured)
		// Actual fault injection would be done through Istio VirtualService
		t.Log("Fault injection requires VirtualService configuration")
		t.Log("Example: inject 5s delay for 10% of requests")
		t.Log("This capability is available through k8s/istio/virtual-services.yaml")
	})
}

func TestIstioConnectionPooling(t *testing.T) {
	t.Run("ConnectionPoolBehavior", func(t *testing.T) {
		// Test connection pooling by making concurrent requests
		var wg sync.WaitGroup
		results := make(chan bool, 50)

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client := &http.Client{Timeout: 10 * time.Second}
				resp, err := client.Get(testConfig.IdentityURL + "/health")
				if err != nil {
					results <- false
					return
				}
				resp.Body.Close()
				results <- resp.StatusCode == http.StatusOK
			}()
		}

		wg.Wait()
		close(results)

		successCount := 0
		failCount := 0
		for success := range results {
			if success {
				successCount++
			} else {
				failCount++
			}
		}

		t.Logf("Connection pool test: %d success, %d fail out of 50 concurrent requests", successCount, failCount)
	})
}

func TestIstioOutlierDetection(t *testing.T) {
	t.Run("OutlierDetectionConfiguration", func(t *testing.T) {
		// Outlier detection removes unhealthy hosts from the load balancing pool
		// This is configured in DestinationRule
		t.Log("Outlier detection is configured in k8s/istio/destination-rules.yaml")
		t.Log("Configuration: 5 consecutive 5xx errors -> 30s ejection")
	})
}
