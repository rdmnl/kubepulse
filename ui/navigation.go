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

// SetupNavigation sets up key navigation for the UI
func SetupNavigation(app *tview.Application, controller *UIController) {
    utils.Info("Navigation setup started.")

    app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        utils.Info(fmt.Sprintf("Key pressed: %v", event))

        // Check if the current element is an input field and ignore global bindings if true
        if _, ok := app.GetFocus().(*tview.InputField); ok {
            return event // Allow typing without interference
        }

        switch event.Key() {
        case tcell.KeyRight:
            newPanelIndex := (controller.UIManager.CurrentPanel + 1) % 3 // Cycle to the next panel
            controller.setPanelFocus(newPanelIndex)

        case tcell.KeyLeft:
            newPanelIndex := (controller.UIManager.CurrentPanel - 1 + 3) % 3 // Cycle to the previous panel
            controller.setPanelFocus(newPanelIndex)

        case tcell.KeyEnter:
            if controller.UIManager.PodListPanel.HasFocus() {
                controller.HandlePodSelection()
            }

        case tcell.KeyRune:
            switch event.Rune() {
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
        }
        return event
    })
    utils.Info("Navigation setup completed.")
}