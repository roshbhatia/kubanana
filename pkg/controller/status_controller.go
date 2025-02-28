package controller

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/roshbhatia/kubanana/pkg/apis/kubanana/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// StatusController watches resource status changes and triggers jobs when conditions match
type StatusController struct {
	kubeClient     kubernetes.Interface
	dynamicClient  dynamic.Interface
	workqueue      workqueue.RateLimitingInterface
	informers      map[schema.GroupVersionKind]cache.SharedIndexInformer
	templates      []v1alpha1.EventTriggeredJob
	resourceStatus map[string]map[string]string // Tracks resource statuses
}

// NewStatusController creates a new StatusController
func NewStatusController(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) *StatusController {
	controller := &StatusController{
		kubeClient:     kubeClient,
		dynamicClient:  dynamicClient,
		workqueue:      workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		informers:      make(map[schema.GroupVersionKind]cache.SharedIndexInformer),
		resourceStatus: make(map[string]map[string]string),
	}

	// Load initial templates if not in a test environment
	_, isTest := os.LookupEnv("TEST_MODE")
	if !isTest {
		controller.refreshTemplates()
	}

	return controller
}

// refreshTemplates loads all EventTriggeredJobs with StatusSelector
func (c *StatusController) refreshTemplates() {
	// Skip for testing when kubeClient might not be fully initialized
	if c.kubeClient == nil {
		klog.Warningf("Skipping refresh templates, kubeClient is nil")
		return
	}

	templateList := &v1alpha1.EventTriggeredJobList{}
	err := c.kubeClient.CoreV1().RESTClient().
		Get().
		AbsPath("/apis/kubanana.roshanbhatia.com/v1alpha1/eventtriggeredjobs").
		Do(context.Background()).
		Into(templateList)

	if err != nil {
		klog.Errorf("Failed to get templates: %v", err)
		return
	}

	// Filter templates that have a StatusSelector
	c.templates = []v1alpha1.EventTriggeredJob{}
	watchedKinds := make(map[string]bool)

	for _, template := range templateList.Items {
		if template.Spec.StatusSelector != nil {
			c.templates = append(c.templates, template)
			watchedKinds[template.Spec.StatusSelector.ResourceKind] = true
		}
	}

	// Setup informers for each kind of resource we need to watch
	for kind := range watchedKinds {
		c.setupInformerForKind(kind)
	}

	klog.Infof("Loaded %d status-based templates", len(c.templates))
}

// setupInformerForKind creates an informer for a specific resource kind
func (c *StatusController) setupInformerForKind(kind string) {
	// Map common resource kinds to their correct GVR
	var gvr schema.GroupVersionResource

	switch kind {
	case "Pod":
		gvr = schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "pods",
		}
	case "Deployment":
		gvr = schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}
	case "StatefulSet":
		gvr = schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "statefulsets",
		}
	case "DaemonSet":
		gvr = schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "daemonsets",
		}
	case "Job":
		gvr = schema.GroupVersionResource{
			Group:    "batch",
			Version:  "v1",
			Resource: "jobs",
		}
	case "CronJob":
		gvr = schema.GroupVersionResource{
			Group:    "batch",
			Version:  "v1",
			Resource: "cronjobs",
		}
	default:
		// Default to core v1 resources with simple pluralization
		gvr = schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: kindToResource(kind),
		}
	}

	// Create GVK from GVR for cache key
	gvk := schema.GroupVersionKind{
		Group:   gvr.Group,
		Version: gvr.Version,
		Kind:    kind,
	}

	// Check if we already have an informer for this GVK
	if _, exists := c.informers[gvk]; exists {
		return
	}

	// Create a dynamic list/watch for the resource
	listFunc := func(options metav1.ListOptions) (runtime.Object, error) {
		return c.dynamicClient.Resource(gvr).Namespace("").List(context.Background(), options)
	}

	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		return c.dynamicClient.Resource(gvr).Namespace("").Watch(context.Background(), options)
	}

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  listFunc,
			WatchFunc: watchFunc,
		},
		&unstructured.Unstructured{},
		0,
		cache.Indexers{},
	)

	// Add event handlers
	// Using AddEventHandlerWithResyncPeriod which doesn't return a value in our version
	informer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: c.handleObject,
		UpdateFunc: func(old, new interface{}) {
			c.handleObject(new)
		},
		DeleteFunc: c.handleObject,
	}, 0)

	c.informers[gvk] = informer
	klog.Infof("Set up informer for resource kind: %s", kind)
}

// Run starts the controller
func (c *StatusController) Run(workers int, stopCh <-chan struct{}) error {
	defer c.workqueue.ShutDown()

	klog.Info("Starting status controller")

	// Start all the informers
	for gvk, informer := range c.informers {
		klog.Infof("Starting informer for %s", gvk.String())
		go informer.Run(stopCh)
	}

	// Wait for all informers to sync
	for gvk, informer := range c.informers {
		klog.Infof("Waiting for informer for %s to sync", gvk.String())
		if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
			return fmt.Errorf("failed to wait for caches to sync")
		}
	}

	klog.Info("Status controller synced and ready")

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	klog.Info("Shutting down status controller")
	return nil
}

func (c *StatusController) runWorker() {
	for {
		if !c.processNextItem() {
			return
		}
	}
}

func (c *StatusController) processNextItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	defer c.workqueue.Done(obj)

	key, ok := obj.(string)
	if !ok {
		klog.Errorf("Expected string in workqueue but got %#v", obj)
		c.workqueue.Forget(obj)
		return true
	}

	// Process the resource status change
	if err := c.processStatusChange(key); err != nil {
		klog.Errorf("Error processing status change for key %s: %v", key, err)
		c.workqueue.AddRateLimited(key)
		return true
	}

	c.workqueue.Forget(obj)
	return true
}

func (c *StatusController) handleObject(obj interface{}) {
	// Ensure we have a valid object
	var metaObj metav1.Object
	var ok bool

	if metaObj, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("Error decoding object, invalid type")
			return
		}
		metaObj, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			klog.Errorf("Error decoding object tombstone, invalid type")
			return
		}
	}

	// Get the key to put in the queue
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("Failed to get key from object: %v", err)
		return
	}

	// Get the current object from the unstructured data
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		klog.Errorf("Expected unstructured object, got %T", obj)
		return
	}

	// Extract conditions from status
	conditions, found, err := unstructured.NestedSlice(unstructuredObj.Object, "status", "conditions")
	if err != nil || !found {
		// No conditions found, but not an error - just means this object might not have conditions
		// We'll still process it to handle resources that have just started reporting conditions
		klog.V(5).Infof("No conditions found for %s/%s", metaObj.GetNamespace(), metaObj.GetName())
		c.workqueue.Add(key)
		return
	}

	// Parse conditions and check if they've changed
	changed := false
	currentStatus := make(map[string]string)

	// Try to extract conditions - different resources may store conditions differently
	for _, cond := range conditions {
		condition, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		condType, typeFound := condition["type"].(string)
		condStatus, statusFound := condition["status"].(string)

		if typeFound && statusFound {
			currentStatus[condType] = condStatus
		}
	}

	// Only process if we have actual conditions
	if len(currentStatus) == 0 {
		klog.V(5).Infof("No valid conditions extracted for %s/%s", metaObj.GetNamespace(), metaObj.GetName())
		return
	}

	// Check if status has changed
	previousStatus, exists := c.resourceStatus[key]
	if !exists || !statusEqual(previousStatus, currentStatus) {
		changed = true
		c.resourceStatus[key] = currentStatus
		klog.V(4).Infof("Status changed for %s/%s: %v", metaObj.GetNamespace(), metaObj.GetName(), currentStatus)
	}

	if changed {
		// Add to workqueue for processing
		c.workqueue.Add(key)
	}
}

// statusEqual checks if two status maps are equal
func statusEqual(old, new map[string]string) bool {
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

// processStatusChange processes a status change and triggers jobs if templates match
func (c *StatusController) processStatusChange(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("invalid resource key: %s", key)
	}

	// Get the resource from the informer
	var obj runtime.Object
	var resourceKind string

	// Find the right informer based on the resource kind
	// This is simplified - you'd need to identify the correct GVK
	for gvk, informer := range c.informers {
		if item, exists, err := informer.GetStore().GetByKey(key); err == nil && exists {
			var ok bool
			if obj, ok = item.(runtime.Object); ok {
				resourceKind = gvk.Kind
				break
			}
		}
	}

	if obj == nil {
		// Object may have been deleted, clean up our status tracking
		delete(c.resourceStatus, key)
		return nil
	}

	// Convert to unstructured to access fields
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return fmt.Errorf("failed to convert object to unstructured: %w", err)
	}

	// For testing, skip the template retrieval
	if os.Getenv("TEST_MODE") == "true" {
		klog.V(4).Infof("Running in test mode, skipping template retrieval")
		return nil
	}

	// Here we would normally get labels for label selector matching
	// This will be implemented later when we fully support label selectors
	/*
		objMeta, err := meta.Accessor(obj)
		if err != nil {
			return fmt.Errorf("failed to get object metadata: %w", err)
		}
		labels := objMeta.GetLabels() // Will use this with selectors
	*/

	// Get status conditions
	conditions, found, err := unstructured.NestedSlice(unstructuredObj, "status", "conditions")
	if err != nil {
		return fmt.Errorf("failed to get conditions: %w", err)
	}

	if !found || len(conditions) == 0 {
		// No conditions found, nothing to process
		return nil
	}

	// Process conditions
	conditionMap := make(map[string]string)
	for _, cond := range conditions {
		condition, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		condType, typeFound := condition["type"].(string)
		condStatus, statusFound := condition["status"].(string)

		if typeFound && statusFound {
			conditionMap[condType] = condStatus
		}
	}

	// Check each template for a match
	for _, template := range c.templates {
		// Skip templates without a StatusSelector
		if template.Spec.StatusSelector == nil {
			continue
		}

		// Check if resource kind matches
		if template.Spec.StatusSelector.ResourceKind != resourceKind {
			continue
		}

		// Check name pattern if specified
		if template.Spec.StatusSelector.NamePattern != "" {
			if !matchNamePattern(template.Spec.StatusSelector.NamePattern, name) {
				continue
			}
		}

		// Check namespace pattern if specified
		if template.Spec.StatusSelector.NamespacePattern != "" {
			if !matchNamePattern(template.Spec.StatusSelector.NamespacePattern, namespace) {
				continue
			}
		}

		// TODO: Implement label selector matching if needed
		if template.Spec.StatusSelector.LabelSelector != nil {
			// Skip for now - in a real implementation you'd check label selectors
			// This requires proper label selector matching
			klog.V(4).Infof("Label selector matching not implemented yet")
		}

		// Check if conditions match
		conditionsMatch := true
		for _, requiredCond := range template.Spec.StatusSelector.Conditions {
			actualStatus, exists := conditionMap[requiredCond.Type]
			if !exists || actualStatus != requiredCond.Status {
				conditionsMatch = false
				break
			}
		}

		if !conditionsMatch {
			continue
		}

		// Template matched, create a job
		klog.Infof("Template %s matched status conditions for %s/%s, creating job",
			template.Name, resourceKind, name)

		// Create job based on the template
		if err := c.createJobFromTemplate(&template, resourceKind, namespace, name, conditionMap); err != nil {
			klog.Errorf("Failed to create job from template %s: %v", template.Name, err)
			continue
		}
	}

	return nil
}

// createJobFromTemplate creates a job based on a template when status conditions match
func (c *StatusController) createJobFromTemplate(
	template *v1alpha1.EventTriggeredJob,
	resourceKind, namespace, name string,
	conditions map[string]string) error {

	// Create job name based on template name
	jobName := fmt.Sprintf("%s-%s-%s",
		template.Name,
		strings.ToLower(resourceKind),
		"status")

	// Create labels for the job
	labels := map[string]string{
		"kubanana-template":      template.Name,
		"kubanana-resource-kind": resourceKind,
		"kubanana-resource-name": name,
		"kubanana-trigger-type":  "status",
	}

	// Get resource condition types and statuses for labels
	for condType, condStatus := range conditions {
		safeCondType := strings.ReplaceAll(condType, ".", "-")    // Replace dots with dashes
		safeCondType = strings.ReplaceAll(safeCondType, " ", "-") // Replace spaces with dashes

		// Add condition to labels with a prefix to avoid conflicts
		labels[fmt.Sprintf("condition-%s", safeCondType)] = condStatus
	}

	// Create a job from the template
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: jobName + "-",
			// Use template namespace for the job, not resource namespace
			Namespace: template.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "kubanana.roshanbhatia.com/v1alpha1",
					Kind:       "EventTriggeredJob",
					Name:       template.Name,
					UID:        template.UID,
					Controller: &[]bool{true}[0],
				},
			},
		},
		Spec: template.Spec.JobTemplate.Spec,
	}

	// Apply variable substitution to the job spec
	for i, container := range job.Spec.Template.Spec.Containers {
		for j, cmd := range container.Command {
			job.Spec.Template.Spec.Containers[i].Command[j] = substituteStatusVariables(cmd, resourceKind, name, namespace, conditions)
		}

		// Add environment variables for the resource
		envVars := []corev1.EnvVar{
			{Name: "RESOURCE_KIND", Value: resourceKind},
			{Name: "RESOURCE_NAME", Value: name},
			{Name: "RESOURCE_NAMESPACE", Value: namespace},
			{Name: "TRIGGER_TYPE", Value: "status"},
		}

		// Add condition environment variables
		for condType, condStatus := range conditions {
			envVarName := fmt.Sprintf("STATUS_%s", strings.ReplaceAll(condType, "-", "_"))
			envVars = append(envVars, corev1.EnvVar{Name: envVarName, Value: condStatus})
		}

		// Add environment variables to the container
		for _, env := range envVars {
			// Check if the env var already exists
			exists := false
			for _, existingEnv := range container.Env {
				if existingEnv.Name == env.Name {
					exists = true
					break
				}
			}

			if !exists {
				job.Spec.Template.Spec.Containers[i].Env = append(
					job.Spec.Template.Spec.Containers[i].Env, env)
			}
		}
	}

	// Create the job in the template's namespace, not the resource's namespace
	createdJob, err := c.kubeClient.BatchV1().Jobs(template.Namespace).Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	klog.Infof("Created job %s/%s for status match", createdJob.Namespace, createdJob.Name)
	return nil
}

// substituteStatusVariables substitutes variables in a string for status-triggered jobs
func substituteStatusVariables(input, resourceKind, name, namespace string, conditions map[string]string) string {
	// Replace $RESOURCE_KIND with resourceKind
	input = strings.ReplaceAll(input, "$RESOURCE_KIND", resourceKind)

	// Replace $RESOURCE_NAME with name
	input = strings.ReplaceAll(input, "$RESOURCE_NAME", name)

	// Replace $RESOURCE_NAMESPACE with namespace
	input = strings.ReplaceAll(input, "$RESOURCE_NAMESPACE", namespace)

	// Replace $STATUS_X with condition values
	for condType, condStatus := range conditions {
		varName := fmt.Sprintf("$STATUS_%s", strings.ReplaceAll(condType, "-", "_"))
		input = strings.ReplaceAll(input, varName, condStatus)
	}

	return input
}

// kindToResource converts Kind to resource name (pluralized lowercase)
// This is a simplified version - a real implementation would use schema discovery
func kindToResource(kind string) string {
	resource := strings.ToLower(kind)
	// Simple pluralization
	if strings.HasSuffix(resource, "y") {
		return resource[:len(resource)-1] + "ies"
	}
	if strings.HasSuffix(resource, "s") {
		return resource
	}
	return resource + "s"
}
