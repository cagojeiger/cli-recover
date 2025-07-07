package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
)

type App struct {
	app        *tview.Application
	pages      *tview.Pages
	mainMenu   *MainMenu
	cliPath    string
	logChannel chan string
}

func NewApp(cliPath string) *App {
	app := &App{
		app:        tview.NewApplication(),
		pages:      tview.NewPages(),
		cliPath:    cliPath,
		logChannel: make(chan string, 100),
	}

	app.mainMenu = NewMainMenu(app)
	app.pages.AddPage("main", app.mainMenu.GetView(), true, true)

	return app
}

func (a *App) Run() error {
	a.app.SetRoot(a.pages, true)
	return a.app.Run()
}

func (a *App) Stop() {
	a.app.Stop()
}

func (a *App) QueueUpdateDraw(f func()) *tview.Application {
	return a.app.QueueUpdateDraw(f)
}

func (a *App) ShowPage(name string) {
	a.pages.SwitchToPage(name)
}

func (a *App) AddPage(name string, item tview.Primitive, resize, visible bool) {
	a.pages.AddPage(name, item, resize, visible)
}

func (a *App) RemovePage(name string) {
	a.pages.RemovePage(name)
}

func (a *App) ExecuteCLI(args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(a.cliPath, args...)
	return cmd, nil
}

func (a *App) ShowError(title, message string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("[red]%s[white]\n\n%s", title, message)).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("error")
			a.pages.SwitchToPage("main")
		})

	a.pages.AddPage("error", modal, true, true)
}

func (a *App) ShowInfo(title, message string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("[green]%s[white]\n\n%s", title, message)).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("info")
			a.pages.SwitchToPage("main")
		})

	a.pages.AddPage("info", modal, true, true)
}

func (a *App) Log(message string) {
	select {
	case a.logChannel <- message:
	default:
	}
}

func parseNamespaces(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	namespaces := make([]string, 0, len(lines))
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "NAME") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				namespaces = append(namespaces, fields[0])
			}
		}
	}
	
	return namespaces
}

func parsePods(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	pods := make([]string, 0, len(lines))
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "NAME") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				pods = append(pods, fields[0])
			}
		}
	}
	
	return pods
}