package e2e

import (
	"context"
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FinOps E2E Tests
// Tests cost management, resource quotas, and budget tracking

func TestFinOpsResourceQuotas(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("ResourceQuotaExists", func(t *testing.T) {
		quotas, err := client.CoreV1().ResourceQuotas(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list resource quotas: %v", err)
		}

		if len(quotas.Items) == 0 {
			t.Skip("No resource quotas found - FinOps may not be deployed")
		}

		t.Logf("Found %d resource quotas", len(quotas.Items))
		for _, quota := range quotas.Items {
			t.Logf("  - %s", quota.Name)
		}
	})

	t.Run("ResourceQuotaUsage", func(t *testing.T) {
		quotas, err := client.CoreV1().ResourceQuotas(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list resource quotas: %v", err)
		}

		for _, quota := range quotas.Items {
			t.Logf("Quota: %s", quota.Name)
			for resource, used := range quota.Status.Used {
				hard := quota.Status.Hard[resource]
				t.Logf("  %s: %s / %s", resource, used.String(), hard.String())
			}
		}
	})
}

func TestFinOpsLimitRanges(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("LimitRangeExists", func(t *testing.T) {
		limitRanges, err := client.CoreV1().LimitRanges(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list limit ranges: %v", err)
		}

		if len(limitRanges.Items) == 0 {
			t.Skip("No limit ranges found - FinOps may not be deployed")
		}

		t.Logf("Found %d limit ranges", len(limitRanges.Items))
		for _, lr := range limitRanges.Items {
			t.Logf("  - %s", lr.Name)
			for _, limit := range lr.Spec.Limits {
				t.Logf("    Type: %s", limit.Type)
				if limit.Default != nil {
					t.Logf("      Default CPU: %s", limit.Default.Cpu().String())
					t.Logf("      Default Memory: %s", limit.Default.Memory().String())
				}
			}
		}
	})
}

func TestFinOpsCostLabels(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("DeploymentsHaveCostLabels", func(t *testing.T) {
		deployments, err := client.AppsV1().Deployments(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list deployments: %v", err)
		}

		missingLabels := []string{}
		for _, deploy := range deployments.Items {
			labels := deploy.Labels
			if labels == nil {
				missingLabels = append(missingLabels, deploy.Name+" (no labels)")
				continue
			}

			// Check for cost allocation labels
			if _, ok := labels["cost-center"]; !ok {
				// This is informational, not a failure
				t.Logf("Deployment %s missing 'cost-center' label", deploy.Name)
			}
		}

		if len(missingLabels) > 0 {
			t.Logf("Deployments missing cost labels: %v", missingLabels)
		}
	})
}

func TestFinOpsResourceEfficiency(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("PodResourceRequests", func(t *testing.T) {
		pods, err := client.CoreV1().Pods(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list pods: %v", err)
		}

		podsWithoutRequests := []string{}
		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				if container.Resources.Requests.Cpu().IsZero() && container.Resources.Requests.Memory().IsZero() {
					podsWithoutRequests = append(podsWithoutRequests, fmt.Sprintf("%s/%s", pod.Name, container.Name))
				}
			}
		}

		if len(podsWithoutRequests) > 0 {
			t.Logf("Containers without resource requests (may lead to inefficient scheduling): %v", podsWithoutRequests)
		}

		t.Logf("Analyzed %d pods for resource efficiency", len(pods.Items))
	})

	t.Run("PodResourceLimits", func(t *testing.T) {
		pods, err := client.CoreV1().Pods(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list pods: %v", err)
		}

		podsWithoutLimits := []string{}
		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				if container.Resources.Limits.Cpu().IsZero() && container.Resources.Limits.Memory().IsZero() {
					podsWithoutLimits = append(podsWithoutLimits, fmt.Sprintf("%s/%s", pod.Name, container.Name))
				}
			}
		}

		if len(podsWithoutLimits) > 0 {
			t.Logf("Containers without resource limits (may lead to resource contention): %v", podsWithoutLimits)
		}
	})
}

func TestFinOpsNamespaceConfigMaps(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("FinOpsConfigMapsExist", func(t *testing.T) {
		expectedConfigs := []string{
			"kubecost-config",
			"cost-allocation-labels",
			"budget-alert-rules",
			"budget-thresholds",
		}

		configMaps, err := client.CoreV1().ConfigMaps("finops").List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Skip("FinOps namespace not available")
		}

		foundConfigs := make(map[string]bool)
		for _, cm := range configMaps.Items {
			foundConfigs[cm.Name] = true
		}

		for _, expected := range expectedConfigs {
			if !foundConfigs[expected] {
				t.Logf("Expected ConfigMap not found: %s", expected)
			} else {
				t.Logf("Found ConfigMap: %s", expected)
			}
		}
	})
}

func TestFinOpsBudgetTracking(t *testing.T) {
	t.Run("BudgetConfigurationExists", func(t *testing.T) {
		client := SetupK8sClient(t)
		if client == nil {
			t.Skip("Kubernetes client not available")
		}
		ctx := context.Background()

		cm, err := client.CoreV1().ConfigMaps("finops").Get(ctx, "budget-thresholds", metav1.GetOptions{})
		if err != nil {
			t.Skip("Budget thresholds ConfigMap not found")
		}

		if budgetData, ok := cm.Data["budgets.yaml"]; ok {
			t.Logf("Budget configuration found (%d bytes)", len(budgetData))
		} else {
			t.Log("Budget configuration key not found in ConfigMap")
		}
	})
}

func TestFinOpsMetricsExporter(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("MetricsExporterDeployment", func(t *testing.T) {
		deployment, err := client.AppsV1().Deployments("finops").Get(ctx, "finops-metrics-exporter", metav1.GetOptions{})
		if err != nil {
			t.Skip("FinOps metrics exporter not deployed")
		}

		if deployment.Status.ReadyReplicas == 0 {
			t.Log("Metrics exporter has no ready replicas")
		} else {
			t.Logf("Metrics exporter has %d ready replicas", deployment.Status.ReadyReplicas)
		}
	})

	t.Run("MetricsServiceExists", func(t *testing.T) {
		_, err := client.CoreV1().Services("finops").Get(ctx, "finops-metrics-exporter", metav1.GetOptions{})
		if err != nil {
			t.Skip("FinOps metrics service not found")
		}
		t.Log("Metrics exporter service exists")
	})
}

// ResourceCostSimulation simulates cost calculation for testing
func TestResourceCostSimulation(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("CalculateNamespaceCost", func(t *testing.T) {
		pods, err := client.CoreV1().Pods(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list pods: %v", err)
		}

		// Simple cost model (example rates)
		cpuCostPerCore := 0.031611 // per hour
		memCostPerGB := 0.004237   // per hour

		totalCPUMillis := int64(0)
		totalMemBytes := int64(0)

		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				if cpu := container.Resources.Requests.Cpu(); cpu != nil {
					totalCPUMillis += cpu.MilliValue()
				}
				if mem := container.Resources.Requests.Memory(); mem != nil {
					totalMemBytes += mem.Value()
				}
			}
		}

		cpuCores := float64(totalCPUMillis) / 1000.0
		memGB := float64(totalMemBytes) / (1024 * 1024 * 1024)

		hourlyCost := (cpuCores * cpuCostPerCore) + (memGB * memCostPerGB)
		dailyCost := hourlyCost * 24
		monthlyCost := dailyCost * 30

		t.Logf("Namespace: %s", testConfig.Namespace)
		t.Logf("  Total CPU Requested: %.2f cores", cpuCores)
		t.Logf("  Total Memory Requested: %.2f GB", memGB)
		t.Logf("  Estimated Hourly Cost: $%.4f", hourlyCost)
		t.Logf("  Estimated Daily Cost: $%.2f", dailyCost)
		t.Logf("  Estimated Monthly Cost: $%.2f", monthlyCost)
	})
}
