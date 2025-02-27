package main

import (
	"flag"

	"github.com/roshbhatia/kubevent/pkg/controller"
	"github.com/roshbhatia/kubevent/pkg/util"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	var kubeconfig string
	var masterURL string

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building dynamic client: %s", err.Error())
	}

	stopCh := util.SetupSignalHandler()

	// Create controllers
	eventController := controller.NewEventController(kubeClient)
	statusController := controller.NewStatusController(kubeClient, dynamicClient)

	// Run the event controller
	go func() {
		if err := eventController.Run(2, stopCh); err != nil {
			klog.Fatalf("Error running event controller: %s", err.Error())
		}
	}()

	// Run the status controller (blocking)
	if err := statusController.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running status controller: %s", err.Error())
	}
}
