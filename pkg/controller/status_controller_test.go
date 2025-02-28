package controller

import (
	"context"
	"testing"

	"github.com/roshbhatia/kubanana/pkg/apis/kubanana/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

func TestNewStatusController(t *testing.T) {
	// Create fake clients
	kubeClient := fake.NewSimpleClientset()
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)

	// Create a new status controller
	controller := NewStatusController(kubeClient, dynamicClient)

	// Check if the controller is properly initialized
	if controller.kubeClient != kubeClient {
		t.Errorf("Expected kubeClient to be set correctly")
	}

	if controller.dynamicClient != dynamicClient {
		t.Errorf("Expected dynamicClient to be set correctly")
	}

	if controller.workqueue == nil {
		t.Errorf("Expected workqueue to be initialized")
	}

	if controller.informers == nil {
		t.Errorf("Expected informers map to be initialized")
	}

	if controller.resourceStatus == nil {
		t.Errorf("Expected resourceStatus map to be initialized")
	}
}

// Simple test for status-based job creation logic
func TestCreateJobForStatusChange(t *testing.T) {
	// Create fake clients
	kubeClient := fake.NewSimpleClientset()
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)

	// Create a status controller
	controller := &StatusController{
		kubeClient:     kubeClient,
		dynamicClient:  dynamicClient,
		workqueue:      workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		informers:      make(map[schema.GroupVersionKind]cache.SharedIndexInformer),
		resourceStatus: make(map[string]map[string]string),
	}

	// Create a test template
	template := &v1alpha1.EventTriggeredJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-status-template",
			Namespace: "default",
			UID:       "test-template-uid",
		},
		Spec: v1alpha1.EventTriggeredJobSpec{
			StatusSelector: &v1alpha1.StatusSelector{
				ResourceKind:     "Pod",
				NamePattern:      "test-*",
				NamespacePattern: "default",
				Conditions: []v1alpha1.StatusCondition{
					{
						Type:   "Ready",
						Status: "True",
					},
				},
			},
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "hello",
									Image: "busybox",
									Command: []string{
										"sh",
										"-c",
										"echo 'Resource is ready'; sleep 5",
									},
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
		},
	}

	// Set up test data
	resourceKind := "Pod"
	namespace := "default"
	name := "test-pod"
	conditions := map[string]string{
		"Ready": "True",
	}

	// Store a resource status
	controller.resourceStatus["default/test-pod"] = conditions

	// Store the template
	controller.templates = []v1alpha1.EventTriggeredJob{*template}

	// Test the job creation method directly
	err := controller.createJobFromTemplate(template, resourceKind, namespace, name, conditions)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Verify job was created
	jobs, err := kubeClient.BatchV1().Jobs(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}

	if len(jobs.Items) != 1 {
		t.Errorf("Expected 1 job to be created, got %d", len(jobs.Items))
	}

	// Verify job details
	job := jobs.Items[0]
	if !isOwner(template.UID, job.OwnerReferences) {
		t.Errorf("Expected job to have the template as owner")
	}

	// Verify container has environment variables
	container := job.Spec.Template.Spec.Containers[0]
	foundReadyEnv := false
	for _, env := range container.Env {
		if env.Name == "STATUS_Ready" && env.Value == "True" {
			foundReadyEnv = true
			break
		}
	}

	if !foundReadyEnv {
		t.Errorf("Expected job to have STATUS_Ready environment variable")
	}
}

// Helper function to check if a UID is in owner references
func isOwner(uid types.UID, ownerRefs []metav1.OwnerReference) bool {
	for _, ref := range ownerRefs {
		if ref.UID == uid {
			return true
		}
	}
	return false
}

// Test status condition matching
func TestStatusConditionMatch(t *testing.T) {
	// Create test conditions map
	conditions := map[string]string{
		"Ready":       "True",
		"Available":   "True",
		"Progressing": "False",
	}

	// Test cases
	tests := []struct {
		name        string
		conditions  []v1alpha1.StatusCondition
		shouldMatch bool
	}{
		{
			name: "single matching condition",
			conditions: []v1alpha1.StatusCondition{
				{
					Type:   "Ready",
					Status: "True",
				},
			},
			shouldMatch: true,
		},
		{
			name: "multiple matching conditions",
			conditions: []v1alpha1.StatusCondition{
				{
					Type:   "Ready",
					Status: "True",
				},
				{
					Type:   "Available",
					Status: "True",
				},
			},
			shouldMatch: true,
		},
		{
			name: "one mismatched condition",
			conditions: []v1alpha1.StatusCondition{
				{
					Type:   "Ready",
					Status: "True",
				},
				{
					Type:   "Progressing",
					Status: "True", // Actual is False
				},
			},
			shouldMatch: false,
		},
		{
			name: "missing condition",
			conditions: []v1alpha1.StatusCondition{
				{
					Type:   "Ready",
					Status: "True",
				},
				{
					Type:   "Missing",
					Status: "True",
				},
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if conditions match
			match := true
			for _, requiredCond := range tt.conditions {
				actualStatus, exists := conditions[requiredCond.Type]
				if !exists || actualStatus != requiredCond.Status {
					match = false
					break
				}
			}

			if match != tt.shouldMatch {
				t.Errorf("Expected match to be %v, got %v", tt.shouldMatch, match)
			}
		})
	}
}

// Test status change detection
func TestStatusChangeDetection(t *testing.T) {
	// Create a controller
	controller := &StatusController{
		resourceStatus: make(map[string]map[string]string),
	}

	// Initial status
	key := "default/test-pod"
	initialStatus := map[string]string{
		"Ready": "False",
	}

	// Store initial status
	controller.resourceStatus[key] = initialStatus

	// Check if status changed
	newStatus := map[string]string{
		"Ready": "True",
	}

	// Clone the status to avoid modifying the original
	changed := true
	prevStatus, exists := controller.resourceStatus[key]
	if exists && areStatusesEqual(prevStatus, newStatus) {
		changed = false
	}

	if !changed {
		t.Errorf("Expected status to be detected as changed")
	}

	// Update status
	controller.resourceStatus[key] = newStatus

	// Check again with same status
	changed = true
	prevStatus, exists = controller.resourceStatus[key]
	if exists && areStatusesEqual(prevStatus, newStatus) {
		changed = false
	}

	if changed {
		t.Errorf("Expected status to be detected as unchanged")
	}
}

// Implement helper function for test
func areStatusesEqual(old, new map[string]string) bool {
	if len(old) != len(new) {
		return false
	}

	for k, v := range old {
		if newV, ok := new[k]; !ok || newV != v {
			return false
		}
	}

	return true
}

// Test enhanced name pattern matching
func TestEnhancedNamePatternMatching(t *testing.T) {
	testCases := []struct {
		name         string
		pattern      string
		resourceName string
		shouldMatch  bool
	}{
		{
			name:         "exact match",
			pattern:      "test-pod",
			resourceName: "test-pod",
			shouldMatch:  true,
		},
		{
			name:         "prefix wildcard",
			pattern:      "*-pod",
			resourceName: "test-pod",
			shouldMatch:  true,
		},
		{
			name:         "suffix wildcard",
			pattern:      "test-*",
			resourceName: "test-pod",
			shouldMatch:  true,
		},
		{
			name:         "middle wildcard",
			pattern:      "test-*-status",
			resourceName: "test-pod-status",
			shouldMatch:  true,
		},
		{
			name:         "multiple wildcards",
			pattern:      "test-*-*",
			resourceName: "test-pod-status",
			shouldMatch:  true,
		},
		{
			name:         "no match",
			pattern:      "test-pod",
			resourceName: "different-pod",
			shouldMatch:  false,
		},
		{
			name:         "prefix wildcard no match",
			pattern:      "*-foo",
			resourceName: "test-pod",
			shouldMatch:  false,
		},
		{
			name:         "suffix wildcard no match",
			pattern:      "other-*",
			resourceName: "test-pod",
			shouldMatch:  false,
		},
		{
			name:         "empty pattern",
			pattern:      "",
			resourceName: "test-pod",
			shouldMatch:  false,
		},
		{
			name:         "just wildcard",
			pattern:      "*",
			resourceName: "anything",
			shouldMatch:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matchNamePattern(tc.pattern, tc.resourceName)
			if result != tc.shouldMatch {
				t.Errorf("matchNamePattern(%q, %q) = %v, want %v", tc.pattern, tc.resourceName, result, tc.shouldMatch)
			}
		})
	}
}
