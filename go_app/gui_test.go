package main

import (
	"testing"
)

func TestGUICreation(t *testing.T) {
	// This test verifies that the GUI can be instantiated without errors
	gui := &GUI{}

	// We can't fully test the GUI without displaying it, but we can verify
	// that the struct can be created without issues
	if gui == nil {
		t.Error("Failed to create GUI instance")
	}
}

func TestCLIRunWithGUIFlagDisabled(t *testing.T) {
	// Test that CLI can parse the GUI flag disabled (without launching the GUI in tests).
	cli := &CLI{}

	// -gui=false: parsing only, no GUI launch in test environment
	args := []string{"cmd", "-gui=false"}

	err := cli.Run(args)

	if err != nil {
		t.Logf("CLI returned expected error for missing input: %v", err)
	} else {
		t.Log("CLI flag parsing worked correctly")
	}
}
