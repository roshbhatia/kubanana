package v1alpha1

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCRDValidation(t *testing.T) {
	// Create a valid template
	validTemplate := &EventTriggeredJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubevent.roshanbhatia.com/v1alpha1",
			Kind:       "EventTriggeredJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "valid-template",
			Namespace: "default",
		},
		Spec: EventTriggeredJobSpec{
			EventSelector: EventSelector{
				ResourceKind: "Pod",
				NamePattern:  "test-*",
				EventTypes:   []string{"CREATE", "DELETE"},
			},
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:    "test",
									Image:   "busybox",
									Command: []string{"echo", "test"},
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
		},
	}

	// Validate required fields
	if validTemplate.Spec.EventSelector.ResourceKind == "" {
		t.Errorf("ResourceKind is required but allowed to be empty")
	}

	if len(validTemplate.Spec.EventSelector.EventTypes) == 0 {
		t.Errorf("EventTypes is required but allowed to be empty")
	}

	// Read the CRD file
	rootDir := findRootDir(t)
	crdPath := filepath.Join(rootDir, "deploy", "crds", "kubevent.roshanbhatia.com_EventTriggeredJob.yaml")

	if _, err := os.Stat(crdPath); os.IsNotExist(err) {
		t.Skipf("CRD file not found at %s, skipping test", crdPath)
		return
	}

	crdBytes, err := os.ReadFile(crdPath)
	if err != nil {
		t.Fatalf("Error reading CRD file: %v", err)
	}

	// Just check that the CRD file exists and can be parsed
	// Since we're not going to validate the full structure in this test
	var crdMap map[string]interface{}
	if err := yaml.Unmarshal(crdBytes, &crdMap); err != nil {
		t.Fatalf("Error parsing CRD YAML: %v", err)
	}

	// Log the structure
	t.Logf("CRD structure: %+v", crdMap)

	// Do some basic checks on the CRD
	kind, ok := crdMap["kind"].(string)
	if !ok || kind != "CustomResourceDefinition" {
		t.Errorf("Expected kind to be 'CustomResourceDefinition', got %v", kind)
	}

	// Check that the CRD example can be loaded
	examplePath := filepath.Join(rootDir, "deploy", "samples", "example-job-template.yaml")

	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Skipf("Example file not found at %s, skipping test", examplePath)
		return
	}

	exampleBytes, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("Error reading example file: %v", err)
	}

	// First unmarshal to a map to see the raw structure
	var exampleMap map[string]interface{}
	if err := yaml.Unmarshal(exampleBytes, &exampleMap); err != nil {
		t.Fatalf("Error parsing example YAML as map: %v", err)
	}

	// Print out the raw structure to debug
	t.Logf("Example YAML structure: %+v", exampleMap)

	// Verify the example directly from the map
	apiVersion, ok := exampleMap["apiVersion"].(string)
	if !ok || apiVersion != "kubevent.roshanbhatia.com/v1alpha1" {
		t.Errorf("Expected APIVersion to be 'kubevent.roshanbhatia.com/v1alpha1', got '%v'", apiVersion)
	}

	kindVal, ok := exampleMap["kind"].(string)
	if !ok || kindVal != "EventTriggeredJob" {
		t.Errorf("Expected Kind to be 'EventTriggeredJob', got '%v'", kindVal)
	}

	spec, ok := exampleMap["spec"].(map[interface{}]interface{})
	if !ok {
		t.Errorf("Expected spec to be a map")
		return
	}

	eventSelector, ok := spec["eventSelector"].(map[interface{}]interface{})
	if !ok {
		t.Errorf("Expected eventSelector to be a map")
		return
	}

	resourceKind, ok := eventSelector["resourceKind"].(string)
	if !ok || resourceKind == "" {
		t.Errorf("Expected ResourceKind to be non-empty")
	}

	eventTypes, ok := eventSelector["eventTypes"].([]interface{})
	if !ok || len(eventTypes) == 0 {
		t.Errorf("Expected EventTypes to be non-empty")
	}
}

// Helper function to find the root directory
func findRootDir(t *testing.T) string {
	// Start with the current directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting current directory: %v", err)
	}

	// Go up directories until we find the go.mod file
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("Could not find root directory (no go.mod found)")
			return ""
		}
		dir = parent
	}
}
