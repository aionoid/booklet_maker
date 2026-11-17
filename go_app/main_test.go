package main

import (
	"os"
	"strings"
	"testing"
)

func TestCLIParsesGUIFlagWithoutError(t *testing.T) {
	// Save original os.Args
	origArgs := os.Args

	// Test with -gui=true flag (this would normally launch GUI, but we're testing flag parsing)
	os.Args = []string{"cmd", "-gui=false"} // Use false to prevent actual GUI launch
	defer func() {
		// Restore original os.Args
		os.Args = origArgs
	}()

	cli := &CLI{}

	// This should not return an error during flag parsing
	err := cli.Run([]string{"cmd", "-gui=false"})

	// Since we didn't provide an input file, we expect an error about missing input file
	// but not about flag parsing
	if err != nil {
		if !strings.Contains(err.Error(), "input file is required") {
			t.Errorf("Expected error about missing input file, got: %v", err)
		}
		// This is expected - missing input file error is OK
	}

	t.Log("CLI flag parsing worked correctly")
}

func TestCLIParsesAllFlagsWithoutError(t *testing.T) {
	// Save original os.Args
	origArgs := os.Args

	// Test with all flags but no input file (will cause expected error)
	os.Args = []string{"cmd", "-input", "", "-pages", "2", "-direction", "LTR", "-sections", "6", "-blank", "1"}
	defer func() {
		// Restore original os.Args
		os.Args = origArgs
	}()

	cli := &CLI{}

	// This should not return an error during flag parsing
	err := cli.Run([]string{"cmd", "-input", "", "-pages", "2", "-direction", "LTR", "-sections", "6", "-blank", "1"})

	// Since we provided an empty input file, we expect an error about missing input file
	if err != nil {
		if !strings.Contains(err.Error(), "input file is required") {
			t.Errorf("Expected error about missing input file, got: %v", err)
		}
		// This is expected - missing input file error is OK
	}

	t.Log("CLI flag parsing worked correctly for all flags")
}
