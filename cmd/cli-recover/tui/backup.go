package tui

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
)

type BackupFlow struct {
	app       *App
	form      *tview.Form
	namespace string
	pod       string
	path      string
	output    string
	compress  bool
}

func NewBackupFlow(app *App) *BackupFlow {
	b := &BackupFlow{
		app:      app,
		form:     tview.NewForm(),
		path:     "/",
		compress: true,
	}

	b.setupForm()
	return b
}

func (b *BackupFlow) setupForm() {
	b.form.
		SetBorder(true).
		SetTitle(" Create Backup ").
		SetTitleAlign(tview.AlignCenter)

	namespaces := b.getNamespaces()
	if len(namespaces) == 0 {
		namespaces = []string{"default"}
	}

	b.form.
		AddDropDown("Namespace", namespaces, 0, func(option string, index int) {
			b.namespace = option
			b.updatePodList()
		}).
		AddDropDown("Pod", []string{}, 0, func(option string, index int) {
			b.pod = option
		}).
		AddInputField("Path", b.path, 40, nil, func(text string) {
			b.path = text
		}).
		AddInputField("Output File", "", 40, nil, func(text string) {
			b.output = text
		}).
		AddCheckbox("Compress (.tar.gz)", b.compress, func(checked bool) {
			b.compress = checked
		}).
		AddButton("Start Backup", func() {
			b.executeBackup()
		}).
		AddButton("Cancel", func() {
			b.app.RemovePage("backup")
			b.app.ShowPage("main")
		})

	b.namespace = namespaces[0]
	b.updatePodList()
}

func (b *BackupFlow) getNamespaces() []string {
	cmd := exec.Command("kubectl", "get", "namespaces", "-o", "custom-columns=NAME:.metadata.name", "--no-headers")
	output, err := cmd.Output()
	if err != nil {
		return []string{"default"}
	}

	return parseNamespaces(string(output))
}

func (b *BackupFlow) updatePodList() {
	podDropDown := b.form.GetFormItem(1).(*tview.DropDown)
	
	pods := b.getPods()
	if len(pods) == 0 {
		pods = []string{"No pods found"}
	}
	
	podDropDown.SetOptions(pods, func(option string, index int) {
		b.pod = option
	})
	
	if len(pods) > 0 {
		b.pod = pods[0]
		podDropDown.SetCurrentOption(0)
	}
}

func (b *BackupFlow) getPods() []string {
	cmd := exec.Command("kubectl", "get", "pods", "-n", b.namespace, "-o", "custom-columns=NAME:.metadata.name", "--no-headers")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	return parsePods(string(output))
}

func (b *BackupFlow) executeBackup() {
	if b.pod == "" || b.pod == "No pods found" {
		b.app.ShowError("Error", "Please select a valid pod")
		return
	}
	if b.path == "" {
		b.app.ShowError("Error", "Please enter a path to backup")
		return
	}
	if b.output == "" {
		if b.compress {
			b.output = fmt.Sprintf("backup-%s-%s.tar.gz", b.pod, strings.ReplaceAll(b.path, "/", "_"))
		} else {
			b.output = fmt.Sprintf("backup-%s-%s.tar", b.pod, strings.ReplaceAll(b.path, "/", "_"))
		}
	}

	progressView := NewProgressView(b.app, "Backup Progress")
	b.app.AddPage("progress", progressView.GetView(), true, true)
	b.app.ShowPage("progress")

	go b.runBackup(progressView)
}

func (b *BackupFlow) runBackup(progress *ProgressView) {
	args := []string{"backup", "filesystem", b.pod, b.path}
	args = append(args, "--namespace", b.namespace)
	args = append(args, "--output", b.output)
	if b.compress {
		args = append(args, "--compress", "gzip")
	}

	cmd, err := b.app.ExecuteCLI(args...)
	if err != nil {
		b.app.QueueUpdateDraw(func() {
			b.app.RemovePage("progress")
			b.app.ShowError("Error", fmt.Sprintf("Failed to start backup: %v", err))
		})
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		b.app.QueueUpdateDraw(func() {
			b.app.RemovePage("progress")
			b.app.ShowError("Error", fmt.Sprintf("Failed to create stdout pipe: %v", err))
		})
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		b.app.QueueUpdateDraw(func() {
			b.app.RemovePage("progress")
			b.app.ShowError("Error", fmt.Sprintf("Failed to create stderr pipe: %v", err))
		})
		return
	}

	if err := cmd.Start(); err != nil {
		b.app.QueueUpdateDraw(func() {
			b.app.RemovePage("progress")
			b.app.ShowError("Error", fmt.Sprintf("Failed to start backup: %v", err))
		})
		return
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			b.app.QueueUpdateDraw(func() {
				progress.AddLine(line)
			})
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			b.app.QueueUpdateDraw(func() {
				progress.AddLine(fmt.Sprintf("[red]%s[white]", line))
			})
		}
	}()

	err = cmd.Wait()
	b.app.QueueUpdateDraw(func() {
		if err != nil {
			progress.AddLine(fmt.Sprintf("\n[red]Backup failed: %v[white]", err))
		} else {
			progress.AddLine(fmt.Sprintf("\n[green]Backup completed successfully![white]"))
			progress.AddLine(fmt.Sprintf("Output file: %s", b.output))
		}
		progress.SetDone()
	})
}

func (b *BackupFlow) GetView() tview.Primitive {
	return b.form
}