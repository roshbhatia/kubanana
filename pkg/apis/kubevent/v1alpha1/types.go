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
	EventSelector EventSelector `json:"eventSelector"`

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

// EventTriggeredJobList contains a list of EventTriggeredJob
type EventTriggeredJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EventTriggeredJob `json:"items"`
}
