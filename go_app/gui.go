package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// GUI handles the graphical user interface
type GUI struct {
	app    fyne.App
	window fyne.Window
}

// Run starts the GUI application
func (gui *GUI) Run() {
	gui.app = app.New()
	gui.window = gui.app.NewWindow("Booklet Maker")
	gui.window.Resize(fyne.NewSize(600, 500))

	// Create form fields
	inputFileEntry := widget.NewEntry()
	inputFileEntry.SetPlaceHolder("Select input PDF file...")

	outputFileEntry := widget.NewEntry()
	outputFileEntry.SetText("booklet.pdf")
	outputFileEntry.Validator = validation.NewRegexp(`^[^<>:"/\\|?*]+$`, "Invalid filename")

	pagesPerSheetSelector := widget.NewSelect([]string{"1", "2", "4", "8"}, nil)
	pagesPerSheetSelector.Selected = "1"

	directionSelector := widget.NewSelect([]string{"RTL", "LTR"}, nil)
	directionSelector.Selected = "RTL"

	sectionsEntry := widget.NewEntry()
	sectionsEntry.SetText("8")
	sectionsEntry.Validator = validation.NewRegexp(`^\d+$`, "Must be a number")

	blankSelector := widget.NewSelect([]string{"0", "1"}, nil)
	blankSelector.Selected = "1"

	// Create file picker button for input file
	filePickerButton := widget.NewButton("Browse PDF File", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				inputFileEntry.SetText(reader.URI().String()[7:]) // Remove "file://" prefix
			}
		}, gui.window)
		fd.SetFilter(&storage.ExtensionFileFilter{Extensions: []string{".pdf"}})
		fd.Show()
	})

	// Create a progress label
	progressLabel := widget.NewLabel("Ready to create booklet")
	progressLabel.Wrapping = fyne.TextWrapWord

	// Create the submit button
	submitButton := widget.NewButton("Create Booklet", func() {
		// Validate inputs
		if inputFileEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("please select an input PDF file"), gui.window)
			return
		}

		// Parse the form values
		pagesPerSheet := 1
		if val := pagesPerSheetSelector.Selected; val != "" {
			fmt.Sscanf(val, "%d", &pagesPerSheet)
		}

		readingDirection := directionSelector.Selected
		if readingDirection == "" {
			readingDirection = "RTL"
		}

		sections := 8
		if val := sectionsEntry.Text; val != "" {
			fmt.Sscanf(val, "%d", &sections)
		}

		addBlank := 1
		if val := blankSelector.Selected; val != "" {
			fmt.Sscanf(val, "%d", &addBlank)
		}

		outputFile := outputFileEntry.Text
		if outputFile == "" {
			outputFile = "booklet.pdf"
		}

		// Create the configuration
		config := &BookletConfig{
			InputFile:        inputFileEntry.Text,
			OutputFile:       outputFile,
			PagesPerSheet:    pagesPerSheet,
			ReadingDirection: readingDirection,
			Sections:         sections,
			AddBlank:         addBlank,
		}

		// Validate the configuration
		if err := validateConfig(config); err != nil {
			dialog.ShowError(err, gui.window)
			return
		}

		// Update progress
		progressLabel.SetText("Processing booklet...")

		// Process the booklet in a goroutine to prevent UI freezing
		go func() {
			err := ProcessBooklet(config)
			if err != nil {
				dialog.ShowError(err, gui.window)
				progressLabel.SetText("Error: " + err.Error())
			} else {
				progressLabel.SetText("Booklet created successfully! Output: " + config.OutputFile)
			}
		}()
	})

	// Create the form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Input File", Widget: container.NewBorder(nil, nil, inputFileEntry, filePickerButton)},
			{Text: "Output File", Widget: outputFileEntry},
			{Text: "Pages Per Sheet", Widget: pagesPerSheetSelector},
			{Text: "Reading Direction", Widget: directionSelector},
			{Text: "Sections", Widget: sectionsEntry},
			{Text: "Add Blank Pages", Widget: blankSelector},
		},
		OnSubmit: func() {
			submitButton.OnTapped() // Trigger the same action as the submit button
		},
	}

	// Create the main content container
	content := container.NewVBox(
		widget.NewRichTextFromMarkdown("# Booklet Maker"),
		widget.NewLabel("Fill in the options below to create your booklet:"),
		layout.NewSpacer(),
		form,
		layout.NewSpacer(),
		progressLabel,
		submitButton,
	)

	gui.window.SetContent(content)
	gui.window.ShowAndRun()
}

// validateConfig validates the booklet configuration
func validateConfig(config *BookletConfig) error {
	if config.InputFile == "" {
		return fmt.Errorf("input file is required")
	}

	if config.PagesPerSheet != 1 && config.PagesPerSheet != 2 && config.PagesPerSheet != 4 && config.PagesPerSheet != 8 {
		return fmt.Errorf("pages per sheet must be 1, 2, 4, or 8, got %d", config.PagesPerSheet)
	}

	if config.ReadingDirection != "RTL" && config.ReadingDirection != "LTR" {
		return fmt.Errorf("reading direction must be RTL or LTR, got %s", config.ReadingDirection)
	}

	if config.AddBlank != 0 && config.AddBlank != 1 {
		return fmt.Errorf("add blank must be 0 or 1, got %d", config.AddBlank)
	}

	return nil
}
