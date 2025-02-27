package v1alpha1

import (
	"fmt"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEventTriggeredJob(t *testing.T) {
	now := metav1.Now()

	template := EventTriggeredJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventTriggeredJob",
			APIVersion: "kubevent.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-template",
			Namespace: "default",
		},
		Spec: EventTriggeredJobSpec{
			EventSelector: &EventSelector{
				ResourceKind:     "Pod",
				NamePattern:      "web-*",
				NamespacePattern: "prod-*",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "myapp",
					},
				},
				EventTypes: []string{"CREATE", "DELETE"},
			},
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-job",
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:    "test-container",
									Image:   "busybox",
									Command: []string{"echo", "hello"},
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
		},
		Status: EventTriggeredJobStatus{
			JobsCreated:       5,
			LastTriggeredTime: &now,
			Conditions: []metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					LastTransitionTime: now,
					Reason:             "TemplateReady",
					Message:            "Template is ready to process events",
				},
			},
		},
	}

	// Verify that fields are correctly set
	if template.Spec.EventSelector.ResourceKind != "Pod" {
		t.Errorf("Expected ResourceKind to be Pod, got %s", template.Spec.EventSelector.ResourceKind)
	}

	if len(template.Spec.EventSelector.EventTypes) != 2 {
		t.Errorf("Expected 2 event types, got %d", len(template.Spec.EventSelector.EventTypes))
	}

	if template.Spec.EventSelector.NamePattern != "web-*" {
		t.Errorf("Expected NamePattern to be web-*, got %s", template.Spec.EventSelector.NamePattern)
	}

	if template.Status.JobsCreated != 5 {
		t.Errorf("Expected JobsCreated to be 5, got %d", template.Status.JobsCreated)
	}

	if template.Status.LastTriggeredTime != &now {
		t.Errorf("Expected LastTriggeredTime to be set correctly")
	}

	if len(template.Status.Conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(template.Status.Conditions))
	}

	condition := template.Status.Conditions[0]
	if condition.Type != "Ready" {
		t.Errorf("Expected condition type Ready, got %s", condition.Type)
	}
}

func TestStatusTriggeredJob(t *testing.T) {
	now := metav1.Now()

	template := EventTriggeredJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventTriggeredJob",
			APIVersion: "kubevent.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-status-template",
			Namespace: "default",
		},
		Spec: EventTriggeredJobSpec{
			StatusSelector: &StatusSelector{
				ResourceKind:     "Pod",
				NamePattern:      "web-*",
				NamespacePattern: "prod-*",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "myapp",
					},
				},
				Conditions: []StatusCondition{
					{
						Type:   "Ready",
						Status: "True",
					},
				},
			},
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-job",
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:    "test-container",
									Image:   "busybox",
									Command: []string{"echo", "hello"},
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
		},
		Status: EventTriggeredJobStatus{
			JobsCreated:       2,
			LastTriggeredTime: &now,
			Conditions: []metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					LastTransitionTime: now,
					Reason:             "TemplateReady",
					Message:            "Template is ready to process status changes",
				},
			},
		},
	}

	// Verify that fields are correctly set
	if template.Spec.StatusSelector.ResourceKind != "Pod" {
		t.Errorf("Expected ResourceKind to be Pod, got %s", template.Spec.StatusSelector.ResourceKind)
	}

	if len(template.Spec.StatusSelector.Conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(template.Spec.StatusSelector.Conditions))
	}

	if template.Spec.StatusSelector.Conditions[0].Type != "Ready" {
		t.Errorf("Expected condition Type to be Ready, got %s", template.Spec.StatusSelector.Conditions[0].Type)
	}

	if template.Spec.StatusSelector.NamePattern != "web-*" {
		t.Errorf("Expected NamePattern to be web-*, got %s", template.Spec.StatusSelector.NamePattern)
	}
}

func TestEventTriggeredJobList(t *testing.T) {
	list := EventTriggeredJobList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventTriggeredJobList",
			APIVersion: "kubevent.io/v1alpha1",
		},
		Items: []EventTriggeredJob{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "template1",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "template2",
				},
			},
		},
	}

	if len(list.Items) != 2 {
		t.Errorf("Expected 2 items in list, got %d", len(list.Items))
	}

	if list.Items[0].Name != "template1" {
		t.Errorf("Expected first item name to be template1, got %s", list.Items[0].Name)
	}

	if list.Items[1].Name != "template2" {
		t.Errorf("Expected second item name to be template2, got %s", list.Items[1].Name)
	}
}

func TestEventSelector(t *testing.T) {
	tests := []struct {
		name     string
		selector *EventSelector
		wantErr  bool
	}{
		{
			name: "valid selector",
			selector: &EventSelector{
				ResourceKind:     "Pod",
				NamePattern:      "web-*",
				NamespacePattern: "prod-*",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "myapp",
					},
				},
				EventTypes: []string{"CREATE", "DELETE"},
			},
			wantErr: false,
		},
		{
			name: "empty resource kind",
			selector: &EventSelector{
				ResourceKind: "",
				EventTypes:   []string{"CREATE"},
			},
			wantErr: true,
		},
		{
			name: "empty event types",
			selector: &EventSelector{
				ResourceKind: "Pod",
				EventTypes:   []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid event type",
			selector: &EventSelector{
				ResourceKind: "Pod",
				EventTypes:   []string{"INVALID"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEventSelector(tt.selector)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateEventSelector() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStatusSelector(t *testing.T) {
	tests := []struct {
		name     string
		selector *StatusSelector
		wantErr  bool
	}{
		{
			name: "valid selector",
			selector: &StatusSelector{
				ResourceKind:     "Pod",
				NamePattern:      "web-*",
				NamespacePattern: "prod-*",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "myapp",
					},
				},
				Conditions: []StatusCondition{
					{
						Type:   "Ready",
						Status: "True",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty resource kind",
			selector: &StatusSelector{
				ResourceKind: "",
				Conditions: []StatusCondition{
					{
						Type:   "Ready",
						Status: "True",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty conditions",
			selector: &StatusSelector{
				ResourceKind: "Pod",
				Conditions:   []StatusCondition{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStatusSelector(tt.selector)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStatusSelector() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// validateEventSelector validates the EventSelector fields
// This is a helper function for testing, in production this would be
// implemented in a webhook or admission controller
func validateEventSelector(selector *EventSelector) error {
	if selector.ResourceKind == "" {
		return fmt.Errorf("resourceKind is required")
	}

	if len(selector.EventTypes) == 0 {
		return fmt.Errorf("at least one eventType is required")
	}

	validEventTypes := map[string]bool{
		"CREATE": true,
		"UPDATE": true,
		"DELETE": true,
	}

	for _, eventType := range selector.EventTypes {
		if !validEventTypes[eventType] {
			return fmt.Errorf("invalid eventType: %s", eventType)
		}
	}

	return nil
}

// validateStatusSelector validates the StatusSelector fields
// This is a helper function for testing, in production this would be
// implemented in a webhook or admission controller
func validateStatusSelector(selector *StatusSelector) error {
	if selector.ResourceKind == "" {
		return fmt.Errorf("resourceKind is required")
	}

	if len(selector.Conditions) == 0 {
		return fmt.Errorf("at least one condition is required")
	}

	return nil
}
