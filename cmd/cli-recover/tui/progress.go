package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ProgressView struct {
	app      *App
	textView *tview.TextView
	frame    *tview.Frame
	done     bool
	title    string
}

func NewProgressView(app *App, title string) *ProgressView {
	p := &ProgressView{
		app:      app,
		textView: tview.NewTextView(),
		title:    title,
	}

	p.textView.
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			p.app.QueueUpdateDraw(func() {
				p.textView.ScrollToEnd()
			})
		})

	p.frame = tview.NewFrame(p.textView).
		SetBorders(0, 0, 0, 0, 0, 0)

	p.frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if p.done && (event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyEscape) {
			p.app.RemovePage("progress")
			p.app.ShowPage("main")
			return nil
		}
		return event
	})

	return p
}

func (p *ProgressView) AddLine(line string) {
	current := p.textView.GetText(false)
	if current != "" {
		current += "\n"
	}
	p.textView.SetText(current + line)
}

func (p *ProgressView) SetDone() {
	p.done = true
	p.AddLine("\n[yellow]Press Enter or Escape to continue...[white]")
}

func (p *ProgressView) GetView() tview.Primitive {
	p.textView.SetBorder(true).SetTitle(p.title)
	return p.textView
}
