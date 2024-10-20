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
	GetNodes() ([]string, error)
	GetNodeMetrics(nodeName string) (cpuUsage string, memoryUsage string, err error)
	GetPods() ([]Pod, error)
	GetPodsByNode(nodeName string) ([]Pod, error)
	GetPodMetrics(pod Pod) (cpuUsage string, memoryUsage string, err error)
	GetPodDetails(pod Pod) (string, error)
	GetPodLogs(pod Pod) (string, error)
	SetNamespace(namespace string)
	ListNamespaces() ([]string, error)
}

type Pod struct {
	Name      string
	Namespace string
	NodeName  string
}

type Client struct {
	clientset     *kubernetes.Clientset
	metricsClient *metricsclient.Clientset
	namespace     string
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

func (c *Client) SetNamespace(namespace string) {
	c.namespace = namespace
}

func (c *Client) GetNodes() ([]string, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var nodeNames []string
	for _, node := range nodes.Items {
		nodeNames = append(nodeNames, node.Name)
	}

	return nodeNames, nil
}

func (c *Client) GetNodeMetrics(nodeName string) (cpuUsage string, memoryUsage string, err error) {
	nodeMetrics, err := c.metricsClient.MetricsV1beta1().NodeMetricses().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}

	cpuQuantity := nodeMetrics.Usage[v1.ResourceCPU]
	memoryQuantity := nodeMetrics.Usage[v1.ResourceMemory]

	totalCPU := cpuQuantity.MilliValue()
	totalMemory := memoryQuantity.Value()

	return fmt.Sprintf("%dm", totalCPU), fmt.Sprintf("%dMi", totalMemory/(1024*1024)), nil
}

func (c *Client) GetPods() ([]Pod, error) {
	pods, err := c.clientset.CoreV1().Pods(c.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var podList []Pod
	for _, pod := range pods.Items {
		podList = append(podList, Pod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			NodeName:  pod.Spec.NodeName,
		})
	}
	return podList, nil
}

func (c *Client) GetPodsByNode(nodeName string) ([]Pod, error) {
	pods, err := c.clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	})
	if err != nil {
		return nil, err
	}
	var podList []Pod
	for _, pod := range pods.Items {
		podList = append(podList, Pod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			NodeName:  pod.Spec.NodeName,
		})
	}
	return podList, nil
}

func (c *Client) GetPodMetrics(pod Pod) (cpuUsage string, memoryUsage string, err error) {
	podMetrics, err := c.metricsClient.MetricsV1beta1().PodMetricses(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
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

func (c *Client) GetPodDetails(pod Pod) (string, error) {
	podObj, err := c.clientset.CoreV1().Pods(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	details := fmt.Sprintf("[yellow::b]Pod Info[-::-]\n")
	details += fmt.Sprintf("[lightcyan]Pod Name:[-] %s\n", podObj.Name)
	details += fmt.Sprintf("[lightcyan]Namespace:[-] %s\n", podObj.Namespace)
	details += fmt.Sprintf("[lightcyan]Status:[-] %s\n", podObj.Status.Phase)
	details += fmt.Sprintf("[lightcyan]Node:[-] %s\n\n", podObj.Spec.NodeName)

	details += "[yellow::b]Containers[-::-]\n"
	for _, container := range podObj.Spec.Containers {
		details += fmt.Sprintf("  [green]Container:[-] %s\n", container.Name)

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

func (c *Client) GetPodLogs(pod Pod) (string, error) {
	logOptions := &v1.PodLogOptions{}
	req := c.clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)
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
