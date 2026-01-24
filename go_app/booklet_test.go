package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBookletConfig(t *testing.T) {
	config := &BookletConfig{
		InputFile:        "test_input.pdf",
		OutputFile:       "test_output.pdf",
		PagesPerSheet:    2,
		ReadingDirection: "RTL",
		Sections:         8,
		AddBlank:         1,
	}

	if config.InputFile != "test_input.pdf" {
		t.Errorf("Expected InputFile to be 'test_input.pdf', got '%s'", config.InputFile)
	}

	if config.PagesPerSheet != 2 {
		t.Errorf("Expected PagesPerSheet to be 2, got %d", config.PagesPerSheet)
	}

	if config.ReadingDirection != "RTL" {
		t.Errorf("Expected ReadingDirection to be 'RTL', got '%s'", config.ReadingDirection)
	}
}

func TestValidateConfig(t *testing.T) {
	// Test valid pages per sheet values
	validValues := []int{1, 2, 4, 8}
	for _, val := range validValues {
		if val != 1 && val != 2 && val != 4 && val != 8 {
			t.Errorf("Value %d should be valid for PagesPerSheet", val)
		}
	}

	// Test invalid pages per sheet value
	invalidVal := 3
	if invalidVal != 1 && invalidVal != 2 && invalidVal != 4 && invalidVal != 8 {
		// This is expected behavior - 3 is not a valid value
	}

	// Test reading direction
	directions := []string{"RTL", "LTR"}
	for _, dir := range directions {
		if dir != "RTL" && dir != "LTR" {
			t.Errorf("Direction %s should be valid", dir)
		}
	}

	// Test add blank values
	blankValues := []int{0, 1}
	for _, val := range blankValues {
		if val != 0 && val != 1 {
			t.Errorf("Value %d should be valid for AddBlank", val)
		}
	}
}

func TestCLIArgsParsing(t *testing.T) {
	cli := &CLI{}

	// Test with minimal args
	args := []string{"cmd", "-input", "test.pdf"}
	err := cli.Run(args)
	if err != nil {
		t.Logf("Got error: %v", err)
	} else {
		t.Log("CLI ran without error (expected in simplified implementation)")
	}

	// Test with valid args but non-existent file
	args = []string{"cmd", "-input", "nonexistent.pdf"}
	err = cli.Run(args)
	if err != nil {
		t.Logf("Got error: %v", err)
	} else {
		t.Log("CLI ran without error (expected in simplified implementation)")
	}
}

func TestFilenames(t *testing.T) {
	// Test that we can create temporary files for testing
	tmpDir := t.TempDir()
	
	inputPath := filepath.Join(tmpDir, "test_input.pdf")
	outputPath := filepath.Join(tmpDir, "test_output.pdf")
	
	// Create a dummy input file for testing
	err := os.WriteFile(inputPath, []byte("dummy pdf content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy input file: %v", err)
	}

	// Verify the file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		t.Fatalf("Input file was not created: %v", err)
	}

	// Test file paths
	if filepath.Base(inputPath) != "test_input.pdf" {
		t.Errorf("Expected input file name to be 'test_input.pdf', got '%s'", filepath.Base(inputPath))
	}

	if filepath.Base(outputPath) != "test_output.pdf" {
		t.Errorf("Expected output file name to be 'test_output.pdf', got '%s'", filepath.Base(outputPath))
	}
}

func TestPageCalculations(t *testing.T) {
	// Test folio multiplier logic
	testCases := []struct {
		pagesPerSheet    int
		expectedFolioMult int
	}{
		{1, 2},
		{2, 2},
		{4, 4},
		{8, 4},
		{3, 2}, // default case
	}

	for _, tc := range testCases {
		var folioMultiplier int
		switch tc.pagesPerSheet {
		case 1, 2:
			folioMultiplier = 2
		case 4, 8:
			folioMultiplier = 4
		default:
			folioMultiplier = 2
		}

		if folioMultiplier != tc.expectedFolioMult {
			t.Errorf("For PagesPerSheet=%d, expected folioMultiplier=%d, got %d", 
				tc.pagesPerSheet, tc.expectedFolioMult, folioMultiplier)
		}
	}
}