package controller

import (
	"testing"

	"github.com/roshbhatia/kubevent/pkg/apis/kubevent/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func TestMatchResourceKind(t *testing.T) {
	tests := []struct {
		name         string
		template     v1alpha1.EventTriggeredJob
		resourceKind string
		shouldMatch  bool
	}{
		{
			name: "exact match",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						ResourceKind: "Pod",
					},
				},
			},
			resourceKind: "Pod",
			shouldMatch:  true,
		},
		{
			name: "no match",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						ResourceKind: "Pod",
					},
				},
			},
			resourceKind: "ConfigMap",
			shouldMatch:  false,
		},
		{
			name: "case sensitive match",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						ResourceKind: "Pod",
					},
				},
			},
			resourceKind: "pod",
			shouldMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := testMatchResourceKind(&tt.template, tt.resourceKind)
			if match != tt.shouldMatch {
				t.Errorf("testMatchResourceKind() = %v, want %v", match, tt.shouldMatch)
			}
		})
	}
}

func TestMatchNamePattern(t *testing.T) {
	tests := []struct {
		name         string
		template     v1alpha1.EventTriggeredJob
		resourceName string
		shouldMatch  bool
	}{
		{
			name: "exact match",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						NamePattern: "test-pod",
					},
				},
			},
			resourceName: "test-pod",
			shouldMatch:  true,
		},
		{
			name: "glob match",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						NamePattern: "test-*",
					},
				},
			},
			resourceName: "test-pod",
			shouldMatch:  true,
		},
		{
			name: "no match",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						NamePattern: "web-*",
					},
				},
			},
			resourceName: "test-pod",
			shouldMatch:  false,
		},
		{
			name: "empty pattern should match anything",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						NamePattern: "",
					},
				},
			},
			resourceName: "test-pod",
			shouldMatch:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := testMatchNamePattern(&tt.template, tt.resourceName)
			if match != tt.shouldMatch {
				t.Errorf("testMatchNamePattern() = %v, want %v", match, tt.shouldMatch)
			}
		})
	}
}

func TestMatchNamespacePattern(t *testing.T) {
	tests := []struct {
		name              string
		template          v1alpha1.EventTriggeredJob
		resourceNamespace string
		shouldMatch       bool
	}{
		{
			name: "exact match",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						NamespacePattern: "default",
					},
				},
			},
			resourceNamespace: "default",
			shouldMatch:       true,
		},
		{
			name: "glob match",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						NamespacePattern: "prod-*",
					},
				},
			},
			resourceNamespace: "prod-east",
			shouldMatch:       true,
		},
		{
			name: "no match",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						NamespacePattern: "prod-*",
					},
				},
			},
			resourceNamespace: "dev",
			shouldMatch:       false,
		},
		{
			name: "empty pattern should match anything",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						NamespacePattern: "",
					},
				},
			},
			resourceNamespace: "default",
			shouldMatch:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := testMatchNamespacePattern(&tt.template, tt.resourceNamespace)
			if match != tt.shouldMatch {
				t.Errorf("testMatchNamespacePattern() = %v, want %v", match, tt.shouldMatch)
			}
		})
	}
}

func TestMatchLabelSelector(t *testing.T) {
	tests := []struct {
		name           string
		template       v1alpha1.EventTriggeredJob
		resourceLabels map[string]string
		shouldMatch    bool
	}{
		{
			name: "match with labels",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "nginx",
							},
						},
					},
				},
			},
			resourceLabels: map[string]string{
				"app": "nginx",
				"env": "prod",
			},
			shouldMatch: true,
		},
		{
			name: "no match with labels",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "nginx",
							},
						},
					},
				},
			},
			resourceLabels: map[string]string{
				"app": "apache",
				"env": "prod",
			},
			shouldMatch: false,
		},
		{
			name: "no labels on resource",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "nginx",
							},
						},
					},
				},
			},
			resourceLabels: map[string]string{},
			shouldMatch:    false,
		},
		{
			name: "no selector on template",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						LabelSelector: nil,
					},
				},
			},
			resourceLabels: map[string]string{
				"app": "nginx",
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchLabels := labels.Set(tt.resourceLabels)
			match := testMatchLabelSelector(&tt.template, matchLabels)
			if match != tt.shouldMatch {
				t.Errorf("testMatchLabelSelector() = %v, want %v", match, tt.shouldMatch)
			}
		})
	}
}

func TestMatchEventType(t *testing.T) {
	tests := []struct {
		name        string
		template    v1alpha1.EventTriggeredJob
		eventType   string
		shouldMatch bool
	}{
		{
			name: "match CREATE event",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						EventTypes: []string{"CREATE", "DELETE"},
					},
				},
			},
			eventType:   "CREATE",
			shouldMatch: true,
		},
		{
			name: "match DELETE event",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						EventTypes: []string{"CREATE", "DELETE"},
					},
				},
			},
			eventType:   "DELETE",
			shouldMatch: true,
		},
		{
			name: "no match UPDATE event",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						EventTypes: []string{"CREATE", "DELETE"},
					},
				},
			},
			eventType:   "UPDATE",
			shouldMatch: false,
		},
		{
			name: "empty event types should not match anything",
			template: v1alpha1.EventTriggeredJob{
				Spec: v1alpha1.EventTriggeredJobSpec{
					EventSelector: v1alpha1.EventSelector{
						EventTypes: []string{},
					},
				},
			},
			eventType:   "CREATE",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := testMatchEventType(&tt.template, tt.eventType)
			if match != tt.shouldMatch {
				t.Errorf("testMatchEventType() = %v, want %v", match, tt.shouldMatch)
			}
		})
	}
}

// These are the functions we're testing
func testMatchResourceKind(template *v1alpha1.EventTriggeredJob, resourceKind string) bool {
	return template.Spec.EventSelector.ResourceKind == resourceKind
}

func testMatchNamePattern(template *v1alpha1.EventTriggeredJob, resourceName string) bool {
	pattern := template.Spec.EventSelector.NamePattern
	if pattern == "" {
		return true // Empty pattern matches anything
	}
	// Simple glob matching
	// TODO: Implement proper glob matching with path.Match
	return resourceName == pattern ||
		(pattern[len(pattern)-1] == '*' && len(resourceName) >= len(pattern)-1 &&
			resourceName[:len(pattern)-1] == pattern[:len(pattern)-1])
}

func testMatchNamespacePattern(template *v1alpha1.EventTriggeredJob, resourceNamespace string) bool {
	pattern := template.Spec.EventSelector.NamespacePattern
	if pattern == "" {
		return true // Empty pattern matches anything
	}
	// Simple glob matching
	// TODO: Implement proper glob matching with path.Match
	return resourceNamespace == pattern ||
		(pattern[len(pattern)-1] == '*' && len(resourceNamespace) >= len(pattern)-1 &&
			resourceNamespace[:len(pattern)-1] == pattern[:len(pattern)-1])
}

func testMatchLabelSelector(template *v1alpha1.EventTriggeredJob, resourceLabels labels.Labels) bool {
	if template.Spec.EventSelector.LabelSelector == nil {
		return true // No label selector matches anything
	}

	selector, err := metav1.LabelSelectorAsSelector(template.Spec.EventSelector.LabelSelector)
	if err != nil {
		return false
	}

	return selector.Matches(resourceLabels)
}

func testMatchEventType(template *v1alpha1.EventTriggeredJob, eventType string) bool {
	for _, t := range template.Spec.EventSelector.EventTypes {
		if t == eventType {
			return true
		}
	}
	return false
}
