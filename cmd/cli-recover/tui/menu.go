package tui

import (
	"github.com/rivo/tview"
)

type MainMenu struct {
	app  *App
	list *tview.List
}

func NewMainMenu(app *App) *MainMenu {
	m := &MainMenu{
		app:  app,
		list: tview.NewList(),
	}

	m.list.
		SetBorder(true).
		SetTitle(" CLI-Recover TUI ").
		SetTitleAlign(tview.AlignCenter)

	m.list.
		AddItem("Backup", "Create a new backup", 'b', func() {
			m.app.AddPage("backup", NewBackupFlow(m.app).GetView(), true, true)
			m.app.ShowPage("backup")
		}).
		AddItem("Restore", "Restore from backup", 'r', func() {
			m.app.AddPage("restore", NewRestoreFlow(m.app).GetView(), true, true)
			m.app.ShowPage("restore")
		}).
		AddItem("List Backups", "View existing backups", 'l', func() {
			m.app.AddPage("list", NewListBackups(m.app).GetView(), true, true)
			m.app.ShowPage("list")
		}).
		AddItem("View Logs", "View operation logs", 'v', func() {
			m.app.AddPage("logs", NewLogsView(m.app).GetView(), true, true)
			m.app.ShowPage("logs")
		}).
		AddItem("Exit", "Exit the application", 'q', func() {
			m.app.Stop()
		})

	m.list.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		switch shortcut {
		case 'q':
			m.app.Stop()
		}
	})

	return m
}

func (m *MainMenu) GetView() tview.Primitive {
	return m.list
}