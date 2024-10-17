// KubePulse - Kubernetes Cluster Monitor (TUI)
//
// Author: Erdem Unal
// Year: 2024
// Version: 0.1.0
// License: MIT

package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rdmnl/kubepulse/utils"
	"github.com/rivo/tview"
)

func SetStyles() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
	tview.Styles.ContrastBackgroundColor = tcell.ColorBlack
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.BorderColor = tcell.ColorLightCyan
	tview.Styles.TitleColor = tcell.ColorLightYellow
	tview.Styles.GraphicsColor = tcell.ColorLightCyan
	utils.Info("Styles setup completed.")
}