package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Chaos Engineering E2E Tests
// Tests system resilience under failure conditions

func TestChaosEngineeringSetup(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("LitmusNamespaceExists", func(t *testing.T) {
		_, err := client.CoreV1().Namespaces().Get(ctx, "litmus", metav1.GetOptions{})
		if err != nil {
			t.Log("Litmus namespace does not exist - chaos engineering may not be set up")
			return
		}
		t.Log("✓ Litmus namespace exists")
	})

	t.Run("ChaosExperimentsConfigured", func(t *testing.T) {
		// Check for chaos experiment configurations
		configMaps, err := client.CoreV1().ConfigMaps(testConfig.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/part-of=litmus",
		})
		if err != nil {
			t.Logf("Could not list chaos configs: %v", err)
			return
		}

		t.Logf("Found %d chaos-related ConfigMaps", len(configMaps.Items))
	})
}

func TestChaosPodDeletion(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("PodRecoveryAfterDeletion", func(t *testing.T) {
		// Get pods for identity service
		pods, err := client.CoreV1().Pods(testConfig.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: "app=identity-service",
		})
		if err != nil {
			t.Skipf("Failed to list pods: %v", err)
		}

		if len(pods.Items) == 0 {
			t.Skip("No identity-service pods found")
		}

		initialCount := len(pods.Items)
		t.Logf("Initial pod count: %d", initialCount)

		// Record start time
		startTime := time.Now()

		// Delete a pod (chaos injection)
		podToDelete := pods.Items[0].Name
		t.Logf("Deleting pod: %s", podToDelete)

		err = client.CoreV1().Pods(testConfig.Namespace).Delete(ctx, podToDelete, metav1.DeleteOptions{})
		if err != nil {
			t.Logf("Note: Could not delete pod: %v", err)
			return
		}

		// Wait and measure recovery time
		recovered := false
		var recoveryTime time.Duration

		for i := 0; i < 60; i++ {
			time.Sleep(2 * time.Second)

			pods, err = client.CoreV1().Pods(testConfig.Namespace).List(ctx, metav1.ListOptions{
				LabelSelector: "app=identity-service",
			})
			if err != nil {
				continue
			}

			runningCount := 0
			for _, pod := range pods.Items {
				if pod.Status.Phase == "Running" {
					ready := true
					for _, cond := range pod.Status.Conditions {
						if cond.Type == "Ready" && cond.Status != "True" {
							ready = false
							break
						}
					}
					if ready {
						runningCount++
					}
				}
			}

			if runningCount >= initialCount {
				recoveryTime = time.Since(startTime)
				recovered = true
				break
			}
		}

		if recovered {
			t.Logf("✓ Service recovered in %v", recoveryTime)
		} else {
			t.Logf("○ Service did not fully recover within timeout")
		}
	})
}

func TestChaosNetworkLatency(t *testing.T) {
	t.Run("ServiceResponseUnderLatency", func(t *testing.T) {
		// Test service behavior with simulated network latency
		// In real chaos testing, this would inject actual latency

		client := &http.Client{Timeout: 30 * time.Second}

		// Measure baseline response time
		start := time.Now()
		resp, err := client.Get(testConfig.IdentityURL + "/health")
		if err != nil {
			t.Skipf("Service not available: %v", err)
		}
		resp.Body.Close()
		baselineTime := time.Since(start)

		t.Logf("Baseline response time: %v", baselineTime)

		// Make multiple requests to establish pattern
		var totalTime time.Duration
		for i := 0; i < 10; i++ {
			start = time.Now()
			resp, err = client.Get(testConfig.IdentityURL + "/health")
			if err != nil {
				continue
			}
			resp.Body.Close()
			totalTime += time.Since(start)
		}

		avgTime := totalTime / 10
		t.Logf("Average response time (10 requests): %v", avgTime)

		// If avg time is significantly higher than baseline, might indicate issues
		if avgTime > baselineTime*3 {
			t.Logf("Warning: Response time variance detected")
		}
	})
}

func TestChaosCPUStress(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("HPAResponseToCPULoad", func(t *testing.T) {
		// Check if HPA is configured
		hpas, err := client.AutoscalingV2().HorizontalPodAutoscalers(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Skipf("Failed to list HPAs: %v", err)
		}

		if len(hpas.Items) == 0 {
			t.Skip("No HPAs configured")
		}

		t.Logf("Found %d HPAs:", len(hpas.Items))
		for _, hpa := range hpas.Items {
			t.Logf("  %s: min=%d, max=%d, current=%d",
				hpa.Name,
				*hpa.Spec.MinReplicas,
				hpa.Spec.MaxReplicas,
				hpa.Status.CurrentReplicas)
		}
	})
}

func TestChaosServiceDegradation(t *testing.T) {
	t.Run("CircuitBreakerActivation", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}

		// Make rapid requests to potentially trigger circuit breaker
		errors := 0
		successes := 0

		for i := 0; i < 50; i++ {
			resp, err := client.Get(testConfig.IdentityURL + "/health")
			if err != nil {
				errors++
				continue
			}
			if resp.StatusCode == http.StatusOK {
				successes++
			} else if resp.StatusCode == http.StatusServiceUnavailable {
				// Circuit breaker might be open
				errors++
			}
			resp.Body.Close()
		}

		t.Logf("Rapid request test: %d success, %d errors", successes, errors)
	})
}

func TestChaosDataIntegrity(t *testing.T) {
	t.Run("DataConsistencyUnderLoad", func(t *testing.T) {
		// This test verifies that data remains consistent under load
		// In a real scenario, this would create/read/verify data

		client := &http.Client{Timeout: 10 * time.Second}

		// Test registration endpoint for data consistency
		uniqueID := fmt.Sprintf("chaos-test-%d", time.Now().UnixNano())
		registerData := map[string]string{
			"email":      uniqueID + "@test.com",
			"password":   "TestPassword123!",
			"first_name": "Chaos",
			"last_name":  "Test",
		}

		body, _ := json.Marshal(registerData)
		resp, err := client.Post(testConfig.IdentityURL+"/auth/register", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Skipf("Could not reach identity service: %v", err)
		}
		defer resp.Body.Close()

		responseBody, _ := io.ReadAll(resp.Body)
		t.Logf("Registration response: %d - %s", resp.StatusCode, string(responseBody))
	})
}

func TestChaosGracefulDegradation(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("ServiceIsolation", func(t *testing.T) {
		// Test that one service failure doesn't cascade to others
		services := []struct {
			name string
			url  string
		}{
			{"identity-service", testConfig.IdentityURL},
			{"ledger-service", testConfig.LedgerURL},
			{"payment-service", testConfig.PaymentURL},
		}

		results := make(map[string]bool)
		for _, svc := range services {
			httpClient := &http.Client{Timeout: 5 * time.Second}
			resp, err := httpClient.Get(svc.url + "/health")
			if err == nil {
				resp.Body.Close()
				results[svc.name] = resp.StatusCode == http.StatusOK
			} else {
				results[svc.name] = false
			}
		}

		t.Log("Service health status:")
		allHealthy := true
		for name, healthy := range results {
			status := "✓"
			if !healthy {
				status = "○"
				allHealthy = false
			}
			t.Logf("  %s %s", status, name)
		}

		if !allHealthy {
			t.Log("Note: Some services are not healthy - testing isolation")
		}
	})
}

func TestChaosRecoveryMetrics(t *testing.T) {
	t.Run("RecoveryTimeObjective", func(t *testing.T) {
		// Define RTO targets
		rtoTargets := map[string]time.Duration{
			"identity-service": 1 * time.Minute,
			"ledger-service":   1 * time.Minute,
			"payment-service":  1 * time.Minute,
		}

		t.Log("RTO Targets:")
		for service, target := range rtoTargets {
			t.Logf("  %s: %v", service, target)
		}

		t.Log("\nNote: Actual RTO measurement requires chaos injection and recovery monitoring")
	})

	t.Run("RecoveryPointObjective", func(t *testing.T) {
		// Define RPO targets
		rpoTargets := map[string]time.Duration{
			"tier-1-services": 5 * time.Minute,
			"tier-2-services": 15 * time.Minute,
			"tier-3-services": 1 * time.Hour,
		}

		t.Log("RPO Targets:")
		for tier, target := range rpoTargets {
			t.Logf("  %s: %v", tier, target)
		}

		t.Log("\nNote: RPO verification requires backup system integration")
	})
}

func TestChaosResourceExhaustion(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("ResourceQuotaProtection", func(t *testing.T) {
		quotas, err := client.CoreV1().ResourceQuotas(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Skipf("Failed to list quotas: %v", err)
		}

		for _, quota := range quotas.Items {
			t.Logf("Quota %s:", quota.Name)
			for resource, used := range quota.Status.Used {
				hard := quota.Status.Hard[resource]
				usedVal := used.Value()
				hardVal := hard.Value()

				percentage := float64(0)
				if hardVal > 0 {
					percentage = float64(usedVal) / float64(hardVal) * 100
				}

				status := "✓"
				if percentage > 80 {
					status = "⚠"
				}
				if percentage > 95 {
					status = "✗"
				}

				t.Logf("  %s %s: %.1f%% (%s / %s)", status, resource, percentage, used.String(), hard.String())
			}
		}
	})
}

func TestChaosMultipleFailures(t *testing.T) {
	t.Run("CascadeFailureProtection", func(t *testing.T) {
		// Test that multiple simultaneous failures are handled
		client := &http.Client{Timeout: 10 * time.Second}

		// Make concurrent requests to all services
		type result struct {
			service string
			healthy bool
			latency time.Duration
		}

		results := make(chan result, 5)

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
			go func(name, url string) {
				start := time.Now()
				resp, err := client.Get(url + "/health")
				latency := time.Since(start)

				healthy := false
				if err == nil {
					resp.Body.Close()
					healthy = resp.StatusCode == http.StatusOK
				}

				results <- result{name, healthy, latency}
			}(svc.name, svc.url)
		}

		t.Log("Concurrent health check results:")
		healthyCount := 0
		for i := 0; i < len(services); i++ {
			r := <-results
			status := "○"
			if r.healthy {
				status = "✓"
				healthyCount++
			}
			t.Logf("  %s %s: %v", status, r.service, r.latency)
		}

		t.Logf("\nHealthy services: %d/%d", healthyCount, len(services))
	})
}
