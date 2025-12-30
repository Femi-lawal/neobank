package e2e

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Kubernetes Cluster E2E Tests
// Tests cluster health, configuration, and infrastructure

func TestClusterHealth(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("NodeStatus", func(t *testing.T) {
		nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list nodes: %v", err)
		}

		t.Logf("Found %d nodes:", len(nodes.Items))
		for _, node := range nodes.Items {
			ready := "NotReady"
			for _, cond := range node.Status.Conditions {
				if cond.Type == "Ready" && cond.Status == "True" {
					ready = "Ready"
					break
				}
			}
			t.Logf("  %s: %s", node.Name, ready)
		}
	})

	t.Run("SystemPodsHealth", func(t *testing.T) {
		pods, err := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list system pods: %v", err)
		}

		running := 0
		notRunning := 0
		for _, pod := range pods.Items {
			if pod.Status.Phase == "Running" {
				running++
			} else {
				notRunning++
				t.Logf("  Non-running system pod: %s (%s)", pod.Name, pod.Status.Phase)
			}
		}

		t.Logf("System pods: %d running, %d not running", running, notRunning)
	})
}

func TestNamespaceConfiguration(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("RequiredNamespacesExist", func(t *testing.T) {
		requiredNamespaces := []string{
			"neobank",
			"default",
			"kube-system",
		}

		optionalNamespaces := []string{
			"istio-system",
			"finops",
			"velero",
		}

		for _, ns := range requiredNamespaces {
			_, err := client.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
			if err != nil {
				t.Errorf("Required namespace missing: %s", ns)
			} else {
				t.Logf("✓ Required namespace exists: %s", ns)
			}
		}

		for _, ns := range optionalNamespaces {
			_, err := client.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
			if err != nil {
				t.Logf("○ Optional namespace missing: %s", ns)
			} else {
				t.Logf("✓ Optional namespace exists: %s", ns)
			}
		}
	})
}

func TestDeploymentConfiguration(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("NeobankDeployments", func(t *testing.T) {
		deployments, err := client.AppsV1().Deployments(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list deployments: %v", err)
		}

		expectedDeployments := []string{
			"identity-service",
			"ledger-service",
			"payment-service",
			"product-service",
			"card-service",
			"frontend",
		}

		foundDeployments := make(map[string]bool)
		for _, deploy := range deployments.Items {
			foundDeployments[deploy.Name] = true
		}

		for _, expected := range expectedDeployments {
			if foundDeployments[expected] {
				t.Logf("✓ Deployment found: %s", expected)
			} else {
				t.Logf("○ Deployment missing: %s", expected)
			}
		}
	})

	t.Run("DeploymentReplicas", func(t *testing.T) {
		deployments, err := client.AppsV1().Deployments(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list deployments: %v", err)
		}

		for _, deploy := range deployments.Items {
			if *deploy.Spec.Replicas == 0 {
				t.Logf("○ %s: 0 replicas configured", deploy.Name)
			} else if deploy.Status.ReadyReplicas < *deploy.Spec.Replicas {
				t.Logf("⚠ %s: %d/%d replicas ready", deploy.Name, deploy.Status.ReadyReplicas, *deploy.Spec.Replicas)
			} else {
				t.Logf("✓ %s: %d/%d replicas ready", deploy.Name, deploy.Status.ReadyReplicas, *deploy.Spec.Replicas)
			}
		}
	})
}

func TestServiceConfiguration(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("NeobankServices", func(t *testing.T) {
		services, err := client.CoreV1().Services(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list services: %v", err)
		}

		t.Logf("Found %d services:", len(services.Items))
		for _, svc := range services.Items {
			ports := []string{}
			for _, port := range svc.Spec.Ports {
				ports = append(ports, string(port.Port))
			}
			t.Logf("  %s: type=%s, ports=%v", svc.Name, svc.Spec.Type, ports)
		}
	})
}

func TestConfigMaps(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("NeobankConfigMaps", func(t *testing.T) {
		configMaps, err := client.CoreV1().ConfigMaps(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list ConfigMaps: %v", err)
		}

		t.Logf("Found %d ConfigMaps:", len(configMaps.Items))
		for _, cm := range configMaps.Items {
			keys := []string{}
			for key := range cm.Data {
				keys = append(keys, key)
			}
			t.Logf("  %s: keys=%v", cm.Name, keys)
		}
	})
}

func TestSecrets(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("NeobankSecrets", func(t *testing.T) {
		secrets, err := client.CoreV1().Secrets(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list Secrets: %v", err)
		}

		t.Logf("Found %d Secrets:", len(secrets.Items))
		for _, secret := range secrets.Items {
			// Don't log secret values, just names and types
			t.Logf("  %s: type=%s, keys=%d", secret.Name, secret.Type, len(secret.Data))
		}
	})
}

func TestPodSecurityContext(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("SecurityContextConfiguration", func(t *testing.T) {
		pods, err := client.CoreV1().Pods(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list pods: %v", err)
		}

		for _, pod := range pods.Items {
			securityIssues := []string{}

			// Check pod security context
			if pod.Spec.SecurityContext != nil {
				if pod.Spec.SecurityContext.RunAsNonRoot == nil || !*pod.Spec.SecurityContext.RunAsNonRoot {
					securityIssues = append(securityIssues, "runAsNonRoot not set")
				}
			} else {
				securityIssues = append(securityIssues, "no pod security context")
			}

			// Check container security contexts
			for _, container := range pod.Spec.Containers {
				if container.SecurityContext == nil {
					securityIssues = append(securityIssues, container.Name+": no security context")
				} else {
					if container.SecurityContext.AllowPrivilegeEscalation == nil || *container.SecurityContext.AllowPrivilegeEscalation {
						securityIssues = append(securityIssues, container.Name+": privilege escalation allowed")
					}
					if container.SecurityContext.ReadOnlyRootFilesystem == nil || !*container.SecurityContext.ReadOnlyRootFilesystem {
						securityIssues = append(securityIssues, container.Name+": root filesystem not read-only")
					}
				}
			}

			if len(securityIssues) == 0 {
				t.Logf("✓ %s: security context OK", pod.Name)
			} else {
				t.Logf("⚠ %s: %v", pod.Name, securityIssues)
			}
		}
	})
}

func TestNetworkPolicies(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("NetworkPoliciesExist", func(t *testing.T) {
		policies, err := client.NetworkingV1().NetworkPolicies(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list NetworkPolicies: %v", err)
		}

		if len(policies.Items) == 0 {
			t.Log("○ No NetworkPolicies found")
		} else {
			t.Logf("Found %d NetworkPolicies:", len(policies.Items))
			for _, policy := range policies.Items {
				t.Logf("  %s", policy.Name)
			}
		}
	})
}

func TestHorizontalPodAutoscalers(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("HPAConfiguration", func(t *testing.T) {
		hpas, err := client.AutoscalingV2().HorizontalPodAutoscalers(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list HPAs: %v", err)
		}

		if len(hpas.Items) == 0 {
			t.Log("○ No HPAs found")
		} else {
			t.Logf("Found %d HPAs:", len(hpas.Items))
			for _, hpa := range hpas.Items {
				t.Logf("  %s: min=%d, max=%d, current=%d, desired=%d",
					hpa.Name,
					*hpa.Spec.MinReplicas,
					hpa.Spec.MaxReplicas,
					hpa.Status.CurrentReplicas,
					hpa.Status.DesiredReplicas)
			}
		}
	})
}

func TestPodDisruptionBudgets(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("PDBConfiguration", func(t *testing.T) {
		pdbs, err := client.PolicyV1().PodDisruptionBudgets(testConfig.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list PDBs: %v", err)
		}

		if len(pdbs.Items) == 0 {
			t.Log("○ No PodDisruptionBudgets found")
		} else {
			t.Logf("Found %d PDBs:", len(pdbs.Items))
			for _, pdb := range pdbs.Items {
				t.Logf("  %s: minAvailable=%v, currentHealthy=%d, desiredHealthy=%d",
					pdb.Name,
					pdb.Spec.MinAvailable,
					pdb.Status.CurrentHealthy,
					pdb.Status.DesiredHealthy)
			}
		}
	})
}

func TestKubectlConnectivity(t *testing.T) {
	t.Run("KubectlVersion", func(t *testing.T) {
		cmd := exec.Command("kubectl", "version", "--client", "--short")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Skipf("kubectl not available: %v", err)
		}
		t.Logf("kubectl version: %s", strings.TrimSpace(string(output)))
	})

	t.Run("ClusterInfo", func(t *testing.T) {
		cmd := exec.Command("kubectl", "cluster-info")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("kubectl cluster-info failed: %v", err)
			return
		}
		lines := strings.Split(string(output), "\n")
		for _, line := range lines[:min(3, len(lines))] {
			if line != "" {
				t.Log(line)
			}
		}
	})
}

func TestDockerDesktopKubernetes(t *testing.T) {
	t.Run("ContextCheck", func(t *testing.T) {
		cmd := exec.Command("kubectl", "config", "current-context")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Skipf("Failed to get current context: %v", err)
		}

		context := strings.TrimSpace(string(output))
		t.Logf("Current context: %s", context)

		if strings.Contains(context, "docker-desktop") {
			t.Log("✓ Running on Docker Desktop Kubernetes")
		} else if strings.Contains(context, "kind") {
			t.Log("✓ Running on kind cluster")
		} else if strings.Contains(context, "minikube") {
			t.Log("✓ Running on minikube")
		} else {
			t.Logf("○ Running on non-local cluster: %s", context)
		}
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
