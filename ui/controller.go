// KubePulse - Kubernetes Cluster Monitor (TUI)
//
// Author: Erdem Unal
// Year: 2024
// Version: 0.1.0
// License: MIT

package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rdmnl/kubepulse/pkg/kubernetes"
	"github.com/rdmnl/kubepulse/utils"
	"github.com/rivo/tview"
)

const (
    quitInstruction         = "'q' Quit"
    switchPanelsInstruction = "'Tab/Arrow' Switch Panels"
    backInstruction         = "'b' Back"
    selectPodInstruction    = "'Enter' Select Pod"
    viewLogsInstruction     = "'l' View Logs"
    filterNamespaceInstruction = "'f' Filter Namespace"
)




type UIController struct {
    UIManager        *UIManager
    Application      *tview.Application
    KubernetesClient kubernetes.KubernetesClient
}

// NewUIController initializes a new UIController with the necessary state
func NewUIController(app *tview.Application, uiManager *UIManager, client kubernetes.KubernetesClient) *UIController {
    return &UIController{
        Application:      app,
        UIManager:        uiManager,
        KubernetesClient: client,
    }
}




func (controller *UIController) setPanelFocus(panelIndex int) {
    if panelIndex < 0 || panelIndex > 2 {
        errorMessage := fmt.Sprintf("Invalid panel index: %d", panelIndex)
        utils.Warn(errorMessage)
        controller.UIManager.StatusBar.SetText("[red]" + errorMessage)
        return
    }

    controller.UIManager.CurrentPanel = panelIndex
    panels := []tview.Primitive{controller.UIManager.PodListPanel, controller.UIManager.DetailsPanel, controller.UIManager.LogsViewPanel}
    controller.Application.SetFocus(panels[panelIndex])

    // Update status bar and focus indicator
    controller.updateStatusBar()
    controller.updateFocusIndicator()
    utils.Info(fmt.Sprintf("Switched focus to panel: %d", panelIndex))
}



// HandlePodSelection handles selecting a pod from the PodListPanel
func (controller *UIController) HandlePodSelection() {
    row, _ := controller.UIManager.PodListPanel.GetSelection()
    if row < 0 || row >= controller.UIManager.PodListPanel.GetRowCount() {
        utils.Warn(fmt.Sprintf("Selected row index %d is out of bounds", row))
        return
    }
    selectedPod := controller.UIManager.PodListPanel.GetCell(row, 0).Text
    if selectedPod == "" {
        utils.Warn("Selected pod name is empty")
        return
    }

    // Update details panel with selected pod details
    podDetails, err := controller.KubernetesClient.GetPodDetails(selectedPod)
    if err != nil {
        utils.Errorf("Error fetching pod details for %s: %v", selectedPod, err)
        return
    }

    controller.UIManager.SelectedPod = selectedPod
    controller.UIManager.DetailsPanel.Clear()
    controller.UIManager.DetailsPanel.SetText(podDetails)

    // Update focus and status bar
    controller.Application.SetFocus(controller.UIManager.DetailsPanel)
    controller.updateStatusBar()
    controller.updateFocusIndicator()
    utils.Info(fmt.Sprintf("Updated details panel for pod: %s", selectedPod))
}





// HandleLogView handles viewing logs for the selected pod
func (controller *UIController) HandleLogView() {
    selectedPod, err := controller.getSelectedPod()
    if err != nil {
        errorMessage := fmt.Sprintf("Error selecting pod: %v", err)
        utils.Warn(errorMessage)
        controller.UIManager.StatusBar.SetText("[red]" + errorMessage)
        return
    }

    // Fetch pod logs using the KubernetesClient instance
    podLogs, err := controller.KubernetesClient.GetPodLogs(selectedPod)
    if err != nil {
        errorMessage := fmt.Sprintf("Error fetching pod logs for %s: %v", selectedPod, err)
        utils.Errorf(errorMessage)
        controller.UIManager.StatusBar.SetText("[red]" + errorMessage)
        return
    }

    controller.UIManager.SelectedPod = selectedPod
    controller.UIManager.LogsViewPanel.SetText(fmt.Sprintf("Logs for pod %s:\n%s", selectedPod, podLogs))

    // Update focus and status bar
    controller.Application.SetFocus(controller.UIManager.LogsViewPanel)
    controller.updateStatusBar()
    controller.updateFocusIndicator()
    utils.Info(fmt.Sprintf("Logs view panel for pod %s displayed", selectedPod))
}




// HandleBackNavigation handles the back navigation action
func (controller *UIController) HandleBackNavigation() {
    if controller.UIManager.LogsViewPanel.HasFocus() {
        controller.setPanelFocus(1) // Switch to DetailsPanel
    } else if controller.UIManager.DetailsPanel.HasFocus() {
        controller.setPanelFocus(0) // Switch to PodListPanel
    } else {
        errorMessage := "Unexpected focus state during back navigation"
        utils.Warn(errorMessage)
        controller.UIManager.StatusBar.SetText("[red]" + errorMessage)
        return
    }
}

// HandleNamespaceFilter handles filtering the pod list by namespace
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
            
            // Restore the full layout including header and status bar
            controller.Application.SetRoot(controller.UIManager.Layout, true)
        }).
        AddButton("Cancel", func() {
            // Restore the full layout including header and status bar
            controller.Application.SetRoot(controller.UIManager.Layout, true)
        })

    // Temporarily set the form as the root, preserving the full layout structure
    controller.Application.SetRoot(tview.NewFlex().
        SetDirection(tview.FlexRow).
        AddItem(controller.UIManager.Header, 3, 1, false).
        AddItem(form, 0, 1, true).
        AddItem(controller.UIManager.StatusBar, 1, 1, false), true)
}




// updatePodList updates the pod list panel with the current namespace's pods
func (controller *UIController) updatePodList() {
    pods, err := controller.KubernetesClient.GetPods()
    if err != nil {
        utils.Warn(fmt.Sprintf("Error fetching pods: %v", err))
        controller.UIManager.StatusBar.SetText("[red]Error fetching pods")
        return
    }

    controller.UIManager.PodListPanel.Clear()
    for row, pod := range pods {
        controller.UIManager.PodListPanel.SetCell(row, 0, tview.NewTableCell(pod).
            SetTextColor(tcell.ColorLightYellow).
            SetSelectable(true))
    }

    controller.updateStatusBar()
}


// UpdateFocusIndicator updates the visual indication of the currently focused panel
func (controller *UIController) updateFocusIndicator() {
    if controller.UIManager.CurrentPanel < 0 || controller.UIManager.CurrentPanel > 2 {
        errorMessage := fmt.Sprintf("Invalid panel index: %d", controller.UIManager.CurrentPanel)
        utils.Warn(errorMessage)
        controller.UIManager.StatusBar.SetText("[red]" + errorMessage)
        return
    }

    panels := []*tview.Table{
        controller.UIManager.PodListPanel,
    }
    textPanels := []*tview.TextView{
        controller.UIManager.DetailsPanel,
        controller.UIManager.LogsViewPanel,
    }

    for i, panel := range panels {
        if i == controller.UIManager.CurrentPanel {
            panel.SetBorderColor(tcell.ColorYellow) // Highlight focused panel
            panel.SetTitle(fmt.Sprintf("[yellow]%s[white]", panel.GetTitle())) // Highlight title in yellow
        } else {
            panel.SetBorderColor(tcell.ColorLightCyan) // Default color for unfocused
            panel.SetTitle(panel.GetTitle()) // Reset title
        }
    }

    for i, panel := range textPanels {
        if i+1 == controller.UIManager.CurrentPanel { // +1 because tables and text panels are treated differently here
            panel.SetBorderColor(tcell.ColorYellow) // Highlight focused panel
            panel.SetTitle(fmt.Sprintf("[yellow]%s[white]", panel.GetTitle())) // Highlight title in yellow
        } else {
            panel.SetBorderColor(tcell.ColorLightCyan) // Default color for unfocused
            panel.SetTitle(panel.GetTitle()) // Reset title
        }
    }

    utils.Info(fmt.Sprintf("Focus indicator updated for panel: %d", controller.UIManager.CurrentPanel))
}

func (controller *UIController) getSelectedPod() (string, error) {
    row, _ := controller.UIManager.PodListPanel.GetSelection()
    if row < 0 || row >= controller.UIManager.PodListPanel.GetRowCount() {
        return "", fmt.Errorf("selected row index %d is out of bounds", row)
    }
    selectedPod := controller.UIManager.PodListPanel.GetCell(row, 0).Text
    if selectedPod == "" {
        return "", fmt.Errorf("selected pod name is empty")
    }
    return selectedPod, nil
}

func (controller *UIController) getStatusBarMessage(panel int, selectedPod string) string {
    switch panel {
    case 0: // PodListPanel
        if controller.UIManager.PodListPanel.GetRowCount() > 0 {
            return fmt.Sprintf("%s | %s | %s | %s | %s", quitInstruction, filterNamespaceInstruction, viewLogsInstruction, switchPanelsInstruction, selectPodInstruction)
        } else {
            return fmt.Sprintf("No pods available. %s | %s", quitInstruction, switchPanelsInstruction)
        }
    case 1: // DetailsPanel
        if selectedPod != "" {
            return fmt.Sprintf("%s | %s | %s", backInstruction, switchPanelsInstruction, quitInstruction)
        } else {
            return fmt.Sprintf("No pod selected. %s", backInstruction)
        }
    case 2: // LogsViewPanel
        if selectedPod != "" {
            return fmt.Sprintf("%s | %s | %s", backInstruction, switchPanelsInstruction, quitInstruction)
        } else {
            return fmt.Sprintf("No logs available. %s", backInstruction)
        }
    default:
        return fmt.Sprintf("[red]Invalid panel index: %d", panel)
    }
}



func (controller *UIController) updateStatusBar() {
    statusMessage := controller.getStatusBarMessage(controller.UIManager.CurrentPanel, controller.UIManager.SelectedPod)
    controller.UIManager.StatusBar.SetText(statusMessage)
    utils.Info("Status bar updated")
}
