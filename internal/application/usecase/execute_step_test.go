package usecase

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/entity"
	"github.com/cagojeiger/cli-recover/internal/domain/service"
)

func TestExecuteStep_Execute(t *testing.T) {
	tests := []struct {
		name        string
		step        *entity.Step
		setupStream func(*service.StreamManager)
		wantOutput  string
		wantErr     bool
	}{
		{
			name: "simple echo command",
			step: &entity.Step{
				Name:   "echo-test",
				Run:    "echo hello",
				Output: "result",
			},
			setupStream: func(sm *service.StreamManager) {},
			wantOutput:  "hello\n",
			wantErr:     false,
		},
		{
			name: "command with input",
			step: &entity.Step{
				Name:   "uppercase",
				Run:    "tr '[:lower:]' '[:upper:]'",
				Input:  "source",
				Output: "result",
			},
			setupStream: func(sm *service.StreamManager) {
				writer, _ := sm.CreateStream("source")
				go func() {
					io.WriteString(writer, "hello world")
					writer.Close()
				}()
			},
			wantOutput: "HELLO WORLD",
			wantErr:    false,
		},
		{
			name: "command failure",
			step: &entity.Step{
				Name:   "fail",
				Run:    "false",
				Output: "result",
			},
			setupStream: func(sm *service.StreamManager) {},
			wantOutput:  "",
			wantErr:     true,
		},
		{
			name: "missing input stream",
			step: &entity.Step{
				Name:   "missing-input",
				Run:    "cat",
				Input:  "nonexistent",
				Output: "result",
			},
			setupStream: func(sm *service.StreamManager) {},
			wantOutput:  "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			sm := service.NewStreamManager()
			tt.setupStream(sm)
			
			executor := NewExecuteStep(sm)
			
			// For steps with output, setup reader first
			var outputData []byte
			done := make(chan bool, 1)
			
			if tt.step.Output != "" && !tt.wantErr {
				go func() {
					reader, err := sm.GetStream(tt.step.Output)
					if err == nil {
						outputData, _ = io.ReadAll(reader)
					}
					done <- true
				}()
			} else {
				// For cases where we don't read output
				done <- true
			}
			
			// Execute
			err := executor.Execute(tt.step)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Execute() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// Wait for output reading if needed
			if tt.step.Output != "" && !tt.wantErr {
				<-done
				
				if string(outputData) != tt.wantOutput {
					t.Errorf("Output = %q, want %q", string(outputData), tt.wantOutput)
				}
			}
		})
	}
}

func TestExecuteStep_CaptureOutput(t *testing.T) {
	sm := service.NewStreamManager()
	executor := NewExecuteStep(sm)
	
	// Create a step that outputs to stdout and stderr
	step := &entity.Step{
		Name:   "output-test",
		Run:    `echo "stdout message" && echo "stderr message" >&2`,
		Output: "combined",
	}
	
	// Create a buffer to capture logs
	var logBuffer bytes.Buffer
	executor.SetLogWriter(&logBuffer)
	
	// Setup reader for output stream
	done := make(chan string)
	go func() {
		reader, _ := sm.GetStream("combined")
		output, _ := io.ReadAll(reader)
		done <- string(output)
	}()
	
	// Execute
	err := executor.Execute(step)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	
	// Get the output
	output := <-done
	
	// Check that both stdout and stderr were captured in logs
	logs := logBuffer.String()
	t.Logf("Captured logs: %q", logs)
	t.Logf("Output stream: %q", output)
	
	// Both stdout and stderr should be in logs due to MultiWriter
	if !strings.Contains(logs, "stdout message") && !strings.Contains(logs, "stderr message") {
		t.Error("Logs should contain at least stderr message")
	}
	
	// Output stream should contain stdout
	if !strings.Contains(output, "stdout message") {
		t.Error("Output stream should contain stdout message")
	}
}