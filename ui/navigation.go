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
	"github.com/rdmnl/kubepulse/utils"
	"github.com/rivo/tview"
)

func SetupNavigation(app *tview.Application, controller *UIController) {
	utils.Info("Navigation setup started.")

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		utils.Info(fmt.Sprintf("Key pressed: %v", event))

		if _, ok := app.GetFocus().(*tview.InputField); ok {
			return event
		}

		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'p':
				controller.setPanelFocus(0)
			case 'n':
				controller.setPanelFocus(1)
			case 'd':
				controller.setPanelFocus(2)
			case 'l':
				if controller.UIManager.PodListPanel.HasFocus() {
					controller.HandleLogView()
				}
			case 'b':
				controller.HandleBackNavigation()
			case 'f':
				controller.HandleNamespaceFilter()
			case 'q':
				utils.Info("Quit key pressed")
				app.Stop()
			}
		case tcell.KeyEnter:
			if controller.UIManager.PodListPanel.HasFocus() {
				controller.HandlePodSelection()
			} else if controller.UIManager.NodeListPanel.HasFocus() {
				controller.HandleNodeSelection()
			}
		}
		return event
	})
	utils.Info("Navigation setup completed.")
}
