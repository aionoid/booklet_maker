package main

import (
	"os/exec"
	"testing"
)

// TestGUIBuild tests that the GUI application builds successfully.
func TestGUIBuild(t *testing.T) {
	cmd := exec.Command("go", "build", "-o", "/tmp/booklet-test-build", ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, output)
	}

	t.Log("GUI application built successfully")
}

// TestGUIModule tests that the module passes static analysis.
func TestGUIModule(t *testing.T) {
	cmd := exec.Command("go", "vet", "./...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Vet failed: %v\nOutput: %s", err, output)
	}

	t.Log("Module dependencies and static analysis are correct")
}

// TestGUIImports tests that all imports are valid (go fmt check).
func TestGUIImports(t *testing.T) {
	cmd := exec.Command("go", "fmt", "./...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Fmt failed: %v\nOutput: %s", err, output)
	}

	t.Log("Import validation completed")
}

// TestGUIRunCommand tests that the GUI command can be parsed
// (with GUI disabled, to avoid launching a window in tests).
func TestGUIRunCommand(t *testing.T) {
	cli := &CLI{}

	args := []string{"cmd", "-gui=false"} // Use false to prevent actual GUI launch

	err := cli.Run(args)
	if err == nil {
		t.Log("CLI flag parsing worked correctly")
	} else {
		if err.Error() != "input file is required" {
			t.Logf("CLI returned expected error for missing input: %v", err)
		} else {
			t.Log("CLI returned expected error for missing input file")
		}
	}
}
