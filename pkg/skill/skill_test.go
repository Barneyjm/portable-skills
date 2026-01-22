package skill

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestNewInput(t *testing.T) {
	// Create a temp file for testing
	tmpFile, err := os.CreateTemp("", "test-input-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	_, _ = tmpFile.WriteString("test content")
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// Test with file path
	input, err := NewInput(tmpFile.Name())
	if err != nil {
		t.Errorf("NewInput failed for file: %v", err)
	}
	if input.Path != tmpFile.Name() {
		t.Errorf("Path = %q, want %q", input.Path, tmpFile.Name())
	}
	_ = input.Close()

	// Test with non-existent file
	_, err = NewInput("/nonexistent/file/path")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test with dash (stdin)
	input, err = NewInput("-")
	if err != nil {
		t.Errorf("NewInput failed for stdin: %v", err)
	}
	if input.Path != "" {
		t.Errorf("Path for stdin should be empty, got %q", input.Path)
	}
	if input.Reader != os.Stdin {
		t.Error("Reader should be os.Stdin")
	}
}

func TestInputGetPath(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test-input-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	_, _ = tmpFile.WriteString("test content")
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	input, _ := NewInput(tmpFile.Name())
	path, err := input.GetPath(".txt")
	if err != nil {
		t.Errorf("GetPath failed: %v", err)
	}
	if path != tmpFile.Name() {
		t.Errorf("GetPath = %q, want %q", path, tmpFile.Name())
	}
	_ = input.Close()
}

func TestNewOutput(t *testing.T) {
	// Test with dash (stdout)
	output, err := NewOutput("-")
	if err != nil {
		t.Errorf("NewOutput failed for stdout: %v", err)
	}
	if output.Writer != os.Stdout {
		t.Error("Writer should be os.Stdout")
	}

	// Test with file path
	tmpFile, err := os.CreateTemp("", "test-output-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	output, err = NewOutput(tmpFile.Name())
	if err != nil {
		t.Errorf("NewOutput failed for file: %v", err)
	}
	if output.Path != tmpFile.Name() {
		t.Errorf("Path = %q, want %q", output.Path, tmpFile.Name())
	}
	_ = output.Close()
}

func TestResult(t *testing.T) {
	result := Result{
		Success: true,
		Data:    map[string]int{"count": 42},
		Output:  "/path/to/output",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal Result: %v", err)
	}

	var decoded Result
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Result: %v", err)
	}

	if !decoded.Success {
		t.Error("Success should be true")
	}
	if decoded.Output != "/path/to/output" {
		t.Errorf("Output = %q, want %q", decoded.Output, "/path/to/output")
	}
}

func TestManifestGenerateSKILLMD(t *testing.T) {
	manifest := Manifest{
		Name:        "test-skill",
		Description: "A test skill for testing",
		Version:     "1.0.0",
		Commands: []Command{
			{
				Name:        "test",
				Description: "Run a test",
				Usage:       "test-skill test <input>",
				Flags: []Flag{
					{Name: "output", Short: "o", Description: "Output path", Type: "string"},
					{Name: "verbose", Short: "v", Description: "Verbose output", Type: "bool", Default: "false"},
				},
				Examples: []string{
					"test-skill test input.txt",
					"test-skill test input.txt -o output.txt",
				},
			},
		},
		Formats: []string{"txt", "json", "yaml"},
	}

	md := manifest.GenerateSKILLMD()

	// Check for expected content
	expectedStrings := []string{
		"---",
		"name: test-skill",
		"description: A test skill for testing",
		"version: 1.0.0",
		"## Available Commands",
		"### test",
		"Run a test",
		"`-o, --output`",
		"`-v, --verbose`",
		"## Supported Formats",
		"txt, json, yaml",
		"## Installation",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(md, expected) {
			t.Errorf("Generated SKILL.md missing %q", expected)
		}
	}
}

func TestToTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test-skill", "Test Skill"},
		{"my_skill_name", "My Skill Name"},
		{"simple", "Simple"},
		{"ALLCAPS", "Allcaps"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toTitle(tt.input)
			if result != tt.expected {
				t.Errorf("toTitle(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInputClose(t *testing.T) {
	// Test closing a temp file input
	tmpFile, err := os.CreateTemp("", "test-close-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	_, _ = tmpFile.WriteString("test content")
	_ = tmpFile.Close()

	// Create input with temp file flag
	input := &Input{
		Path:   tmpFile.Name(),
		IsTemp: true,
	}

	// Close should remove the temp file
	err = input.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify file is removed
	if _, err := os.Stat(tmpFile.Name()); !os.IsNotExist(err) {
		t.Error("Temp file should have been removed")
		_ = os.Remove(tmpFile.Name()) // Clean up if test failed
	}
}

// Mock test for WriteJSON - can't easily test stdout
func TestWriteJSONFormat(t *testing.T) {
	data := map[string]interface{}{
		"name":  "test",
		"count": 42,
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(data)

	output := buf.String()
	if !strings.Contains(output, "\"name\": \"test\"") {
		t.Error("JSON output should contain formatted name")
	}
	if !strings.Contains(output, "\"count\": 42") {
		t.Error("JSON output should contain count")
	}
}
