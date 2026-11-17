package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
)

// CLI handles command-line interface
type CLI struct{}

// Run executes the CLI application
func (cli *CLI) Run(args []string) error {
	var (
		inputFile        string
		outputFile       string = "booklet.pdf"
		pagesPerSheet    int    = 1
		readingDirection string = "RTL"
		sections         int    = 8
		addBlank         int    = 1
		pageNumbers      bool   = false
		volumes          int    = 0
		runGui           bool   = false
	)

	cliFlags := flag.NewFlagSet("", flag.ContinueOnError) // ContinueOnError to handle errors ourselves
	cliFlags.StringVar(&inputFile, "input", "", "Input PDF file (required)")
	cliFlags.StringVar(&inputFile, "i", "", "Input PDF file (shorthand)")
	cliFlags.StringVar(&outputFile, "output", "booklet.pdf", "Output PDF file")
	cliFlags.StringVar(&outputFile, "o", "booklet.pdf", "Output PDF file (shorthand)")
	cliFlags.IntVar(&pagesPerSheet, "pages", 1, "Pages per sheet (1, 2, 4, or 8)")
	cliFlags.IntVar(&pagesPerSheet, "p", 1, "Pages per sheet (shorthand)")
	cliFlags.StringVar(&readingDirection, "direction", "RTL", "Reading direction (RTL or LTR)")
	cliFlags.StringVar(&readingDirection, "d", "RTL", "Reading direction (shorthand)")
	cliFlags.IntVar(&sections, "sections", 8, "Number of sections")
	cliFlags.IntVar(&sections, "s", 8, "Number of sections (shorthand)")
	cliFlags.IntVar(&addBlank, "blank", 1, "Add blank pages (0 or 1)")
	cliFlags.IntVar(&addBlank, "b", 1, "Add blank pages (shorthand)")
	cliFlags.BoolVar(&pageNumbers, "page-numbers", false, "Stamp page numbers on content pages (cover left blank)")
	cliFlags.BoolVar(&pageNumbers, "pnum", false, "Stamp page numbers (shorthand)")
	cliFlags.IntVar(&volumes, "volumes", 0, "Split the PDF into N volumes with covers (0 = off)")
	cliFlags.IntVar(&volumes, "vol", 0, "Split into N volumes (shorthand)")
	cliFlags.BoolVar(&runGui, "gui", false, "Run in GUI mode")
	cliFlags.BoolVar(&runGui, "g", false, "Run in GUI mode (shorthand)")

	err := cliFlags.Parse(args[1:])
	if err != nil {
		return err
	}

	// Check if GUI mode is requested.
	if runGui {
		gui := &GUI{}
		gui.Run()
		return nil // Return after GUI is closed
	}

	// Validate required input file.
	if inputFile == "" {
		cli.printUsage()
		return fmt.Errorf("input file is required")
	}

	ctx := context.Background()

	// Alternative modes: volume splitting and page numbering operate on the raw
	// input PDF (independent of the booklet pipeline).
	if volumes > 0 {
		volDir := filepath.Join(filepath.Dir(inputFile), "volumes")
		_, err := SplitIntoVolumes(ctx, inputFile, volDir, volumes)
		return err
	}
	if pageNumbers {
		return AddPageNumbers(ctx, inputFile, outputFile)
	}

	// Booklet mode: validate the layout options.
	if pagesPerSheet != 1 && pagesPerSheet != 2 && pagesPerSheet != 4 && pagesPerSheet != 8 {
		return fmt.Errorf("pages per sheet must be 1, 2, 4, or 8, got %d", pagesPerSheet)
	}
	if readingDirection != "RTL" && readingDirection != "LTR" {
		return fmt.Errorf("reading direction must be RTL or LTR, got %s", readingDirection)
	}
	if addBlank != 0 && addBlank != 1 {
		return fmt.Errorf("add blank must be 0 or 1, got %d", addBlank)
	}

	config := &BookletConfig{
		InputFile:        inputFile,
		OutputFile:       outputFile,
		PagesPerSheet:    pagesPerSheet,
		ReadingDirection: readingDirection,
		Sections:         sections,
		AddBlank:         addBlank,
	}

	return ProcessBooklet(config)
}

// printUsage prints the usage information
func (cli *CLI) printUsage() {
	fmt.Println("Usage: booklet-maker -input <input.pdf> [OPTIONS]")
	fmt.Println("Options:")
	fmt.Println("  -input, -i          Input PDF file (required)")
	fmt.Println("  -output, -o         Output PDF file (default: booklet.pdf)")
	fmt.Println("  -pages, -p          Pages per sheet (1, 2, 4, or 8) (default: 1)")
	fmt.Println("  -direction, -d      Reading direction (RTL or LTR) (default: RTL)")
	fmt.Println("  -sections, -s       Number of sections (default: 8)")
	fmt.Println("  -blank, -b          Add blank pages (0 or 1) (default: 1)")
	fmt.Println("  -page-numbers, -pnum   Stamp page numbers on content pages (cover left blank)")
	fmt.Println("  -volumes, -vol N    Split the PDF into N volumes with covers")
	fmt.Println("  -gui, -g            Run in GUI mode")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  booklet-maker -input mybook.pdf")
	fmt.Println("  booklet-maker -i mybook.pdf -o output.pdf -p 2 -d LTR")
	fmt.Println("  booklet-maker -input book.pdf -pages 4 -sections 6 -blank 0")
	fmt.Println("  booklet-maker -i book.pdf --page-numbers -o numbered.pdf")
	fmt.Println("  booklet-maker -i book.pdf -vol 3")
	fmt.Println("  booklet-maker -gui")
}
