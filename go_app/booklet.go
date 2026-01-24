package main

import (
	"fmt"
	"math"
)

// BookletConfig holds the configuration for booklet creation
type BookletConfig struct {
	InputFile        string
	OutputFile       string
	PagesPerSheet    int
	ReadingDirection string // "RTL" or "LTR"
	Sections         int
	AddBlank         int // 0 or 1
}

// ProcessBooklet processes a PDF file to create a booklet
func ProcessBooklet(config *BookletConfig) error {
	fmt.Printf("Processing booklet: %s -> %s\n", config.InputFile, config.OutputFile)
	fmt.Printf("Config: pagesPerSheet=%d, direction=%s, sections=%d, addBlank=%d\n", 
		config.PagesPerSheet, config.ReadingDirection, config.Sections, config.AddBlank)

	// Step 1: Prepare the PDF with blank pages if needed
	tempFile := config.OutputFile + ".tmp"
	err := prepareBookletPages(config.InputFile, tempFile, config.AddBlank, config.Sections, config.PagesPerSheet)
	if err != nil {
		return fmt.Errorf("failed to prepare booklet pages: %w", err)
	}

	// Step 2: Handle reading direction by reversing pages if RTL
	reversedFile := tempFile + ".rev"
	if config.ReadingDirection == "RTL" {
		err = handleReadingDirection(tempFile, reversedFile)
		if err != nil {
			return fmt.Errorf("failed to handle reading direction: %w", err)
		}
	} else {
		reversedFile = tempFile // If LTR, no reversal needed
	}

	// Step 3: Create the actual booklet layout
	err = createBooklet(reversedFile, config.OutputFile, config.PagesPerSheet)
	if err != nil {
		return fmt.Errorf("failed to create booklet: %w", err)
	}

	// Step 4: Add stations (sewing points) to the booklet
	err = addStations(config.OutputFile, config.PagesPerSheet)
	if err != nil {
		return fmt.Errorf("failed to add stations: %w", err)
	}

	// Step 5: Add section marking to the booklet
	err = addSectionMarking(config.OutputFile, config.Sections, config.PagesPerSheet)
	if err != nil {
		return fmt.Errorf("failed to add section marking: %w", err)
	}

	// Clean up temporary files
	_ = removeTempFiles(tempFile, reversedFile)

	return nil
}

// prepareBookletPages prepares the PDF with blank pages for proper booklet formatting
func prepareBookletPages(inputFile, outputFile string, addBlank, nsections, pagesPerSheet int) error {
	if addBlank == 0 {
		// Just copy the file if no blank pages needed
		// In a real implementation, this would copy the file
		fmt.Printf("Copying %s to %s\n", inputFile, outputFile)
		return nil
	}

	// Determine multiplier based on pages per sheet
	var folioMultiplier int
	switch pagesPerSheet {
	case 1, 2:
		folioMultiplier = 2
	case 4, 8:
		folioMultiplier = 4
	default:
		folioMultiplier = 2
	}

	pagesPerSignature := nsections * folioMultiplier

	// In a real implementation, this would process the PDF
	fmt.Printf("Preparing booklet pages: %s -> %s, pagesPerSignature: %d\n", inputFile, outputFile, pagesPerSignature)

	return nil
}

// handleReadingDirection reverses pages for RTL reading direction
func handleReadingDirection(inputFile, outputFile string) error {
	fmt.Printf("Reversing pages for RTL: %s -> %s\n", inputFile, outputFile)
	// In a real implementation, this would reverse the pages
	return nil
}

// createBooklet creates the actual booklet layout
func createBooklet(inputFile, outputFile string, pagesPerSheet int) error {
	fmt.Printf("Creating booklet layout: %s -> %s, pagesPerSheet: %d\n", inputFile, outputFile, pagesPerSheet)
	// In a real implementation, this would create the booklet layout
	return nil
}

// addStations adds sewing points/stations to the PDF
func addStations(pdfFile string, pagesPerSheet int) error {
	// Define station configurations based on the x-up format
	var stationsConfig []float64
	switch pagesPerSheet {
	case 1: // A5 (1-up) - 8 points
		stationsConfig = []float64{7.0, 19.3, 31.6, 43.9, 56.1, 68.4, 80.7, 93.0}
	case 2: // A6 (2-up) - 6 points
		stationsConfig = []float64{8.0, 24.8, 41.6, 58.4, 75.2, 92.0}
	case 4, 8: // A7 (4-up) or 8-up - 4 points
		stationsConfig = []float64{10.0, 36.6, 63.3, 90.0}
	default: // Default to 4 points
		stationsConfig = []float64{10.0, 36.6, 63.3, 90.0}
	}

	fmt.Printf("Adding stations to %s, configuration: %v\n", pdfFile, stationsConfig)

	// In a real implementation, this would add the stations to the PDF
	return nil
}

// addSectionMarking adds section marking to the PDF
func addSectionMarking(pdfFile string, nsections, pagesPerSheet int) error {
	fmt.Printf("Adding section marking to %s, sections: %d, pagesPerSheet: %d\n", pdfFile, nsections, pagesPerSheet)

	// Calculate pages per section based on nsections
	foliosPerSection := nsections
	pagesPerSection := foliosPerSection * 2 // 2 pages per folio

	// Calculate how many sections we have (simulated)
	totalPages := 100 // Simulated total pages
	totalSections := int(math.Ceil(float64(totalPages) / float64(pagesPerSection)))

	// Apply section marking for each section
	for sectionNum := 1; sectionNum <= totalSections; sectionNum++ {
		// Calculate the x position for this section (positive values: 10, 25, 40, etc.)
		sectionPosition := 10 + (sectionNum-1)*15

		// Calculate page range for this section
		startPage := (sectionNum-1)*pagesPerSection + 1
		endPage := sectionNum * pagesPerSection
		if endPage > totalPages {
			endPage = totalPages
		}

		// For each even page in the section, apply the folio marking
		folioNum := 1
		for page := startPage; page <= endPage && folioNum <= foliosPerSection; page++ {
			// Only apply to even pages
			if page%2 == 0 {
				// Format the folio number with leading zero
				folioNumberStr := fmt.Sprintf("%02d", folioNum)
				fmt.Printf("  Adding folio marking %s to page %d at position %d\n", folioNumberStr, page, sectionPosition)
				folioNum++
			}
		}
	}

	return nil
}

// removeTempFiles removes temporary files
func removeTempFiles(files ...string) error {
	for _, file := range files {
		if file != "" {
			// In a real implementation, this would remove the file
			fmt.Printf("Removing temp file: %s\n", file)
		}
	}
	return nil
}