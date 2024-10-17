// KubePulse - Kubernetes Cluster Monitor (TUI)
//
// Author: Erdem Unal
// Year: 2024
// Version: 0.1.0
// License: MIT

package utils

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CreateTextView creates a reusable TextView component with customizable attributes
func CreateTextView(defaultText string, borderColor tcell.Color, textColor tcell.Color, title string) *tview.TextView {
	textView := tview.NewTextView()
	textView.SetText(defaultText).
		SetDynamicColors(true).
		SetTextColor(textColor).
		SetBackgroundColor(tcell.ColorBlack).
		SetBorder(true).
		SetBorderColor(borderColor).
		SetTitle(title)

	return textView
}
