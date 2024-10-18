// KubePulse - Kubernetes Cluster Monitor (TUI)
//
// Author: Erdem Unal
// Year: 2024
// Version: 0.1.0
// License: MIT

package panels

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rdmnl/kubepulse/pkg/kubernetes"
	"github.com/rdmnl/kubepulse/utils"
	"github.com/rivo/tview"
)

// SetupPodListPanel sets up the pod list panel using data from the Kubernetes package
func SetupPodListPanel(client kubernetes.KubernetesClient) *tview.Table {
    table := tview.NewTable()

    // Set up table properties without borders
    table.SetBorders(false). // Remove borders between cells
        SetSelectable(true, false).
        SetFixed(1, 0).
        SetBackgroundColor(tcell.ColorBlack).
        SetBorder(true). // Keep the outer border of the entire table view
        SetBorderColor(tcell.ColorLightCyan)

    // Set header row (adjust to align with the main border)
    table.SetCell(0, 0, tview.NewTableCell("Pod Name").
        SetTextColor(tcell.ColorWhite).
        SetSelectable(false).
        SetAlign(tview.AlignCenter))
    table.SetCell(0, 1, tview.NewTableCell("CPU").
        SetTextColor(tcell.ColorWhite).
        SetSelectable(false).
        SetAlign(tview.AlignCenter))
    table.SetCell(0, 2, tview.NewTableCell("Memory").
        SetTextColor(tcell.ColorWhite).
        SetSelectable(false).
        SetAlign(tview.AlignCenter))

    pods, err := client.GetPods()
    if err != nil {
        utils.Info(fmt.Sprintf("Error fetching pods: %v", err))
        return table
    }

    for row, pod := range pods {
        if pod == "" {
            continue
        }

        cpuUsage, memoryUsage, err := client.GetPodMetrics(pod)
        if err != nil {
            cpuUsage = "N/A"
            memoryUsage = "N/A"
            utils.Info(fmt.Sprintf("Error fetching metrics for pod %s: %v", pod, err))
        }

        table.SetCell(row+1, 0, tview.NewTableCell(pod).
            SetTextColor(tcell.ColorLightYellow).
            SetBackgroundColor(tcell.ColorBlack).
            SetSelectable(true).
            SetAlign(tview.AlignLeft)) // Align text to the left for consistency

        table.SetCell(row+1, 1, tview.NewTableCell(cpuUsage).
            SetTextColor(tcell.ColorLightGreen).
            SetBackgroundColor(tcell.ColorBlack).
            SetSelectable(false).
            SetAlign(tview.AlignRight)) // Align usage values to the right for neatness

        table.SetCell(row+1, 2, tview.NewTableCell(memoryUsage).
            SetTextColor(tcell.ColorLightBlue).
            SetBackgroundColor(tcell.ColorBlack).
            SetSelectable(false).
            SetAlign(tview.AlignRight)) // Align usage values to the right for neatness
    }

    utils.Info("PodListPanel setup completed with Kubernetes data.")
    return table
}



