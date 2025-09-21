package man

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestNewManCmd(t *testing.T) {
	// test each field of NewManCmd Cobra command

	expectedUse := "man"
	if NewManCmd().Use != expectedUse {
		t.Errorf("Unexpected command use text: got %q, expected %q", NewManCmd().Use, expectedUse)
	}

	expectedShort := "Generates files2prompt's command line manpages"
	if NewManCmd().Short != expectedShort {
		t.Errorf("Unexpected command short text: got %q, expected %q", NewManCmd().Short, expectedShort)
	}

	expectedSilenceUsage := true
	if NewManCmd().SilenceUsage != expectedSilenceUsage {
		t.Errorf("Unexpected command SilenceUsage field: got %t, expected %t", NewManCmd().SilenceUsage, expectedSilenceUsage)
	}

	expectedDisableFlagsInUseLine := true
	if NewManCmd().DisableFlagsInUseLine != expectedDisableFlagsInUseLine {
		t.Errorf("Unexpected command DisableFlagsInUseLine field: got %t, expected %t", NewManCmd().DisableFlagsInUseLine, expectedDisableFlagsInUseLine)
	}

	expectedHidden := true
	if NewManCmd().Hidden != expectedHidden {
		t.Errorf("Unexpected command Hidden field: got %t, expected %t", NewManCmd().Hidden, expectedHidden)
	}

	// Test that RunE function works correctly
	// Capture stdout to test the man page output
	cmd := NewManCmd()

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

	// Verify output is not empty
	if output == "" {
		t.Error("Command RunE produced no output")
	}

	// Debug: print first few lines to understand format
	lines := strings.Split(output, "\n")
	if len(lines) > 10 {
		t.Logf("First 10 lines of output: %v", lines[:10])
	} else {
		t.Logf("All lines of output: %v", lines)
	}

	// Verify output contains expected man page elements
	// Man pages typically start with .TH (title header) and contain .SH (section headers)
	if !strings.Contains(output, ".TH") {
		t.Error("Output does not contain man page title header (.TH)")
	}

	if !strings.Contains(output, ".SH") {
		t.Error("Output does not contain man page section headers (.SH)")
	}

	// Verify output contains the command name
	if !strings.Contains(output, "files2prompt") {
		t.Error("Output does not contain command name 'files2prompt'")
	}

	// Verify output contains expected sections for the man command itself
	// The man command generates its own man page, not the full application man page
	expectedSections := []string{
		"NAME",
		"SYNOPSIS",
		"DESCRIPTION",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output does not contain expected section: %s", section)
		}
	}

	// Verify specific content that should be present for the man command
	if !strings.Contains(output, "Generates files2prompt's command line manpages") {
		t.Error("Output does not contain expected man command description")
	}

	// Verify it contains the command name "man"
	if !strings.Contains(output, "man - Generates files2prompt's command line manpages") {
		t.Error("Output does not contain expected NAME section content")
	}
}
