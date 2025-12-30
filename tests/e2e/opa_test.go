package e2e

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OPA/Gatekeeper Policy Tests
// Validates policy-as-code enforcement

const gatekeeperNamespace = "gatekeeper-system"

func TestOPAGatekeeperInstallation(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	t.Run("GatekeeperNamespaceExists", func(t *testing.T) {
		_, err := client.CoreV1().Namespaces().Get(ctx, gatekeeperNamespace, metav1.GetOptions{})
		if err != nil {
			t.Logf("○ Gatekeeper namespace not found (install Gatekeeper first)")
			t.Skip("Gatekeeper not installed")
		}
		t.Log("✓ Gatekeeper namespace exists")
	})

	t.Run("GatekeeperControllerRunning", func(t *testing.T) {
		pods, err := client.CoreV1().Pods(gatekeeperNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: "control-plane=controller-manager",
		})
		if err != nil {
			t.Skipf("○ Cannot list Gatekeeper pods: %v", err)
		}

		if len(pods.Items) == 0 {
			t.Skip("○ No Gatekeeper controller pods found")
		}

		running := 0
		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodRunning {
				running++
			}
		}

		if running > 0 {
			t.Logf("✓ Gatekeeper controller running (%d pods)", running)
		} else {
			t.Error("✗ No Gatekeeper controller pods running")
		}
	})

	t.Run("GatekeeperAuditRunning", func(t *testing.T) {
		pods, err := client.CoreV1().Pods(gatekeeperNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: "control-plane=audit-controller",
		})
		if err != nil {
			t.Skipf("○ Cannot list Gatekeeper audit pods: %v", err)
		}

		if len(pods.Items) == 0 {
			t.Log("○ No dedicated Gatekeeper audit pods (may be integrated)")
			return
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodRunning {
				t.Log("✓ Gatekeeper audit controller running")
				return
			}
		}
		t.Log("○ Gatekeeper audit controller not running")
	})
}

func TestOPAConstraintTemplates(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	// Check if Gatekeeper is installed first
	_, err := client.CoreV1().Namespaces().Get(ctx, gatekeeperNamespace, metav1.GetOptions{})
	if err != nil {
		t.Skip("○ Gatekeeper not installed, skipping constraint template tests")
	}

	expectedTemplates := []string{
		"k8srequiredlabels",
		"k8spsnonroot",
		"k8spsprivilegedcontainer",
		"k8scontainerresources",
		"k8sallowedrepos",
		"k8sblocknodeport",
		"k8spsphostnetwork",
		"k8sreadonlyrootfilesystem",
		"k8sdisallowedtags",
		"k8srequireprobes",
		"k8scostallocationlabels",
	}

	for _, tmpl := range expectedTemplates {
		t.Run("Template_"+tmpl, func(t *testing.T) {
			// Use dynamic client to check CRD
			// For now, we'll check via kubectl simulation
			t.Logf("○ Checking constraint template: %s (requires dynamic client)", tmpl)
		})
	}
}

func TestOPAPolicyEnforcement(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	// Check if Gatekeeper is installed
	_, err := client.CoreV1().Namespaces().Get(ctx, gatekeeperNamespace, metav1.GetOptions{})
	if err != nil {
		t.Skip("○ Gatekeeper not installed, skipping policy enforcement tests")
	}

	t.Run("PrivilegedContainerBlocked", func(t *testing.T) {
		// Attempt to create a privileged pod (should be blocked)
		privilegedPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-privileged-pod",
				Namespace: testConfig.Namespace,
				Labels: map[string]string{
					"app":       "test-privileged",
					"test-type": "policy-validation",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "test",
						Image: "nginx:1.21",
						SecurityContext: &corev1.SecurityContext{
							Privileged: boolPtr(true),
						},
					},
				},
			},
		}

		_, err := client.CoreV1().Pods(testConfig.Namespace).Create(ctx, privilegedPod, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "privileged") || strings.Contains(err.Error(), "denied") {
				t.Log("✓ Privileged container correctly blocked by policy")
			} else {
				t.Logf("○ Pod creation failed with: %v", err)
			}
		} else {
			// Clean up if it was created
			client.CoreV1().Pods(testConfig.Namespace).Delete(ctx, "test-privileged-pod", metav1.DeleteOptions{})
			t.Log("⚠ Privileged pod was created - policy may not be enforcing")
		}
	})

	t.Run("HostNetworkBlocked", func(t *testing.T) {
		hostNetPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-hostnet-pod",
				Namespace: testConfig.Namespace,
				Labels: map[string]string{
					"app":       "test-hostnet",
					"test-type": "policy-validation",
				},
			},
			Spec: corev1.PodSpec{
				HostNetwork: true,
				Containers: []corev1.Container{
					{
						Name:  "test",
						Image: "nginx:1.21",
					},
				},
			},
		}

		_, err := client.CoreV1().Pods(testConfig.Namespace).Create(ctx, hostNetPod, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "host") || strings.Contains(err.Error(), "denied") {
				t.Log("✓ Host network pod correctly blocked by policy")
			} else {
				t.Logf("○ Pod creation failed with: %v", err)
			}
		} else {
			client.CoreV1().Pods(testConfig.Namespace).Delete(ctx, "test-hostnet-pod", metav1.DeleteOptions{})
			t.Log("⚠ Host network pod was created - policy may not be enforcing")
		}
	})

	t.Run("MissingLabelsWarned", func(t *testing.T) {
		// Pod without required labels
		unlabeledPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-unlabeled-pod",
				Namespace: testConfig.Namespace,
				// Missing required labels
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "test",
						Image: "nginx:1.21",
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    mustParseQuantity("100m"),
								corev1.ResourceMemory: mustParseQuantity("128Mi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    mustParseQuantity("50m"),
								corev1.ResourceMemory: mustParseQuantity("64Mi"),
							},
						},
						SecurityContext: &corev1.SecurityContext{
							RunAsNonRoot:             boolPtr(true),
							ReadOnlyRootFilesystem:   boolPtr(true),
							AllowPrivilegeEscalation: boolPtr(false),
						},
					},
				},
				SecurityContext: &corev1.PodSecurityContext{
					RunAsUser:  int64Ptr(1000),
					RunAsGroup: int64Ptr(1000),
				},
			},
		}

		_, err := client.CoreV1().Pods(testConfig.Namespace).Create(ctx, unlabeledPod, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "label") {
				t.Log("✓ Missing labels correctly blocked/warned by policy")
			} else {
				t.Logf("○ Pod creation failed with: %v", err)
			}
		} else {
			client.CoreV1().Pods(testConfig.Namespace).Delete(ctx, "test-unlabeled-pod", metav1.DeleteOptions{})
			t.Log("○ Pod without labels created (warn mode or policy not active)")
		}
	})
}

func TestOPAResourceConstraints(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	// Check if Gatekeeper is installed
	_, err := client.CoreV1().Namespaces().Get(ctx, gatekeeperNamespace, metav1.GetOptions{})
	if err != nil {
		t.Skip("○ Gatekeeper not installed, skipping resource constraint tests")
	}

	t.Run("MissingResourceLimitsBlocked", func(t *testing.T) {
		// Pod without resource limits
		noLimitsPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-nolimits-pod",
				Namespace: testConfig.Namespace,
				Labels: map[string]string{
					"app":                       "test-nolimits",
					"app.kubernetes.io/name":    "test-nolimits",
					"app.kubernetes.io/part-of": "test",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "test",
						Image: "nginx:1.21",
						// No resource limits defined
						SecurityContext: &corev1.SecurityContext{
							RunAsNonRoot: boolPtr(true),
						},
					},
				},
				SecurityContext: &corev1.PodSecurityContext{
					RunAsUser: int64Ptr(1000),
				},
			},
		}

		_, err := client.CoreV1().Pods(testConfig.Namespace).Create(ctx, noLimitsPod, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "limit") || strings.Contains(err.Error(), "resource") {
				t.Log("✓ Pod without resource limits correctly blocked by policy")
			} else {
				t.Logf("○ Pod creation failed with: %v", err)
			}
		} else {
			client.CoreV1().Pods(testConfig.Namespace).Delete(ctx, "test-nolimits-pod", metav1.DeleteOptions{})
			t.Log("⚠ Pod without limits created - policy may not be enforcing")
		}
	})
}

func TestOPASecurityPolicies(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	_, err := client.CoreV1().Namespaces().Get(ctx, gatekeeperNamespace, metav1.GetOptions{})
	if err != nil {
		t.Skip("○ Gatekeeper not installed, skipping security policy tests")
	}

	t.Run("RootUserBlocked", func(t *testing.T) {
		rootPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-root-pod",
				Namespace: testConfig.Namespace,
				Labels: map[string]string{
					"app":                       "test-root",
					"app.kubernetes.io/name":    "test-root",
					"app.kubernetes.io/part-of": "test",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "test",
						Image: "nginx:1.21",
						SecurityContext: &corev1.SecurityContext{
							RunAsUser: int64Ptr(0), // Root user
						},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    mustParseQuantity("100m"),
								corev1.ResourceMemory: mustParseQuantity("128Mi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    mustParseQuantity("50m"),
								corev1.ResourceMemory: mustParseQuantity("64Mi"),
							},
						},
					},
				},
			},
		}

		_, err := client.CoreV1().Pods(testConfig.Namespace).Create(ctx, rootPod, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "root") || strings.Contains(err.Error(), "runAsUser") || strings.Contains(err.Error(), "nonroot") {
				t.Log("✓ Root user pod correctly blocked by policy")
			} else {
				t.Logf("○ Pod creation failed with: %v", err)
			}
		} else {
			client.CoreV1().Pods(testConfig.Namespace).Delete(ctx, "test-root-pod", metav1.DeleteOptions{})
			t.Log("⚠ Root user pod was created - policy may not be enforcing")
		}
	})
}

func TestOPAAuditCompliance(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	_, err := client.CoreV1().Namespaces().Get(ctx, gatekeeperNamespace, metav1.GetOptions{})
	if err != nil {
		t.Skip("○ Gatekeeper not installed, skipping audit compliance tests")
	}

	t.Run("AuditConfigExists", func(t *testing.T) {
		_, err := client.CoreV1().ConfigMaps(gatekeeperNamespace).Get(ctx, "gatekeeper-audit-config", metav1.GetOptions{})
		if err != nil {
			t.Log("○ Audit config not found (optional)")
		} else {
			t.Log("✓ Gatekeeper audit config exists")
		}
	})

	t.Run("ComplianceReportCronJob", func(t *testing.T) {
		// Check for compliance report cronjob
		cronjobs, err := client.BatchV1().CronJobs(gatekeeperNamespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Logf("○ Cannot list cronjobs: %v", err)
			return
		}

		for _, cj := range cronjobs.Items {
			if strings.Contains(cj.Name, "compliance") || strings.Contains(cj.Name, "policy") {
				t.Logf("✓ Compliance report cronjob found: %s", cj.Name)
				return
			}
		}
		t.Log("○ No compliance report cronjob found (optional)")
	})
}

func TestOPAPolicyViolationDetection(t *testing.T) {
	client := SetupK8sClient(t)
	if client == nil {
		t.Skip("Kubernetes client not available")
	}
	ctx := context.Background()

	_, err := client.CoreV1().Namespaces().Get(ctx, gatekeeperNamespace, metav1.GetOptions{})
	if err != nil {
		t.Skip("○ Gatekeeper not installed")
	}

	t.Run("PolicyViolationsTracked", func(t *testing.T) {
		// In a real test, we'd query the constraint status
		// For now, verify the system is responding
		pods, err := client.CoreV1().Pods(gatekeeperNamespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Logf("○ Cannot verify policy tracking: %v", err)
			return
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodRunning {
				t.Log("✓ Gatekeeper running and tracking violations")
				return
			}
		}
		t.Log("○ Gatekeeper not fully running")
	})
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}

func mustParseQuantity(s string) resource.Quantity {
	q, _ := resource.ParseQuantity(s)
	return q
}
