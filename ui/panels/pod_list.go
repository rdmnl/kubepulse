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

func SetupPodListPanel(client kubernetes.KubernetesClient) *tview.Table {
	table := tview.NewTable()

	table.SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetBackgroundColor(tcell.ColorBlack).
		SetBorder(true).
		SetBorderColor(tcell.ColorLightCyan)

	table.SetCell(0, 0, tview.NewTableCell("Pod Name").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	table.SetCell(0, 1, tview.NewTableCell("Namespace").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	table.SetCell(0, 2, tview.NewTableCell("CPU").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	table.SetCell(0, 3, tview.NewTableCell("Memory").
		SetTextColor(tcell.ColorWhite).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))

	pods, err := client.GetPods()
	if err != nil {
		utils.Info(fmt.Sprintf("Error fetching pods: %v", err))
		return table
	}

	for row, pod := range pods {
		if pod.Name == "" {
			continue
		}

		cpuUsage, memoryUsage, err := client.GetPodMetrics(pod)
		if err != nil {
			cpuUsage = "N/A"
			memoryUsage = "N/A"
			utils.Info(fmt.Sprintf("Error fetching metrics for pod %s/%s: %v", pod.Namespace, pod.Name, err))
		}

		table.SetCell(row+1, 0, tview.NewTableCell(pod.Name).
			SetTextColor(tcell.ColorLightYellow).
			SetBackgroundColor(tcell.ColorBlack).
			SetSelectable(true).
			SetAlign(tview.AlignLeft))

		table.SetCell(row+1, 1, tview.NewTableCell(pod.Namespace).
			SetTextColor(tcell.ColorLightGreen).
			SetBackgroundColor(tcell.ColorBlack).
			SetSelectable(false).
			SetAlign(tview.AlignLeft))

		table.SetCell(row+1, 2, tview.NewTableCell(cpuUsage).
			SetTextColor(tcell.ColorLightGreen).
			SetSelectable(false).
			SetAlign(tview.AlignRight))

		table.SetCell(row+1, 3, tview.NewTableCell(memoryUsage).
			SetTextColor(tcell.ColorLightBlue).
			SetSelectable(false).
			SetAlign(tview.AlignRight))
	}

	utils.Info("PodListPanel setup completed with Kubernetes data.")
	return table
}
