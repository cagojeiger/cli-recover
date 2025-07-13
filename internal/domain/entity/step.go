package entity

// Step represents a single command in a pipeline
type Step struct {
	Name   string
	Run    string
	Input  string
	Output string
}