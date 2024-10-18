// KubePulse - Kubernetes Cluster Monitor (TUI)
//
// Author: Erdem Unal
// Year: 2024
// Version: 0.1.0
// License: MIT
//
// Description:
// This package (`kubepulse`) provides a terminal-based interface (TUI) for monitoring Kubernetes clusters,
// helping developers and system administrators gain real-time insights into pod statuses, resource usage (CPU and memory),
// and other key metrics. With a clear and interactive user interface, `kubepulse` simplifies navigating and managing
// Kubernetes resources, offering detailed pod views, real-time log streams, and configurable options for different namespaces.
// It is designed to streamline cluster monitoring and troubleshooting from within the terminal.

package main

import (
	"log"
	"os"

	"github.com/rdmnl/kubepulse/pkg/kubernetes"
	"github.com/rdmnl/kubepulse/ui"
	"github.com/rivo/tview"
)

func main() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	log.SetOutput(file)

	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = os.ExpandEnv("$HOME/.kube/config")
	}

	namespace := "default"
	client, err := kubernetes.NewClient(kubeconfigPath, namespace)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v. Ensure your KUBECONFIG environment variable is correctly set or provide a valid kubeconfig path.", err)
	}

	app := tview.NewApplication()

	uiManager, layout := ui.SetupUILayout(app, client)

	controller := ui.NewUIController(app, uiManager, client)
	ui.SetupNavigation(app, controller)

	if err := app.SetRoot(layout, true).Run(); err != nil {
		panic(err)
	}
}
