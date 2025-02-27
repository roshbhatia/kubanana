package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEventTriggeredJobDeepCopy(t *testing.T) {
	original := &EventTriggeredJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-template",
			Namespace: "default",
		},
		Spec: EventTriggeredJobSpec{
			EventSelector: &EventSelector{
				ResourceKind: "Pod",
				NamePattern:  "test-*",
				EventTypes:   []string{"CREATE", "DELETE"},
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "test",
					},
				},
			},
		},
		Status: EventTriggeredJobStatus{
			JobsCreated: 5,
			LastTriggeredTime: &metav1.Time{
				Time: metav1.Now().Time,
			},
			Conditions: []metav1.Condition{
				{
					Type:   "Ready",
					Status: metav1.ConditionTrue,
				},
			},
		},
	}

	// Test DeepCopy
	copy := original.DeepCopy()

	// Verify the copy is not the same object
	if copy == original {
		t.Errorf("DeepCopy returned the same object pointer")
	}

	// Verify all fields are copied correctly
	if copy.Name != original.Name {
		t.Errorf("Expected Name %s, got %s", original.Name, copy.Name)
	}
	if copy.Namespace != original.Namespace {
		t.Errorf("Expected Namespace %s, got %s", original.Namespace, copy.Namespace)
	}
	if copy.Spec.EventSelector.ResourceKind != original.Spec.EventSelector.ResourceKind {
		t.Errorf("Expected ResourceKind %s, got %s", original.Spec.EventSelector.ResourceKind, copy.Spec.EventSelector.ResourceKind)
	}
	if copy.Spec.EventSelector.NamePattern != original.Spec.EventSelector.NamePattern {
		t.Errorf("Expected NamePattern %s, got %s", original.Spec.EventSelector.NamePattern, copy.Spec.EventSelector.NamePattern)
	}
	if len(copy.Spec.EventSelector.EventTypes) != len(original.Spec.EventSelector.EventTypes) {
		t.Errorf("Expected EventTypes length %d, got %d", len(original.Spec.EventSelector.EventTypes), len(copy.Spec.EventSelector.EventTypes))
	}
	if copy.Status.JobsCreated != original.Status.JobsCreated {
		t.Errorf("Expected JobsCreated %d, got %d", original.Status.JobsCreated, copy.Status.JobsCreated)
	}

	// Test DeepCopyObject
	objCopy := original.DeepCopyObject()
	if objCopy == nil {
		t.Errorf("DeepCopyObject returned nil")
	}
	typedObjCopy, ok := objCopy.(*EventTriggeredJob)
	if !ok {
		t.Errorf("DeepCopyObject did not return an *EventTriggeredJob")
	}
	if typedObjCopy.Name != original.Name {
		t.Errorf("Expected Name %s, got %s", original.Name, typedObjCopy.Name)
	}

	// Now change something in the original and make sure the copy is not affected
	original.Name = "changed-name"
	original.Spec.EventSelector.ResourceKind = "Deployment"
	original.Spec.EventSelector.EventTypes = append(original.Spec.EventSelector.EventTypes, "UPDATE")

	if copy.Name == original.Name {
		t.Errorf("Copy was affected by change to original Name")
	}
	if copy.Spec.EventSelector.ResourceKind == original.Spec.EventSelector.ResourceKind {
		t.Errorf("Copy was affected by change to original ResourceKind")
	}
	if len(copy.Spec.EventSelector.EventTypes) == len(original.Spec.EventSelector.EventTypes) {
		t.Errorf("Copy was affected by change to original EventTypes")
	}
}

func TestEventTriggeredJobListDeepCopy(t *testing.T) {
	original := &EventTriggeredJobList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventTriggeredJobList",
			APIVersion: "kubevent.roshanbhatia.com/v1alpha1",
		},
		ListMeta: metav1.ListMeta{
			ResourceVersion: "12345",
		},
		Items: []EventTriggeredJob{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-1",
					Namespace: "default",
				},
				Spec: EventTriggeredJobSpec{
					EventSelector: &EventSelector{
						ResourceKind: "Pod",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-2",
					Namespace: "kube-system",
				},
				Spec: EventTriggeredJobSpec{
					EventSelector: &EventSelector{
						ResourceKind: "Deployment",
					},
				},
			},
		},
	}

	// Test DeepCopy
	copy := original.DeepCopy()

	// Verify the copy is not the same object
	if copy == original {
		t.Errorf("DeepCopy returned the same object pointer")
	}

	// Verify all fields are copied correctly
	if copy.ResourceVersion != original.ResourceVersion {
		t.Errorf("Expected ResourceVersion %s, got %s", original.ResourceVersion, copy.ResourceVersion)
	}
	if len(copy.Items) != len(original.Items) {
		t.Errorf("Expected Items length %d, got %d", len(original.Items), len(copy.Items))
	}
	if copy.Items[0].Name != original.Items[0].Name {
		t.Errorf("Expected first item Name %s, got %s", original.Items[0].Name, copy.Items[0].Name)
	}

	// Test DeepCopyObject
	objCopy := original.DeepCopyObject()
	if objCopy == nil {
		t.Errorf("DeepCopyObject returned nil")
	}
	typedObjCopy, ok := objCopy.(*EventTriggeredJobList)
	if !ok {
		t.Errorf("DeepCopyObject did not return an *EventTriggeredJobList")
	}
	if len(typedObjCopy.Items) != len(original.Items) {
		t.Errorf("Expected Items length %d, got %d", len(original.Items), len(typedObjCopy.Items))
	}

	// Now change something in the original and make sure the copy is not affected
	original.Items[0].Name = "changed-name"
	original.Items = append(original.Items, EventTriggeredJob{
		ObjectMeta: metav1.ObjectMeta{
			Name: "template-3",
		},
	})

	if copy.Items[0].Name == original.Items[0].Name {
		t.Errorf("Copy was affected by change to original item Name")
	}
	if len(copy.Items) == len(original.Items) {
		t.Errorf("Copy was affected by change to original Items length")
	}
}

// Test StatusSelector DeepCopy
func TestStatusSelectorDeepCopy(t *testing.T) {
	original := &StatusSelector{
		ResourceKind: "Pod",
		NamePattern:  "test-*",
		Conditions: []StatusCondition{
			{
				Type:   "Ready",
				Status: "True",
			},
			{
				Type:     "PodScheduled",
				Status:   "True",
				Operator: "Equal",
			},
		},
	}

	// Test DeepCopy
	copy := original.DeepCopy()

	// Verify the copy is not the same object
	if copy == original {
		t.Errorf("DeepCopy returned the same object pointer")
	}

	// Verify all fields are copied correctly
	if copy.ResourceKind != original.ResourceKind {
		t.Errorf("Expected ResourceKind %s, got %s", original.ResourceKind, copy.ResourceKind)
	}
	if copy.NamePattern != original.NamePattern {
		t.Errorf("Expected NamePattern %s, got %s", original.NamePattern, copy.NamePattern)
	}
	if len(copy.Conditions) != len(original.Conditions) {
		t.Errorf("Expected Conditions length %d, got %d", len(original.Conditions), len(copy.Conditions))
	}
	if copy.Conditions[0].Type != original.Conditions[0].Type {
		t.Errorf("Expected Condition Type %s, got %s", original.Conditions[0].Type, copy.Conditions[0].Type)
	}
	if copy.Conditions[0].Status != original.Conditions[0].Status {
		t.Errorf("Expected Condition Status %s, got %s", original.Conditions[0].Status, copy.Conditions[0].Status)
	}

	// Now change something in the original and make sure the copy is not affected
	original.ResourceKind = "Deployment"
	original.Conditions[0].Status = "False"
	original.Conditions = append(original.Conditions, StatusCondition{Type: "Available", Status: "True"})

	if copy.ResourceKind == original.ResourceKind {
		t.Errorf("Copy was affected by change to original ResourceKind")
	}
	if copy.Conditions[0].Status == original.Conditions[0].Status {
		t.Errorf("Copy was affected by change to original Condition Status")
	}
	if len(copy.Conditions) == len(original.Conditions) {
		t.Errorf("Copy was affected by change to original Conditions length")
	}
}
