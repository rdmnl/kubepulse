// KubePulse - Kubernetes Cluster Monitor (TUI)
//
// Author: Erdem Unal
// Year: 2024
// Version: 0.1.0
// License: MIT

package panels

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rdmnl/kubepulse/pkg/kubernetes"
	"github.com/rivo/tview"
)

// SetupNodeListPanel sets up the node list panel with data from Kubernetes
func SetupNodeListPanel(client kubernetes.KubernetesClient) *tview.Table {
    table := tview.NewTable()

    table.SetBorders(false).
        SetSelectable(true, false).
        SetFixed(1, 0).
        SetBackgroundColor(tcell.ColorBlack).
        SetBorder(true).
        SetBorderColor(tcell.ColorLightCyan)

    table.SetCell(0, 0, tview.NewTableCell("Node Name").
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

    nodes, err := client.GetNodes()
    if err != nil {
        return table
    }

    for row, node := range nodes {
        cpuUsage, memoryUsage, err := client.GetNodeMetrics(node)
        if err != nil {
            cpuUsage = "N/A"
            memoryUsage = "N/A"
        }

        table.SetCell(row+1, 0, tview.NewTableCell(node).
            SetTextColor(tcell.ColorLightYellow).
            SetSelectable(true).
            SetAlign(tview.AlignLeft))
        table.SetCell(row+1, 1, tview.NewTableCell(cpuUsage).
            SetTextColor(tcell.ColorLightGreen).
            SetAlign(tview.AlignRight))
        table.SetCell(row+1, 2, tview.NewTableCell(memoryUsage).
            SetTextColor(tcell.ColorLightBlue).
            SetAlign(tview.AlignRight))
    }

    return table
}
