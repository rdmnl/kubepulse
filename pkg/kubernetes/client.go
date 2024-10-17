// KubePulse - Kubernetes Cluster Monitor (TUI)
//
// Author: Erdem Unal
// Year: 2024
// Version: 0.1.0
// License: MIT

package kubernetes

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)


type KubernetesClient interface {
    GetPods() ([]string, error)
    GetPodDetails(podName string) (string, error)
    GetPodLogs(podName string) (string, error)
    SetNamespace(namespace string)
    GetPodMetrics(podName string) (cpuUsage string, memoryUsage string, err error)
    ListNamespaces() ([]string, error)
}



type Client struct {
    clientset     *kubernetes.Clientset
    metricsClient *metricsclient.Clientset
    namespace     string // Namespace to interact with
}

// NewClient initializes a new Kubernetes client
func NewClient(kubeconfigPath string, namespace string) (*Client, error) {
    var config *rest.Config
    var err error

    if kubeconfigPath != "" {
        config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
        if err != nil {
            return nil, fmt.Errorf("failed to load kubeconfig from %s: %v", kubeconfigPath, err)
        }
    } else {
        config, err = rest.InClusterConfig()
        if err != nil {
            return nil, fmt.Errorf("unable to load in-cluster configuration: %v. Please set the KUBECONFIG environment variable to your kubeconfig path", err)
        }
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
    }

    metricsClient, err := metricsclient.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create Kubernetes metrics client: %v", err)
    }

    return &Client{clientset: clientset, metricsClient: metricsClient, namespace: namespace}, nil
}


// SetNamespace allows changing the namespace for the client
func (c *Client) SetNamespace(namespace string) {
    c.namespace = namespace
}

// GetPods retrieves the list of pods from the cluster in the specified namespace
func (c *Client) GetPods() ([]string, error) {
    pods, err := c.clientset.CoreV1().Pods(c.namespace).List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    var podNames []string
    for _, pod := range pods.Items {
        podNames = append(podNames, pod.Name)
    }
    return podNames, nil
}

// GetPodDetails retrieves the details of a specific pod by name
func (c *Client) GetPodDetails(podName string) (string, error) {
    // Fetch the pod from the Kubernetes API
    pod, err := c.clientset.CoreV1().Pods(c.namespace).Get(context.TODO(), podName, metav1.GetOptions{})
    if err != nil {
        return "", err
    }

    // Prepare pod details with improved formatting and color
    details := fmt.Sprintf("[yellow::b]Pod Info[-::-]\n")
    details += fmt.Sprintf("[lightcyan]Pod Name:[-] %s\n", pod.Name)
    details += fmt.Sprintf("[lightcyan]Namespace:[-] %s\n", pod.Namespace)
    details += fmt.Sprintf("[lightcyan]Status:[-] %s\n", pod.Status.Phase)
    details += fmt.Sprintf("[lightcyan]Node:[-] %s\n\n", pod.Spec.NodeName)

    // Add container-specific information, such as resource requests and limits
    details += "[yellow::b]Containers[-::-]\n"
    for _, container := range pod.Spec.Containers {
        details += fmt.Sprintf("  [green]Container:[-] %s\n", container.Name)

        // Resource Requests
        details += "    [lightgreen::b]Requests[-::-]\n"
        if req, ok := container.Resources.Requests[v1.ResourceCPU]; ok {
            details += fmt.Sprintf("      [lightcyan]CPU:[-] %s\n", req.String())
        } else {
            details += "      [lightcyan]CPU:[-] Not Set\n"
        }

        if req, ok := container.Resources.Requests[v1.ResourceMemory]; ok {
            details += fmt.Sprintf("      [lightcyan]Memory:[-] %s\n", req.String())
        } else {
            details += "      [lightcyan]Memory:[-] Not Set\n"
        }

        // Resource Limits
        details += "    [lightgreen::b]Limits[-::-]\n"
        if limit, ok := container.Resources.Limits[v1.ResourceCPU]; ok {
            details += fmt.Sprintf("      [lightcyan]CPU:[-] %s\n", limit.String())
        } else {
            details += "      [lightcyan]CPU:[-] Not Set\n"
        }

        if limit, ok := container.Resources.Limits[v1.ResourceMemory]; ok {
            details += fmt.Sprintf("      [lightcyan]Memory:[-] %s\n", limit.String())
        } else {
            details += "      [lightcyan]Memory:[-] Not Set\n"
        }

        details += "\n"
    }

    return details, nil
}



// GetPodLogs retrieves the logs for a specific pod
func (c *Client) GetPodLogs(podName string) (string, error) {
    logOptions := &v1.PodLogOptions{}
    req := c.clientset.CoreV1().Pods(c.namespace).GetLogs(podName, logOptions)
    logs, err := req.Stream(context.TODO())
    if err != nil {
        return "", err
    }
    defer logs.Close()

    var result string
    buf := make([]byte, 2000)
    for {
        numBytes, err := logs.Read(buf)
        if numBytes == 0 {
            break
        }
        result += string(buf[:numBytes])
        if err != nil {
            if err.Error() == "EOF" {
                break
            }
            return "", err
        }
    }
    return result, nil
}

func (c *Client) GetPodMetrics(podName string) (cpuUsage string, memoryUsage string, err error) {
    podMetrics, err := c.metricsClient.MetricsV1beta1().PodMetricses(c.namespace).Get(context.TODO(), podName, metav1.GetOptions{})
    if err != nil {
        return "", "", err
    }

    var totalCPU, totalMemory int64
    for _, container := range podMetrics.Containers {
        cpuQuantity := container.Usage[v1.ResourceCPU]
        memoryQuantity := container.Usage[v1.ResourceMemory]
        totalCPU += cpuQuantity.MilliValue()
        totalMemory += memoryQuantity.Value()
    }

    return fmt.Sprintf("%dm", totalCPU), fmt.Sprintf("%dMi", totalMemory/(1024*1024)), nil
}

// ListNamespaces retrieves all namespaces available in the cluster
func (c *Client) ListNamespaces() ([]string, error) {
    namespaces, err := c.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return nil, err
    }

    var namespaceNames []string
    for _, ns := range namespaces.Items {
        namespaceNames = append(namespaceNames, ns.Name)
    }

    return namespaceNames, nil
}
