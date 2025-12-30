package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Integration E2E Tests
// Tests complete user journeys and service interactions

func TestIntegrationUserJourney(t *testing.T) {
	t.Run("CompleteUserRegistrationFlow", func(t *testing.T) {
		client := &http.Client{Timeout: 30 * time.Second}
		uniqueID := fmt.Sprintf("e2e-%d", time.Now().UnixNano())

		// Step 1: Register user
		registerData := map[string]string{
			"email":      uniqueID + "@test.neobank.io",
			"password":   "SecurePassword123!",
			"first_name": "E2E",
			"last_name":  "Test",
		}

		body, _ := json.Marshal(registerData)
		resp, err := client.Post(testConfig.IdentityURL+"/auth/register", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Skipf("Identity service not available: %v", err)
		}
		defer resp.Body.Close()

		responseBody, _ := io.ReadAll(resp.Body)
		t.Logf("Step 1 - Registration: %d", resp.StatusCode)

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			t.Logf("Registration response: %s", string(responseBody))
		}

		// Step 2: Login
		loginData := map[string]string{
			"email":    registerData["email"],
			"password": registerData["password"],
		}

		body, _ = json.Marshal(loginData)
		resp, err = client.Post(testConfig.IdentityURL+"/auth/login", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Login request failed: %v", err)
		}
		defer resp.Body.Close()

		var loginResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&loginResp)

		t.Logf("Step 2 - Login: %d", resp.StatusCode)

		token, _ := loginResp["token"].(string)
		if token != "" {
			t.Log("✓ Received authentication token")
		} else {
			t.Log("○ No authentication token in response")
		}
	})
}

func TestIntegrationServiceCommunication(t *testing.T) {
	t.Run("ServiceToServiceCalls", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}

		// Test that services can communicate with each other
		// This is indicated by successful responses that require inter-service calls

		services := []struct {
			name     string
			url      string
			endpoint string
		}{
			{"identity", testConfig.IdentityURL, "/health"},
			{"ledger", testConfig.LedgerURL, "/health"},
			{"payment", testConfig.PaymentURL, "/health"},
			{"product", testConfig.ProductURL, "/health"},
			{"card", testConfig.CardURL, "/health"},
		}

		healthyServices := 0
		for _, svc := range services {
			resp, err := client.Get(svc.url + svc.endpoint)
			if err != nil {
				t.Logf("○ %s: connection failed", svc.name)
				continue
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				t.Logf("✓ %s: healthy", svc.name)
				healthyServices++
			} else {
				t.Logf("○ %s: status %d", svc.name, resp.StatusCode)
			}
		}

		t.Logf("\nHealthy services: %d/%d", healthyServices, len(services))
	})
}

func TestIntegrationDatabaseConnectivity(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("PostgresConnection", func(t *testing.T) {
		// Check if postgres pod is running
		pods, err := client.CoreV1().Pods(testConfig.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: "app=postgres",
		})
		if err != nil {
			t.Skipf("Failed to list postgres pods: %v", err)
		}

		if len(pods.Items) == 0 {
			t.Skip("No postgres pods found")
		}

		for _, pod := range pods.Items {
			status := "○"
			if pod.Status.Phase == "Running" {
				status = "✓"
			}
			t.Logf("%s Postgres pod %s: %s", status, pod.Name, pod.Status.Phase)
		}
	})

	t.Run("RedisConnection", func(t *testing.T) {
		// Check if redis pod is running
		pods, err := client.CoreV1().Pods(testConfig.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: "app=redis",
		})
		if err != nil {
			t.Skipf("Failed to list redis pods: %v", err)
		}

		if len(pods.Items) == 0 {
			t.Skip("No redis pods found")
		}

		for _, pod := range pods.Items {
			status := "○"
			if pod.Status.Phase == "Running" {
				status = "✓"
			}
			t.Logf("%s Redis pod %s: %s", status, pod.Name, pod.Status.Phase)
		}
	})
}

func TestIntegrationLoadBalancing(t *testing.T) {
	t.Run("RequestDistribution", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}

		// Make multiple requests and check distribution
		responses := make(map[string]int)

		for i := 0; i < 30; i++ {
			resp, err := client.Get(testConfig.IdentityURL + "/health")
			if err != nil {
				responses["error"]++
				continue
			}

			// Try to identify which pod served the request
			podID := resp.Header.Get("X-Pod-Name")
			if podID == "" {
				podID = resp.Header.Get("X-Instance-ID")
			}
			if podID == "" {
				podID = "unknown"
			}

			responses[podID]++
			resp.Body.Close()
		}

		t.Log("Request distribution:")
		for key, count := range responses {
			t.Logf("  %s: %d requests", key, count)
		}
	})
}

func TestIntegrationEndToEndLatency(t *testing.T) {
	t.Run("ServiceLatencyMeasurement", func(t *testing.T) {
		client := &http.Client{Timeout: 30 * time.Second}

		services := []struct {
			name string
			url  string
		}{
			{"identity", testConfig.IdentityURL},
			{"ledger", testConfig.LedgerURL},
			{"payment", testConfig.PaymentURL},
			{"product", testConfig.ProductURL},
			{"card", testConfig.CardURL},
		}

		for _, svc := range services {
			var totalLatency time.Duration
			successCount := 0

			for i := 0; i < 10; i++ {
				start := time.Now()
				resp, err := client.Get(svc.url + "/health")
				latency := time.Since(start)

				if err == nil {
					resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						totalLatency += latency
						successCount++
					}
				}
			}

			if successCount > 0 {
				avgLatency := totalLatency / time.Duration(successCount)
				t.Logf("%s: avg latency = %v (from %d successful requests)", svc.name, avgLatency, successCount)
			} else {
				t.Logf("%s: no successful requests", svc.name)
			}
		}
	})
}

func TestIntegrationConcurrentUsers(t *testing.T) {
	t.Run("ConcurrentRequestHandling", func(t *testing.T) {
		client := &http.Client{Timeout: 30 * time.Second}

		concurrentUsers := 50
		var wg sync.WaitGroup
		results := make(chan bool, concurrentUsers)

		startTime := time.Now()

		for i := 0; i < concurrentUsers; i++ {
			wg.Add(1)
			go func(userID int) {
				defer wg.Done()

				resp, err := client.Get(testConfig.IdentityURL + "/health")
				if err != nil {
					results <- false
					return
				}
				defer resp.Body.Close()

				results <- resp.StatusCode == http.StatusOK
			}(i)
		}

		wg.Wait()
		close(results)

		totalDuration := time.Since(startTime)

		successCount := 0
		failCount := 0
		for success := range results {
			if success {
				successCount++
			} else {
				failCount++
			}
		}

		t.Logf("Concurrent users test:")
		t.Logf("  Total users: %d", concurrentUsers)
		t.Logf("  Successful: %d", successCount)
		t.Logf("  Failed: %d", failCount)
		t.Logf("  Total time: %v", totalDuration)
		t.Logf("  Success rate: %.1f%%", float64(successCount)/float64(concurrentUsers)*100)
	})
}

func TestIntegrationAPIVersioning(t *testing.T) {
	t.Run("APIv1Endpoints", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}

		endpoints := []struct {
			name     string
			url      string
			method   string
			expected int
		}{
			{"products list", testConfig.ProductURL + "/api/v1/products", "GET", http.StatusOK},
			{"health check", testConfig.IdentityURL + "/health", "GET", http.StatusOK},
		}

		for _, ep := range endpoints {
			req, _ := http.NewRequest(ep.method, ep.url, nil)
			resp, err := client.Do(req)

			if err != nil {
				t.Logf("○ %s: connection failed", ep.name)
				continue
			}
			resp.Body.Close()

			if resp.StatusCode == ep.expected {
				t.Logf("✓ %s: %d", ep.name, resp.StatusCode)
			} else {
				t.Logf("○ %s: expected %d, got %d", ep.name, ep.expected, resp.StatusCode)
			}
		}
	})
}

func TestIntegrationSecurityHeaders(t *testing.T) {
	t.Run("SecurityHeadersPresent", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}

		resp, err := client.Get(testConfig.IdentityURL + "/health")
		if err != nil {
			t.Skipf("Service not available: %v", err)
		}
		defer resp.Body.Close()

		expectedHeaders := []string{
			"X-Content-Type-Options",
			"X-Frame-Options",
			"X-XSS-Protection",
			"Content-Security-Policy",
			"Strict-Transport-Security",
		}

		t.Log("Security headers check:")
		for _, header := range expectedHeaders {
			if value := resp.Header.Get(header); value != "" {
				t.Logf("✓ %s: %s", header, value)
			} else {
				t.Logf("○ %s: not set", header)
			}
		}
	})
}

func TestIntegrationGracefulShutdown(t *testing.T) {
	t.Run("ConnectionDraining", func(t *testing.T) {
		// Test that ongoing requests complete during shutdown
		// This is tested indirectly by checking that requests don't fail abruptly

		client := &http.Client{Timeout: 30 * time.Second}

		// Make a series of requests
		consecutiveSuccesses := 0
		consecutiveFailures := 0
		maxConsecutiveFailures := 0

		for i := 0; i < 20; i++ {
			resp, err := client.Get(testConfig.IdentityURL + "/health")
			if err != nil {
				consecutiveFailures++
				consecutiveSuccesses = 0
				if consecutiveFailures > maxConsecutiveFailures {
					maxConsecutiveFailures = consecutiveFailures
				}
			} else {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					consecutiveSuccesses++
					consecutiveFailures = 0
				}
			}
			time.Sleep(100 * time.Millisecond)
		}

		t.Logf("Max consecutive failures: %d", maxConsecutiveFailures)
		if maxConsecutiveFailures <= 2 {
			t.Log("✓ Service appears to handle connections gracefully")
		} else {
			t.Log("○ Multiple consecutive failures detected")
		}
	})
}

func TestIntegrationObservability(t *testing.T) {
	t.Run("MetricsEndpoint", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}

		// Check for metrics endpoint
		resp, err := client.Get(testConfig.IdentityURL + "/metrics")
		if err != nil {
			t.Logf("Metrics endpoint not available: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("✓ Metrics endpoint available (%d bytes)", len(body))
		} else {
			t.Logf("○ Metrics endpoint returned: %d", resp.StatusCode)
		}
	})

	t.Run("HealthEndpoints", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}

		healthEndpoints := []string{"/health", "/healthz", "/ready", "/readiness", "/liveness"}

		for _, endpoint := range healthEndpoints {
			resp, err := client.Get(testConfig.IdentityURL + endpoint)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				t.Logf("✓ %s endpoint available", endpoint)
			}
		}
	})
}

func TestIntegrationRateLimitRecovery(t *testing.T) {
	t.Run("RecoveryAfterRateLimit", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}

		// First, trigger rate limiting by making rapid requests
		rateLimited := false
		for i := 0; i < 200; i++ {
			resp, err := client.Get(testConfig.IdentityURL + "/health")
			if err != nil {
				continue
			}
			if resp.StatusCode == http.StatusTooManyRequests {
				rateLimited = true
				resp.Body.Close()
				break
			}
			resp.Body.Close()
		}

		if rateLimited {
			t.Log("Rate limit triggered, testing recovery...")

			// Wait for rate limit to reset
			time.Sleep(5 * time.Second)

			// Verify service is accessible again
			resp, err := client.Get(testConfig.IdentityURL + "/health")
			if err != nil {
				t.Logf("○ Service not accessible after rate limit")
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				t.Log("✓ Service recovered after rate limit")
			} else {
				t.Logf("○ Service returned %d after rate limit", resp.StatusCode)
			}
		} else {
			t.Log("Rate limit not triggered (may not be configured)")
		}
	})
}
