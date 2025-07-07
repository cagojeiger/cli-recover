package tui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
)

type RestoreFlow struct {
	app        *App
	form       *tview.Form
	namespace  string
	pod        string
	backupFile string
	targetPath string
}

func NewRestoreFlow(app *App) *RestoreFlow {
	r := &RestoreFlow{
		app:        app,
		form:       tview.NewForm(),
		targetPath: "/",
	}

	r.setupForm()
	return r
}

func (r *RestoreFlow) setupForm() {
	r.form.
		SetBorder(true).
		SetTitle(" Restore Backup ").
		SetTitleAlign(tview.AlignCenter)

	backupFiles := r.getBackupFiles()
	if len(backupFiles) == 0 {
		backupFiles = []string{"No backups found"}
	}

	namespaces := r.getNamespaces()
	if len(namespaces) == 0 {
		namespaces = []string{"default"}
	}

	r.form.
		AddDropDown("Backup File", backupFiles, 0, func(option string, index int) {
			r.backupFile = option
		}).
		AddDropDown("Target Namespace", namespaces, 0, func(option string, index int) {
			r.namespace = option
			r.updatePodList()
		}).
		AddDropDown("Target Pod", []string{}, 0, func(option string, index int) {
			r.pod = option
		}).
		AddInputField("Target Path", r.targetPath, 40, nil, func(text string) {
			r.targetPath = text
		}).
		AddButton("Start Restore", func() {
			r.executeRestore()
		}).
		AddButton("Cancel", func() {
			r.app.RemovePage("restore")
			r.app.ShowPage("main")
		})

	r.namespace = namespaces[0]
	r.updatePodList()
}

func (r *RestoreFlow) getBackupFiles() []string {
	var files []string
	
	entries, err := os.ReadDir(".")
	if err != nil {
		return files
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if strings.HasSuffix(name, ".tar") || strings.HasSuffix(name, ".tar.gz") {
				files = append(files, name)
			}
		}
	}
	
	return files
}

func (r *RestoreFlow) getNamespaces() []string {
	cmd := exec.Command("kubectl", "get", "namespaces", "-o", "custom-columns=NAME:.metadata.name", "--no-headers")
	output, err := cmd.Output()
	if err != nil {
		return []string{"default"}
	}

	return parseNamespaces(string(output))
}

func (r *RestoreFlow) updatePodList() {
	podDropDown := r.form.GetFormItem(2).(*tview.DropDown)
	
	pods := r.getPods()
	if len(pods) == 0 {
		pods = []string{"No pods found"}
	}
	
	podDropDown.SetOptions(pods, func(option string, index int) {
		r.pod = option
	})
	
	if len(pods) > 0 {
		r.pod = pods[0]
		podDropDown.SetCurrentOption(0)
	}
}

func (r *RestoreFlow) getPods() []string {
	cmd := exec.Command("kubectl", "get", "pods", "-n", r.namespace, "-o", "custom-columns=NAME:.metadata.name", "--no-headers")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	return parsePods(string(output))
}

func (r *RestoreFlow) executeRestore() {
	if r.backupFile == "" || r.backupFile == "No backups found" {
		r.app.ShowError("Error", "Please select a valid backup file")
		return
	}
	if r.pod == "" || r.pod == "No pods found" {
		r.app.ShowError("Error", "Please select a valid pod")
		return
	}
	if r.targetPath == "" {
		r.app.ShowError("Error", "Please enter a target path")
		return
	}

	progressView := NewProgressView(r.app, "Restore Progress")
	r.app.AddPage("progress", progressView.GetView(), true, true)
	r.app.ShowPage("progress")

	go r.runRestore(progressView)
}

func (r *RestoreFlow) runRestore(progress *ProgressView) {
	args := []string{"restore", "filesystem", r.pod, r.backupFile}
	args = append(args, "--namespace", r.namespace)
	args = append(args, "--target", r.targetPath)

	cmd, err := r.app.ExecuteCLI(args...)
	if err != nil {
		r.app.QueueUpdateDraw(func() {
			r.app.RemovePage("progress")
			r.app.ShowError("Error", fmt.Sprintf("Failed to start restore: %v", err))
		})
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		r.app.QueueUpdateDraw(func() {
			r.app.RemovePage("progress")
			r.app.ShowError("Error", fmt.Sprintf("Failed to create stdout pipe: %v", err))
		})
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		r.app.QueueUpdateDraw(func() {
			r.app.RemovePage("progress")
			r.app.ShowError("Error", fmt.Sprintf("Failed to create stderr pipe: %v", err))
		})
		return
	}

	if err := cmd.Start(); err != nil {
		r.app.QueueUpdateDraw(func() {
			r.app.RemovePage("progress")
			r.app.ShowError("Error", fmt.Sprintf("Failed to start restore: %v", err))
		})
		return
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			r.app.QueueUpdateDraw(func() {
				progress.AddLine(line)
			})
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			r.app.QueueUpdateDraw(func() {
				progress.AddLine(fmt.Sprintf("[red]%s[white]", line))
			})
		}
	}()

	err = cmd.Wait()
	r.app.QueueUpdateDraw(func() {
		if err != nil {
			progress.AddLine(fmt.Sprintf("\n[red]Restore failed: %v[white]", err))
		} else {
			progress.AddLine(fmt.Sprintf("\n[green]Restore completed successfully![white]"))
			progress.AddLine(fmt.Sprintf("Restored to: %s:%s", r.pod, r.targetPath))
		}
		progress.SetDone()
	})
}

func (r *RestoreFlow) GetView() tview.Primitive {
	return r.form
}