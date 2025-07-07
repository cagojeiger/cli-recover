package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type LogsView struct {
	app      *App
	table    *tview.Table
	logIDs   []string
	selected int
}

func NewLogsView(app *App) *LogsView {
	l := &LogsView{
		app:    app,
		table:  tview.NewTable(),
		logIDs: make([]string, 0),
	}

	l.setupTable()
	l.loadLogs()

	l.table.SetBorder(true).SetTitle(" Operation Logs ")
	l.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			l.app.RemovePage("logs")
			l.app.ShowPage("main")
			return nil
		case tcell.KeyEnter:
			if l.selected > 0 && l.selected <= len(l.logIDs) {
				l.showLogDetail(l.logIDs[l.selected-1])
			}
			return nil
		}
		return event
	})

	return l
}

func (l *LogsView) setupTable() {
	l.table.
		SetBorders(true).
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectedFunc(func(row, col int) {
			l.selected = row
		})

	headers := []string{"ID", "TYPE", "PROVIDER", "STATUS", "START TIME", "DURATION"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetExpansion(1)
		l.table.SetCell(0, col, cell)
	}
}

func (l *LogsView) loadLogs() {
	cmd, err := l.app.ExecuteCLI("logs", "list")
	if err != nil {
		l.showError(fmt.Sprintf("Failed to execute command: %v", err))
		return
	}

	output, err := cmd.Output()
	if err != nil {
		l.showError(fmt.Sprintf("Failed to list logs: %v", err))
		return
	}

	lines := strings.Split(string(output), "\n")
	row := 1
	l.logIDs = make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "ID") || strings.Contains(line, "----") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 6 {
			l.logIDs = append(l.logIDs, fields[0])
			
			for col := 0; col < 6 && col < len(fields); col++ {
				cell := tview.NewTableCell(fields[col]).
					SetAlign(tview.AlignLeft).
					SetExpansion(1)
				
				if col == 3 && fields[col] == "completed" {
					cell.SetTextColor(tcell.ColorGreen)
				} else if col == 3 && fields[col] == "failed" {
					cell.SetTextColor(tcell.ColorRed)
				} else if col == 3 && fields[col] == "running" {
					cell.SetTextColor(tcell.ColorYellow)
				}
				
				l.table.SetCell(row, col, cell)
			}
			row++
		}
	}

	if row == 1 {
		cell := tview.NewTableCell("No logs found").
			SetAlign(tview.AlignCenter).
			SetExpansion(6)
		l.table.SetCell(1, 0, cell)
	} else {
		info := tview.NewTableCell("[dim]Press Enter to view details, Escape to go back[white]").
			SetAlign(tview.AlignCenter).
			SetExpansion(6)
		l.table.SetCell(row, 0, info)
	}
}

func (l *LogsView) showLogDetail(logID string) {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)

	textView.SetBorder(true).
		SetTitle(fmt.Sprintf(" Log: %s ", logID)).
		SetTitleAlign(tview.AlignCenter)

	cmd, err := l.app.ExecuteCLI("logs", "show", logID)
	if err != nil {
		textView.SetText(fmt.Sprintf("[red]Failed to execute command: %v[white]", err))
	} else {
		output, err := cmd.Output()
		if err != nil {
			textView.SetText(fmt.Sprintf("[red]Failed to show log: %v[white]", err))
		} else {
			textView.SetText(string(output))
		}
	}

	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			l.app.RemovePage("log-detail")
			return nil
		}
		return event
	})

	l.app.AddPage("log-detail", textView, true, true)
}

func (l *LogsView) showError(message string) {
	cell := tview.NewTableCell(fmt.Sprintf("[red]Error: %s[white]", message)).
		SetAlign(tview.AlignCenter).
		SetExpansion(6)
	l.table.SetCell(1, 0, cell)
}

func (l *LogsView) GetView() tview.Primitive {
	return l.table
}