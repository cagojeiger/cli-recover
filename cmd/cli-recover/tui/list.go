package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ListBackups struct {
	app   *App
	table *tview.Table
}

func NewListBackups(app *App) *ListBackups {
	l := &ListBackups{
		app:   app,
		table: tview.NewTable(),
	}

	l.setupTable()
	l.loadBackups()

	l.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			l.app.RemovePage("list")
			l.app.ShowPage("main")
			return nil
		}
		return event
	})

	return l
}

func (l *ListBackups) setupTable() {
	l.table.
		SetBorders(true).
		SetSelectable(true, false).
		SetFixed(1, 0)

	headers := []string{"ID", "TYPE", "NAMESPACE", "POD", "PATH", "SIZE", "STATUS", "DATE"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetExpansion(1)
		l.table.SetCell(0, col, cell)
	}
}

func (l *ListBackups) loadBackups() {
	cmd, err := l.app.ExecuteCLI("list", "backups")
	if err != nil {
		l.showError(fmt.Sprintf("Failed to execute command: %v", err))
		return
	}

	output, err := cmd.Output()
	if err != nil {
		l.showError(fmt.Sprintf("Failed to list backups: %v", err))
		return
	}

	lines := strings.Split(string(output), "\n")
	row := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "ID") || strings.Contains(line, "---") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 8 {
			for col := 0; col < 8 && col < len(fields); col++ {
				cell := tview.NewTableCell(fields[col]).
					SetAlign(tview.AlignLeft).
					SetExpansion(1)

				if col == 6 && fields[col] == "completed" {
					cell.SetTextColor(tcell.ColorGreen)
				} else if col == 6 && fields[col] == "failed" {
					cell.SetTextColor(tcell.ColorRed)
				}

				l.table.SetCell(row, col, cell)
			}
			row++
		}
	}

	if row == 1 {
		cell := tview.NewTableCell("No backups found").
			SetAlign(tview.AlignCenter).
			SetExpansion(8)
		l.table.SetCell(1, 0, cell)
	}
}

func (l *ListBackups) showError(message string) {
	cell := tview.NewTableCell(fmt.Sprintf("[red]Error: %s[white]", message)).
		SetAlign(tview.AlignCenter).
		SetExpansion(8)
	l.table.SetCell(1, 0, cell)
}

func (l *ListBackups) GetView() tview.Primitive {
	l.table.SetBorder(true).SetTitle(" Backup List ")
	return l.table
}
