// KubePulse - Kubernetes Cluster Monitor (TUI)
//
// Author: Erdem Unal
// Year: 2024
// Version: 0.1.0
// License: MIT

package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rdmnl/kubepulse/pkg/kubernetes"
	"github.com/rdmnl/kubepulse/utils"
	"github.com/rivo/tview"
)

const (
	quitInstruction            = "'q' Quit"
	podShortcut                = "'p' Pods"
	nodeShortcut               = "'n' Nodes"
	logShortcut                = "'l' Logs"
	detailShortcut             = "'d' Details"
	filterNamespaceInstruction = "'f' Filter Namespace"
	backInstruction            = "'b' Back"
)

type UIController struct {
	UIManager        *UIManager
	Application      *tview.Application
	KubernetesClient kubernetes.KubernetesClient
}

func NewUIController(app *tview.Application, uiManager *UIManager, client kubernetes.KubernetesClient) *UIController {
	controller := &UIController{
		Application:      app,
		UIManager:        uiManager,
		KubernetesClient: client,
	}

	// TODO temporarily disable for lagging problem
	// go controller.startPeriodicUpdate()

	return controller
}

func (controller *UIController) startPeriodicUpdate() {
	ticker := time.NewTicker(10 * time.Second) // Set update interval, e.g., every 10 seconds
	defer ticker.Stop()

	for {
		<-ticker.C
		controller.Application.QueueUpdateDraw(func() {
			controller.updatePodList()
			controller.updateNodeList()
		})
	}
}

func (controller *UIController) setPanelFocus(panelIndex int) {
	if panelIndex < 0 || panelIndex > 3 {
		errorMessage := fmt.Sprintf("Invalid panel index: %d", panelIndex)
		utils.Warn(errorMessage)
		controller.UIManager.StatusBar.SetText("[red]" + errorMessage)
		return
	}

	controller.UIManager.CurrentPanel = panelIndex
	panels := []tview.Primitive{
		controller.UIManager.PodListPanel,  // Index 0
		controller.UIManager.NodeListPanel, // Index 1
		controller.UIManager.DetailsPanel,  // Index 2
		controller.UIManager.LogsViewPanel, // Index 3
	}
	controller.Application.SetFocus(panels[panelIndex])

	controller.updateStatusBar()
	controller.updateFocusIndicator()
	utils.Info(fmt.Sprintf("Switched focus to panel: %d", panelIndex))
}

func (controller *UIController) HandlePodSelection() {
	row, _ := controller.UIManager.PodListPanel.GetSelection()
	if row < 1 || row >= controller.UIManager.PodListPanel.GetRowCount() {
		utils.Warn(fmt.Sprintf("Selected row index %d is out of bounds", row))
		return
	}
	selectedPod := controller.UIManager.PodListPanel.GetCell(row, 0).Text
	selectedNamespace := controller.UIManager.PodListPanel.GetCell(row, 1).Text
	if selectedPod == "" {
		utils.Warn("Selected pod name is empty")
		return
	}

	pod := kubernetes.Pod{
		Name:      selectedPod,
		Namespace: selectedNamespace,
	}

	podDetails, err := controller.KubernetesClient.GetPodDetails(pod)
	if err != nil {
		utils.Errorf("Error fetching pod details for %s/%s: %v", pod.Namespace, pod.Name, err)
		return
	}

	controller.UIManager.SelectedPod = fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
	controller.UIManager.DetailsPanel.Clear()
	controller.UIManager.DetailsPanel.SetText(podDetails)

	controller.Application.SetFocus(controller.UIManager.DetailsPanel)
	controller.updateStatusBar()
	controller.updateFocusIndicator()
	utils.Info(fmt.Sprintf("Updated details panel for pod: %s/%s", pod.Namespace, pod.Name))
}

func (controller *UIController) HandleLogView() {
	row, _ := controller.UIManager.PodListPanel.GetSelection()
	if row < 1 || row >= controller.UIManager.PodListPanel.GetRowCount() {
		errorMessage := fmt.Sprintf("Selected row index %d is out of bounds", row)
		utils.Warn(errorMessage)
		controller.UIManager.StatusBar.SetText("[red]" + errorMessage)
		return
	}
	selectedPod := controller.UIManager.PodListPanel.GetCell(row, 0).Text
	selectedNamespace := controller.UIManager.PodListPanel.GetCell(row, 1).Text
	if selectedPod == "" {
		errorMessage := "Selected pod name is empty"
		utils.Warn(errorMessage)
		controller.UIManager.StatusBar.SetText("[red]" + errorMessage)
		return
	}

	pod := kubernetes.Pod{
		Name:      selectedPod,
		Namespace: selectedNamespace,
	}

	podLogs, err := controller.KubernetesClient.GetPodLogs(pod)
	if err != nil {
		errorMessage := fmt.Sprintf("Error fetching pod logs for %s/%s: %v", pod.Namespace, pod.Name, err)
		utils.Errorf(errorMessage)
		controller.UIManager.StatusBar.SetText("[red]" + errorMessage)
		return
	}

	controller.UIManager.SelectedPod = fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
	controller.UIManager.LogsViewPanel.SetText(fmt.Sprintf("Logs for pod %s/%s:\n%s", pod.Namespace, pod.Name, podLogs))

	controller.Application.SetFocus(controller.UIManager.LogsViewPanel)
	controller.updateStatusBar()
	controller.updateFocusIndicator()
	utils.Info(fmt.Sprintf("Logs view panel for pod %s/%s displayed", pod.Namespace, pod.Name))
}

func (controller *UIController) HandleBackNavigation() {
	if controller.UIManager.LogsViewPanel.HasFocus() {
		controller.setPanelFocus(1) // Switch to DetailsPanel
	} else if controller.UIManager.DetailsPanel.HasFocus() {
		if controller.UIManager.SelectedNode != "" {
			controller.setPanelFocus(1) // Switch to NodeListPanel
		} else {
			controller.setPanelFocus(0) // Switch to PodListPanel
		}
	} else {
		errorMessage := "Unexpected focus state during back navigation"
		utils.Warn(errorMessage)
		controller.UIManager.StatusBar.SetText("[red]" + errorMessage)
		return
	}
}

func (controller *UIController) HandleNodeSelection() {
	row, _ := controller.UIManager.NodeListPanel.GetSelection()
	if row < 1 || row >= controller.UIManager.NodeListPanel.GetRowCount() {
		utils.Warn(fmt.Sprintf("Selected row index %d is out of bounds", row))
		return
	}
	selectedNode := controller.UIManager.NodeListPanel.GetCell(row, 0).Text
	if selectedNode == "" {
		utils.Warn("Selected node name is empty")
		return
	}

	pods, err := controller.KubernetesClient.GetPodsByNode(selectedNode)
	if err != nil {
		utils.Errorf("Error fetching pods for node %s: %v", selectedNode, err)
		return
	}

	controller.UIManager.SelectedNode = selectedNode
	controller.UIManager.PodListPanel.Clear()

	controller.UIManager.PodListPanel.SetCell(0, 0, tview.NewTableCell("Pod Name").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	controller.UIManager.PodListPanel.SetCell(0, 1, tview.NewTableCell("Namespace").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	controller.UIManager.PodListPanel.SetCell(0, 2, tview.NewTableCell("CPU").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	controller.UIManager.PodListPanel.SetCell(0, 3, tview.NewTableCell("Memory").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))

	for row, pod := range pods {
		cpuUsage, memoryUsage, err := controller.KubernetesClient.GetPodMetrics(pod)
		if err != nil {
			cpuUsage, memoryUsage = "N/A", "N/A"
			utils.Warn(fmt.Sprintf("Error fetching metrics for pod %s/%s: %v", pod.Namespace, pod.Name, err))
		}

		controller.UIManager.PodListPanel.SetCell(row+1, 0, tview.NewTableCell(pod.Name).
			SetTextColor(tcell.ColorLightYellow).
			SetSelectable(true).
			SetAlign(tview.AlignLeft))

		controller.UIManager.PodListPanel.SetCell(row+1, 1, tview.NewTableCell(pod.Namespace).
			SetTextColor(tcell.ColorLightGreen).
			SetSelectable(false).
			SetAlign(tview.AlignLeft))

		controller.UIManager.PodListPanel.SetCell(row+1, 2, tview.NewTableCell(cpuUsage).
			SetTextColor(tcell.ColorLightGreen).
			SetSelectable(false).
			SetAlign(tview.AlignRight))

		controller.UIManager.PodListPanel.SetCell(row+1, 3, tview.NewTableCell(memoryUsage).
			SetTextColor(tcell.ColorLightBlue).
			SetSelectable(false).
			SetAlign(tview.AlignRight))
	}

	controller.Application.SetFocus(controller.UIManager.PodListPanel)
	controller.updateStatusBar()
	controller.updateFocusIndicator()
	utils.Info(fmt.Sprintf("Displayed pods for node: %s", selectedNode))
}

func (controller *UIController) HandleNamespaceFilter() {
	form := tview.NewForm()

	namespaces, err := controller.KubernetesClient.ListNamespaces()
	if err != nil {
		utils.Warn(fmt.Sprintf("Error fetching namespaces: %v", err))
		controller.UIManager.StatusBar.SetText("[red]Error fetching namespaces")
		return
	}

	namespaceDropdown := tview.NewDropDown().
		SetLabel("Namespace: ").
		SetOptions(namespaces, nil)

	form.AddFormItem(namespaceDropdown).
		AddButton("Apply", func() {
			_, namespace := namespaceDropdown.GetCurrentOption()
			if namespace == "" {
				namespace = "default"
			}
			controller.KubernetesClient.SetNamespace(namespace)
			controller.updatePodList()
			controller.updatePodTable()
			controller.Application.SetRoot(controller.UIManager.Layout, true)
		}).
		AddButton("Cancel", func() {
			controller.Application.SetRoot(controller.UIManager.Layout, true)
		})

	modal := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(form, 10, 1, true).
		AddItem(nil, 0, 1, false)

	form.SetBorder(true).
		SetTitle("Select Namespace").
		SetTitleAlign(tview.AlignCenter)

	controller.Application.SetRoot(modal, true)

}

func (controller *UIController) updatePodTable() {
	pods, err := controller.KubernetesClient.GetPods()
	if err != nil {
		utils.Warn(fmt.Sprintf("Error fetching pods: %v", err))
		controller.UIManager.StatusBar.SetText("[red]Error fetching pods")
		return
	}

	controller.UIManager.PodListPanel.Clear()

	controller.UIManager.PodListPanel.SetCell(0, 0, tview.NewTableCell("Pod Name").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	controller.UIManager.PodListPanel.SetCell(0, 1, tview.NewTableCell("Namespace").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	controller.UIManager.PodListPanel.SetCell(0, 2, tview.NewTableCell("CPU").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	controller.UIManager.PodListPanel.SetCell(0, 3, tview.NewTableCell("Memory").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))

	for row, pod := range pods {
		cpuUsage, memoryUsage, err := controller.KubernetesClient.GetPodMetrics(pod)
		if err != nil {
			cpuUsage = "N/A"
			memoryUsage = "N/A"
			utils.Warn(fmt.Sprintf("Error fetching metrics for pod %s/%s: %v", pod.Namespace, pod.Name, err))
		}

		controller.UIManager.PodListPanel.SetCell(row+1, 0, tview.NewTableCell(pod.Name).
			SetTextColor(tcell.ColorLightYellow).
			SetSelectable(true).
			SetAlign(tview.AlignLeft))

		controller.UIManager.PodListPanel.SetCell(row+1, 1, tview.NewTableCell(pod.Namespace).
			SetTextColor(tcell.ColorLightGreen).
			SetSelectable(false).
			SetAlign(tview.AlignLeft))

		controller.UIManager.PodListPanel.SetCell(row+1, 2, tview.NewTableCell(cpuUsage).
			SetTextColor(tcell.ColorLightGreen).
			SetSelectable(false).
			SetAlign(tview.AlignRight))

		controller.UIManager.PodListPanel.SetCell(row+1, 3, tview.NewTableCell(memoryUsage).
			SetTextColor(tcell.ColorLightBlue).
			SetSelectable(false).
			SetAlign(tview.AlignRight))
	}

	controller.updateStatusBar()
}

func (controller *UIController) updateNodeList() {
	nodes, err := controller.KubernetesClient.GetNodes()
	if err != nil {
		controller.UIManager.StatusBar.SetText("[red]Error fetching nodes")
		return
	}

	controller.UIManager.NodeListPanel.Clear()

	controller.UIManager.NodeListPanel.SetCell(0, 0, tview.NewTableCell("Node Name").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	controller.UIManager.NodeListPanel.SetCell(0, 1, tview.NewTableCell("CPU").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	controller.UIManager.NodeListPanel.SetCell(0, 2, tview.NewTableCell("Memory").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))

	for row, node := range nodes {
		cpuUsage, memoryUsage, err := controller.KubernetesClient.GetNodeMetrics(node)
		if err != nil {
			cpuUsage = "N/A"
			memoryUsage = "N/A"
		}

		controller.UIManager.NodeListPanel.SetCell(row+1, 0, tview.NewTableCell(node).
			SetTextColor(tcell.ColorLightYellow).
			SetSelectable(true))
		controller.UIManager.NodeListPanel.SetCell(row+1, 1, tview.NewTableCell(cpuUsage).
			SetTextColor(tcell.ColorLightGreen).
			SetAlign(tview.AlignRight))
		controller.UIManager.NodeListPanel.SetCell(row+1, 2, tview.NewTableCell(memoryUsage).
			SetTextColor(tcell.ColorLightBlue).
			SetAlign(tview.AlignRight))
	}
}

func (controller *UIController) updatePodList() {
	pods, err := controller.KubernetesClient.GetPods()
	if err != nil {
		utils.Warn(fmt.Sprintf("Error fetching pods: %v", err))
		controller.UIManager.StatusBar.SetText("[red]Error fetching pods")
		return
	}

	controller.UIManager.PodListPanel.Clear()

	controller.UIManager.PodListPanel.SetCell(0, 0, tview.NewTableCell("Pod Name").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	controller.UIManager.PodListPanel.SetCell(0, 1, tview.NewTableCell("Namespace").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	controller.UIManager.PodListPanel.SetCell(0, 2, tview.NewTableCell("CPU").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	controller.UIManager.PodListPanel.SetCell(0, 3, tview.NewTableCell("Memory").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))

	for row, pod := range pods {
		if pod.Name == "" {
			continue
		}

		cpuUsage, memoryUsage, err := controller.KubernetesClient.GetPodMetrics(pod)
		if err != nil {
			cpuUsage = "N/A"
			memoryUsage = "N/A"
			utils.Warn(fmt.Sprintf("Error fetching metrics for pod %s/%s: %v", pod.Namespace, pod.Name, err))
		}

		controller.UIManager.PodListPanel.SetCell(row+1, 0, tview.NewTableCell(pod.Name).
			SetTextColor(tcell.ColorLightYellow).
			SetSelectable(true).
			SetAlign(tview.AlignLeft))

		controller.UIManager.PodListPanel.SetCell(row+1, 1, tview.NewTableCell(pod.Namespace).
			SetTextColor(tcell.ColorLightGreen).
			SetSelectable(false).
			SetAlign(tview.AlignLeft))

		controller.UIManager.PodListPanel.SetCell(row+1, 2, tview.NewTableCell(cpuUsage).
			SetTextColor(tcell.ColorLightGreen).
			SetSelectable(false).
			SetAlign(tview.AlignRight))

		controller.UIManager.PodListPanel.SetCell(row+1, 3, tview.NewTableCell(memoryUsage).
			SetTextColor(tcell.ColorLightBlue).
			SetSelectable(false).
			SetAlign(tview.AlignRight))
	}

	controller.updateStatusBar()
}

func (controller *UIController) updateFocusIndicator() {
	if controller.UIManager.CurrentPanel < 0 || controller.UIManager.CurrentPanel > 3 {
		errorMessage := fmt.Sprintf("Invalid panel index: %d", controller.UIManager.CurrentPanel)
		utils.Warn(errorMessage)
		controller.UIManager.StatusBar.SetText("[red]" + errorMessage)
		return
	}

	panels := []*tview.Table{
		controller.UIManager.PodListPanel,
		controller.UIManager.NodeListPanel,
	}
	textPanels := []*tview.TextView{
		controller.UIManager.DetailsPanel,
		controller.UIManager.LogsViewPanel,
	}

	panelTitles := [4]string{" Pods ", " Nodes ", " Detail ", " Logs "}

	for i, panel := range panels {
		if i == controller.UIManager.CurrentPanel {
			panel.SetBorderColor(tcell.ColorLightGreen)
			panel.SetTitle(panelTitles[i])
		} else {
			panel.SetBorderColor(tcell.ColorGray)
			panel.SetTitle("")
		}
	}

	for i, panel := range textPanels {
		if i+2 == controller.UIManager.CurrentPanel {
			panel.SetBorderColor(tcell.ColorLightGreen)
			panel.SetTitle(panelTitles[i+2])
		} else {
			panel.SetBorderColor(tcell.ColorGray)
			panel.SetTitle("")
		}
	}

	utils.Info(fmt.Sprintf("Focus indicator updated for panel: %d", controller.UIManager.CurrentPanel))
}

func (controller *UIController) getSelectedPod() (kubernetes.Pod, error) {
	row, _ := controller.UIManager.PodListPanel.GetSelection()
	if row < 1 || row >= controller.UIManager.PodListPanel.GetRowCount() {
		return kubernetes.Pod{}, fmt.Errorf("selected row index %d is out of bounds", row)
	}
	selectedPod := controller.UIManager.PodListPanel.GetCell(row, 0).Text
	selectedNamespace := controller.UIManager.PodListPanel.GetCell(row, 1).Text
	if selectedPod == "" {
		return kubernetes.Pod{}, fmt.Errorf("selected pod name is empty")
	}
	return kubernetes.Pod{
		Name:      selectedPod,
		Namespace: selectedNamespace,
	}, nil
}

func (controller *UIController) updateStatusBar() {
	statusMessage := controller.getStatusBarMessage(controller.UIManager.CurrentPanel, controller.UIManager.SelectedPod)
	controller.UIManager.StatusBar.SetText(statusMessage)
	utils.Info("Status bar updated")
}

func (controller *UIController) getStatusBarMessage(panel int, selectedPod string) string {
	switch panel {
	case 0: // PodListPanel
		if controller.UIManager.PodListPanel.GetRowCount() > 1 {
			return fmt.Sprintf("%s | %s | %s | %s | %s | %s | %s",
				quitInstruction,
				podShortcut,
				nodeShortcut,
				detailShortcut,
				logShortcut,
				filterNamespaceInstruction,
				backInstruction)
		} else {
			return fmt.Sprintf("No pods available. %s | %s | %s",
				quitInstruction,
				podShortcut,
				nodeShortcut)
		}
	case 1: // NodeListPanel
		return fmt.Sprintf("%s | %s | %s | %s",
			quitInstruction,
			podShortcut,
			nodeShortcut,
			backInstruction)
	case 2: // DetailsPanel
		if selectedPod != "" {
			return fmt.Sprintf("%s | %s | %s | %s",
				backInstruction,
				podShortcut,
				nodeShortcut,
				quitInstruction)
		} else {
			return fmt.Sprintf("No pod selected. %s | %s | %s",
				backInstruction,
				podShortcut,
				nodeShortcut)
		}
	case 3: // LogsViewPanel
		if selectedPod != "" {
			return fmt.Sprintf("%s | %s | %s | %s",
				backInstruction,
				podShortcut,
				nodeShortcut,
				quitInstruction)
		} else {
			return fmt.Sprintf("No logs available. %s | %s | %s",
				backInstruction,
				podShortcut,
				nodeShortcut)
		}
	default:
		return fmt.Sprintf("[red]Invalid panel index: %d", panel)
	}
}
