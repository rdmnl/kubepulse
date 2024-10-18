// KubePulse - Kubernetes Cluster Monitor (TUI)
//
// Author: Erdem Unal
// Year: 2024
// Version: 0.1.0
// License: MIT

package panels

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rdmnl/kubepulse/utils"
	"github.com/rivo/tview"
)

func SetupLogsViewPanel() *tview.TextView {
	return utils.CreateTextView("Logs:\n", tcell.ColorLightCyan, tcell.ColorLightGreen, "Logs")
}
