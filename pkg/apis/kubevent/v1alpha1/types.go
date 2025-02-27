package v1alpha1

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EventTriggeredJob defines a job template that gets executed when specific Kubernetes events are fired
type EventTriggeredJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EventTriggeredJobSpec   `json:"spec"`
	Status EventTriggeredJobStatus `json:"status,omitempty"`
}

// EventTriggeredJobSpec defines the specification for an EventTriggeredJob
type EventTriggeredJobSpec struct {
	// EventSelector specifies which events should trigger job creation
	// +optional
	EventSelector *EventSelector `json:"eventSelector,omitempty"`

	// StatusSelector specifies which resource status conditions should trigger job creation
	// +optional
	StatusSelector *StatusSelector `json:"statusSelector,omitempty"`

	// JobTemplate is the template for the job to be created when an event is triggered
	JobTemplate batchv1.JobTemplateSpec `json:"jobTemplate"`
}

// EventSelector defines criteria for selecting which events trigger job creation
type EventSelector struct {
	// ResourceKind is the kind of the resource to watch (e.g., "Pod", "Deployment")
	ResourceKind string `json:"resourceKind"`

	// NamePattern is a glob pattern to filter resource names
	// +optional
	NamePattern string `json:"namePattern,omitempty"`

	// NamespacePattern is a glob pattern to filter namespaces
	// +optional
	NamespacePattern string `json:"namespacePattern,omitempty"`

	// LabelSelector is a label selector to filter resources
	// +optional
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`

	// EventTypes are the types of events to watch for (e.g., "CREATE", "UPDATE", "DELETE")
	EventTypes []string `json:"eventTypes"`
}

// EventTriggeredJobStatus defines the observed state of EventTriggeredJob
type EventTriggeredJobStatus struct {
	// JobsCreated is the number of jobs created by this template
	JobsCreated int64 `json:"jobsCreated"`

	// LastTriggeredTime is the last time a job was triggered
	// +optional
	LastTriggeredTime *metav1.Time `json:"lastTriggeredTime,omitempty"`

	// Conditions represent the latest available observations of the template's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StatusCondition describes a condition that should match a resource's status
type StatusCondition struct {
	// Type is the condition type to check (e.g., "Ready", "Available")
	Type string `json:"type"`

	// Status is the status value to match (e.g., "True", "False", "Unknown")
	Status string `json:"status"`

	// Operator specifies how to compare the condition (default: "Equal")
	// +optional
	Operator string `json:"operator,omitempty"`
}

// StatusSelector defines criteria for selecting resources based on their status conditions
type StatusSelector struct {
	// ResourceKind is the kind of the resource to watch (e.g., "Pod", "Deployment")
	ResourceKind string `json:"resourceKind"`

	// NamePattern is a glob pattern to filter resource names
	// +optional
	NamePattern string `json:"namePattern,omitempty"`

	// NamespacePattern is a glob pattern to filter namespaces
	// +optional
	NamespacePattern string `json:"namespacePattern,omitempty"`

	// LabelSelector is a label selector to filter resources
	// +optional
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`

	// Conditions are the status conditions to match
	Conditions []StatusCondition `json:"conditions"`
}

// EventTriggeredJobList contains a list of EventTriggeredJob
type EventTriggeredJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EventTriggeredJob `json:"items"`
}
