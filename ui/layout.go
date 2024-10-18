// KubePulse - Kubernetes Cluster Monitor (TUI)
//
// Author: Erdem Unal
// Year: 2024
// Version: 0.1.0
// License: MIT

package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rdmnl/kubepulse/pkg/kubernetes"
	"github.com/rdmnl/kubepulse/ui/panels"
	"github.com/rivo/tview"
)

type UIManager struct {
	Header        *tview.TextView
	NodeListPanel *tview.Table
	PodListPanel  *tview.Table
	DetailsPanel  *tview.TextView
	LogsViewPanel *tview.TextView
	StatusBar     *tview.TextView
	CurrentPanel  int
	SelectedPod   string
	SelectedNode  string
	Layout        *tview.Flex
}

func SetupUILayout(app *tview.Application, client kubernetes.KubernetesClient) (*UIManager, *tview.Flex) {
	header := SetupHeader()
	nodeListPanel := panels.SetupNodeListPanel(client)
	podListPanel := panels.SetupPodListPanel(client)
	detailsPanel := panels.SetupDetailsPanel()
	logsViewPanel := panels.SetupLogsViewPanel()
	statusBar := SetupStatusBar()

	uiManager := &UIManager{
		Header:        header,
		NodeListPanel: nodeListPanel,
		PodListPanel:  podListPanel,
		DetailsPanel:  detailsPanel,
		LogsViewPanel: logsViewPanel,
		StatusBar:     statusBar,
		CurrentPanel:  0,
	}

	nodePodDetailsColumn := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(uiManager.NodeListPanel, 0, 1, false).
		AddItem(uiManager.PodListPanel, 0, 2, true).
		AddItem(uiManager.DetailsPanel, 0, 1, false)

	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nodePodDetailsColumn, 0, 2, true).
		AddItem(uiManager.LogsViewPanel, 0, 3, false)

	fullLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(uiManager.Header, 3, 1, false).
		AddItem(mainLayout, 0, 1, true).
		AddItem(uiManager.StatusBar, 1, 1, false)

	uiManager.Layout = fullLayout

	app.SetFocus(uiManager.PodListPanel)

	return uiManager, fullLayout
}

// TODO Get version properly
func SetupHeader() *tview.TextView {
	header := tview.NewTextView()
	header.SetTextAlign(tview.AlignCenter).
		SetText(" KubePulse - Kubernetes Cluster Monitor ").
		SetDynamicColors(true).
		SetTextColor(tcell.ColorLightCyan).
		SetBackgroundColor(tcell.ColorBlack).
		SetBorder(true)
	return header
}

func SetupStatusBar() *tview.TextView {
	statusBar := tview.NewTextView()

	statusBar.SetText(
		quitInstruction + " | " +
			podShortcut + " | " +
			nodeShortcut + " | " +
			detailShortcut + " | " +
			logShortcut + " | " +
			filterNamespaceInstruction + " | " +
			backInstruction).
		SetDynamicColors(true).
		SetTextColor(tview.Styles.PrimaryTextColor).
		SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	return statusBar
}
