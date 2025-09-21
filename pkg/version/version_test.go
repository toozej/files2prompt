package version

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	// Set up test data
	expectedInfo := Info{
		Commit:  Commit,
		Version: Version,
		Branch:  Branch,
		BuiltAt: BuiltAt,
		Builder: Builder,
	}

	// Call Get() and check the result
	Info, err := Get()
	if err != nil {
		t.Errorf("Error getting Info object: %v", err)
	}
	if Info != expectedInfo {
		t.Errorf("Loaded Info object does not match expected. Got %v, expected %v", Info, expectedInfo)
	}
}

func TestCommand(t *testing.T) {
	// Test that Command() returns a properly configured cobra command
	cmd := Command()

	// Test command basic properties
	if cmd.Use != "version" {
		t.Errorf("Unexpected command use: got %q, expected %q", cmd.Use, "version")
	}

	if cmd.Short != "Print the version." {
		t.Errorf("Unexpected command short: got %q, expected %q", cmd.Short, "Print the version.")
	}

	if cmd.Long != "Print the version and build information." {
		t.Errorf("Unexpected command long: got %q, expected %q", cmd.Long, "Print the version and build information.")
	}

	// Test that RunE function works correctly
	// Capture stdout to test the JSON output
	// Note: This test will use the actual version variables from the package

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the command
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Command RunE failed: %v", err)
	}

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Errorf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Verify output is valid JSON
	var result Info
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Command output is not valid JSON: %v", err)
	}

	// Verify the JSON contains expected fields
	// Version should always be present (defaults to "local" for development)
	if result.Version == "" {
		t.Error("JSON output missing Version field")
	}

	// For development builds, these fields are typically empty (populated via ldflags)
	// The test should pass regardless of whether they're empty or populated
	// This is the expected behavior as documented in the package
}
