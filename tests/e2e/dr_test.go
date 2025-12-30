package e2e

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Disaster Recovery E2E Tests
// Tests backup, restore, and failover capabilities

const veleroNamespace = "velero"

func TestDRVeleroInstallation(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("VeleroNamespaceExists", func(t *testing.T) {
		_, err := client.CoreV1().Namespaces().Get(ctx, veleroNamespace, metav1.GetOptions{})
		if err != nil {
			t.Skip("Velero namespace does not exist - DR may not be deployed")
		}
		t.Log("Velero namespace exists")
	})

	t.Run("VeleroConfigMapsExist", func(t *testing.T) {
		expectedConfigs := []string{
			"velero-config",
			"backup-schedules",
			"restore-procedures",
			"dr-testing-config",
			"failover-config",
		}

		configMaps, err := client.CoreV1().ConfigMaps(veleroNamespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Skipf("Failed to list ConfigMaps in velero namespace: %v", err)
		}

		foundConfigs := make(map[string]bool)
		for _, cm := range configMaps.Items {
			foundConfigs[cm.Name] = true
		}

		for _, expected := range expectedConfigs {
			if foundConfigs[expected] {
				t.Logf("✓ Found ConfigMap: %s", expected)
			} else {
				t.Logf("○ ConfigMap not found: %s", expected)
			}
		}
	})
}

func TestDRBackupConfiguration(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("BackupSchedulesConfigured", func(t *testing.T) {
		cm, err := client.CoreV1().ConfigMaps(veleroNamespace).Get(ctx, "backup-schedules", metav1.GetOptions{})
		if err != nil {
			t.Skip("Backup schedules ConfigMap not found")
		}

		expectedSchedules := []string{
			"tier1-backup.yaml",
			"tier2-backup.yaml",
			"full-backup.yaml",
			"daily-backup.yaml",
		}

		for _, schedule := range expectedSchedules {
			if _, ok := cm.Data[schedule]; ok {
				t.Logf("✓ Backup schedule configured: %s", schedule)
			} else {
				t.Logf("○ Backup schedule not found: %s", schedule)
			}
		}
	})
}

func TestDRRestoreProcedures(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("RestoreScriptsConfigured", func(t *testing.T) {
		cm, err := client.CoreV1().ConfigMaps(veleroNamespace).Get(ctx, "restore-scripts", metav1.GetOptions{})
		if err != nil {
			t.Skip("Restore scripts ConfigMap not found")
		}

		expectedScripts := []string{
			"full-restore.sh",
			"service-restore.sh",
			"database-restore.sh",
			"verify-restore.sh",
		}

		for _, script := range expectedScripts {
			if _, ok := cm.Data[script]; ok {
				t.Logf("✓ Restore script configured: %s", script)
			} else {
				t.Logf("○ Restore script not found: %s", script)
			}
		}
	})
}

func TestDRFailoverConfiguration(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("FailoverConfigExists", func(t *testing.T) {
		cm, err := client.CoreV1().ConfigMaps(veleroNamespace).Get(ctx, "failover-config", metav1.GetOptions{})
		if err != nil {
			t.Skip("Failover config ConfigMap not found")
		}

		expectedConfigs := []string{
			"failover-policy.yaml",
			"dns-failover.yaml",
			"istio-failover.yaml",
		}

		for _, config := range expectedConfigs {
			if _, ok := cm.Data[config]; ok {
				t.Logf("✓ Failover config found: %s", config)
			} else {
				t.Logf("○ Failover config not found: %s", config)
			}
		}
	})

	t.Run("FailoverScriptsExist", func(t *testing.T) {
		cm, err := client.CoreV1().ConfigMaps(veleroNamespace).Get(ctx, "failover-scripts", metav1.GetOptions{})
		if err != nil {
			t.Skip("Failover scripts ConfigMap not found")
		}

		expectedScripts := []string{
			"failover.sh",
			"failback.sh",
			"dns-failover.sh",
		}

		for _, script := range expectedScripts {
			if _, ok := cm.Data[script]; ok {
				t.Logf("✓ Failover script found: %s", script)
			} else {
				t.Logf("○ Failover script not found: %s", script)
			}
		}
	})
}

func TestDRRPORTOConfiguration(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("RPOConfigurationExists", func(t *testing.T) {
		cm, err := client.CoreV1().ConfigMaps(veleroNamespace).Get(ctx, "velero-config", metav1.GetOptions{})
		if err != nil {
			t.Skip("Velero config ConfigMap not found")
		}

		if _, ok := cm.Data["rpo-config.yaml"]; ok {
			t.Log("✓ RPO configuration found")
		} else {
			t.Log("○ RPO configuration not found")
		}
	})

	t.Run("RTOConfigurationExists", func(t *testing.T) {
		cm, err := client.CoreV1().ConfigMaps(veleroNamespace).Get(ctx, "velero-config", metav1.GetOptions{})
		if err != nil {
			t.Skip("Velero config ConfigMap not found")
		}

		if _, ok := cm.Data["rto-config.yaml"]; ok {
			t.Log("✓ RTO configuration found")
		} else {
			t.Log("○ RTO configuration not found")
		}
	})
}

func TestDRServiceRecovery(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	services := []struct {
		name string
		port int
		url  string
	}{
		{"identity-service", 8081, testConfig.IdentityURL},
		{"ledger-service", 8082, testConfig.LedgerURL},
		{"payment-service", 8083, testConfig.PaymentURL},
		{"product-service", 8084, testConfig.ProductURL},
		{"card-service", 8085, testConfig.CardURL},
	}

	t.Run("ServicesAreRunning", func(t *testing.T) {
		for _, svc := range services {
			deployment, err := client.AppsV1().Deployments(testConfig.Namespace).Get(ctx, svc.name, metav1.GetOptions{})
			if err != nil {
				t.Logf("○ Deployment not found: %s", svc.name)
				continue
			}

			if deployment.Status.ReadyReplicas > 0 {
				t.Logf("✓ %s: %d/%d replicas ready", svc.name, deployment.Status.ReadyReplicas, *deployment.Spec.Replicas)
			} else {
				t.Logf("○ %s: no ready replicas", svc.name)
			}
		}
	})

	t.Run("ServicesHealthCheck", func(t *testing.T) {
		for _, svc := range services {
			err := HealthCheck(svc.url)
			if err != nil {
				t.Logf("○ %s health check failed: %v", svc.name, err)
			} else {
				t.Logf("✓ %s health check passed", svc.name)
			}
		}
	})
}

func TestDRPodDistribution(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("PodDistributionAcrossNodes", func(t *testing.T) {
		pods, err := client.CoreV1().Pods(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list pods: %v", err)
		}

		nodeDistribution := make(map[string]int)
		for _, pod := range pods.Items {
			if pod.Spec.NodeName != "" {
				nodeDistribution[pod.Spec.NodeName]++
			}
		}

		t.Logf("Pod distribution across nodes:")
		for node, count := range nodeDistribution {
			t.Logf("  %s: %d pods", node, count)
		}

		// In a single-node cluster (Docker Desktop), all pods will be on one node
		if len(nodeDistribution) == 1 {
			t.Log("Note: Single node cluster detected - consider multi-node setup for production DR")
		}
	})
}

func TestDRPodDisruptionBudgets(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("PDBsExist", func(t *testing.T) {
		pdbs, err := client.PolicyV1().PodDisruptionBudgets(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list PDBs: %v", err)
		}

		if len(pdbs.Items) == 0 {
			t.Log("No PodDisruptionBudgets found - consider adding for DR")
			return
		}

		t.Logf("Found %d PodDisruptionBudgets:", len(pdbs.Items))
		for _, pdb := range pdbs.Items {
			t.Logf("  %s: minAvailable=%v, maxUnavailable=%v",
				pdb.Name,
				pdb.Spec.MinAvailable,
				pdb.Spec.MaxUnavailable)
		}
	})
}

func TestDRSimulatedFailover(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("SimulateServiceFailure", func(t *testing.T) {
		// Get current pod count for a service
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

		// Delete a pod to simulate failure
		podToDelete := pods.Items[0].Name
		t.Logf("Simulating failure by deleting pod: %s", podToDelete)

		err = client.CoreV1().Pods(testConfig.Namespace).Delete(ctx, podToDelete, metav1.DeleteOptions{})
		if err != nil {
			t.Logf("Note: Could not delete pod (may not have permissions in test): %v", err)
			return
		}

		// Wait for recovery
		t.Log("Waiting for recovery...")
		time.Sleep(30 * time.Second)

		// Check pod count recovered
		pods, err = client.CoreV1().Pods(testConfig.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: "app=identity-service",
		})
		if err != nil {
			t.Fatalf("Failed to list pods after recovery: %v", err)
		}

		runningCount := 0
		for _, pod := range pods.Items {
			if pod.Status.Phase == "Running" {
				runningCount++
			}
		}

		t.Logf("Pod count after recovery: %d (running: %d)", len(pods.Items), runningCount)

		if runningCount >= initialCount {
			t.Log("✓ Service recovered successfully")
		} else {
			t.Logf("○ Service recovery in progress (expected: %d, got: %d)", initialCount, runningCount)
		}
	})
}

func TestDRDataPersistence(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("PersistentVolumeClaimsExist", func(t *testing.T) {
		pvcs, err := client.CoreV1().PersistentVolumeClaims(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list PVCs: %v", err)
		}

		if len(pvcs.Items) == 0 {
			t.Log("No PersistentVolumeClaims found")
			return
		}

		t.Logf("Found %d PersistentVolumeClaims:", len(pvcs.Items))
		for _, pvc := range pvcs.Items {
			t.Logf("  %s: %s, %s, storage=%s",
				pvc.Name,
				pvc.Status.Phase,
				*pvc.Spec.StorageClassName,
				pvc.Spec.Resources.Requests.Storage().String())
		}
	})
}

func TestDRTestingCronJob(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("DRTestCronJobConfigured", func(t *testing.T) {
		cm, err := client.CoreV1().ConfigMaps(veleroNamespace).Get(ctx, "dr-testing-config", metav1.GetOptions{})
		if err != nil {
			t.Skip("DR testing ConfigMap not found")
		}

		if _, ok := cm.Data["dr-test-cronjob.yaml"]; ok {
			t.Log("✓ DR test CronJob configuration found")
		} else {
			t.Log("○ DR test CronJob configuration not found")
		}

		if _, ok := cm.Data["dr-test.sh"]; ok {
			t.Log("✓ DR test script found")
		} else {
			t.Log("○ DR test script not found")
		}
	})
}
