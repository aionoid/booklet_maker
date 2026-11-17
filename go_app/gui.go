package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// GUI handles the graphical user interface.
type GUI struct {
	app    fyne.App
	window fyne.Window

	// form widgets (kept so we can read values / reset them)
	inputEntry      *widget.Entry
	outputEntry     *widget.Entry
	pagesSelect     *widget.Select
	directionSelect *widget.Select
	sectionsEntry   *widget.Entry
	blankSelect     *widget.Select
	pageNumCheck    *widget.Check
	volumesEntry    *widget.Entry

	// preview panel widgets
	previewTitle *widget.Label
	previewBox   *fyne.Container // VBox of page thumbnails

	statusLabel *widget.Label
	inputPath   string
}

// Run starts the GUI application.
func (gui *GUI) Run() {
	gui.app = app.New()
	gui.window = gui.app.NewWindow("Booklet Maker")
	gui.window.Resize(fyne.NewSize(1000, 750))

	// --- form fields ------------------------------------------------------
	gui.inputEntry = widget.NewEntry()
	gui.inputEntry.SetPlaceHolder("Input PDF file...")

	gui.outputEntry = widget.NewEntry()
	gui.outputEntry.SetText("booklet.pdf")
	gui.outputEntry.Validator = validation.NewRegexp(`^[^<>:"/\\|?*]+$`, "Invalid filename")

	gui.pagesSelect = widget.NewSelect([]string{"1", "2", "4", "8"}, nil)
	gui.pagesSelect.Selected = "1"

	gui.directionSelect = widget.NewSelect([]string{"RTL", "LTR"}, nil)
	gui.directionSelect.Selected = "RTL"

	gui.sectionsEntry = widget.NewEntry()
	gui.sectionsEntry.SetText("8")
	gui.sectionsEntry.Validator = validation.NewRegexp(`^\d+$`, "Must be a number")

	gui.blankSelect = widget.NewSelect([]string{"0", "1"}, nil)
	gui.blankSelect.Selected = "1"

	gui.pageNumCheck = widget.NewCheck("Add page numbers (cover left blank)", func(b bool) {})
	gui.pageNumCheck.SetChecked(false)

	gui.volumesEntry = widget.NewEntry()
	gui.volumesEntry.SetText("0")
	gui.volumesEntry.Validator = validation.NewRegexp(`^\d+$`, "Number of volumes (0 = off)")

	// --- input file picker -------------------------------------------------
	filePickerButton := widget.NewButton("Browse PDF", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				path := strings.TrimPrefix(reader.URI().String(), "file://")
				gui.inputPath = path
				gui.inputEntry.SetText(path)
				gui.renderPreview()
			}
		}, gui.window)
		fd.SetFilter(&storage.ExtensionFileFilter{Extensions: []string{"pdf"}})
		fd.Show()
	})

	// --- action buttons ----------------------------------------------------
	processBtn := widget.NewButton("Create Booklet", func() { gui.processBooklet() })
	pnumBtn := widget.NewButton("Add Page Numbers", func() { gui.addPageNumbers() })
	volBtn := widget.NewButton("Split Volumes", func() { gui.splitVolumes() })
	resetBtn := widget.NewButton("Reset", func() { gui.reset() })

	gui.statusLabel = widget.NewLabel("Ready")
	gui.statusLabel.Wrapping = fyne.TextWrapWord

	// --- left panel: PDF preview ------------------------------------------
	gui.previewTitle = widget.NewLabel("PDF preview — select a file to begin")
	gui.previewBox = container.NewVBox()
	previewScroll := container.NewVScroll(gui.previewBox)
	leftPanel := container.NewVBox(gui.previewTitle, previewScroll)

	// --- right panel: options form -----------------------------------------
	rightForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Input File", Widget: container.NewBorder(nil, nil, gui.inputEntry, filePickerButton)},
			{Text: "Output File", Widget: gui.outputEntry},
			{Text: "Pages Per Sheet", Widget: gui.pagesSelect},
			{Text: "Reading Direction", Widget: gui.directionSelect},
			{Text: "Sections", Widget: gui.sectionsEntry},
			{Text: "Add Blank Pages", Widget: gui.blankSelect},
			{Text: "Page Numbers", Widget: gui.pageNumCheck},
			{Text: "Split Into Volumes (N)", Widget: gui.volumesEntry},
		},
	}

	// buttons row
	buttonsRow := container.NewHBox(processBtn, pnumBtn, volBtn, resetBtn)

	rightPanel := container.NewVBox(rightForm, layout.NewSpacer(), buttonsRow)

	// --- assemble ----------------------------------------------------------
	content := container.NewVBox(
		widget.NewRichTextFromMarkdown("# Booklet Maker"),
		layout.NewSpacer(),
		container.NewHBox(leftPanel, rightPanel),
		layout.NewSpacer(),
		gui.statusLabel,
	)
	gui.window.SetContent(content)
	gui.window.ShowAndRun()
}

// buildConfig reads the current form values into a BookletConfig.
func (gui *GUI) buildConfig() *BookletConfig {
	pages := 1
	fmt.Sscanf(gui.pagesSelect.Selected, "%d", &pages)
	sections := 8
	fmt.Sscanf(gui.sectionsEntry.Text, "%d", &sections)
	addBlank := 1
	fmt.Sscanf(gui.blankSelect.Selected, "%d", &addBlank)
	return &BookletConfig{
		InputFile:        gui.inputPath,
		OutputFile:       gui.outputEntry.Text,
		PagesPerSheet:    pages,
		ReadingDirection: gui.directionSelect.Selected,
		Sections:         sections,
		AddBlank:         addBlank,
	}
}

// processBooklet runs the full booklet pipeline.
func (gui *GUI) processBooklet() {
	if gui.inputPath == "" {
		dialog.ShowError(fmt.Errorf("please select an input PDF file"), gui.window)
		return
	}
	config := gui.buildConfig()
	if err := validateConfig(config); err != nil {
		dialog.ShowError(err, gui.window)
		return
	}
	gui.statusLabel.SetText("Creating booklet...")
	go func() {
		err := ProcessBooklet(config)
		if err != nil {
			gui.statusLabel.SetText("Error: " + err.Error())
		} else {
			gui.statusLabel.SetText("Booklet created! See output, sections/ and print_ready/ folders.")
		}
	}()
}

// addPageNumbers stamps page numbers on the selected input PDF (cover left blank).
func (gui *GUI) addPageNumbers() {
	if gui.inputPath == "" {
		dialog.ShowError(fmt.Errorf("please select an input PDF file"), gui.window)
		return
	}
	outFile := gui.outputEntry.Text
	if outFile == "" {
		outFile = "booklet_numbered.pdf"
	}
	gui.statusLabel.SetText("Adding page numbers...")
	go func() {
		ctx := context.Background()
		err := AddPageNumbers(ctx, gui.inputPath, outFile)
		if err != nil {
			gui.statusLabel.SetText("Error: " + err.Error())
		} else {
			gui.statusLabel.SetText("Page numbers added to " + outFile)
		}
	}()
}

// splitVolumes splits the selected input PDF into N volumes with covers.
func (gui *GUI) splitVolumes() {
	if gui.inputPath == "" {
		dialog.ShowError(fmt.Errorf("please select an input PDF file"), gui.window)
		return
	}
	n := 0
	fmt.Sscanf(gui.volumesEntry.Text, "%d", &n)
	if n <= 1 {
		dialog.ShowError(fmt.Errorf("enter a number of volumes greater than 1"), gui.window)
		return
	}
	gui.statusLabel.SetText("Splitting into volumes...")
	go func() {
		ctx := context.Background()
		volDir := filepath.Join(filepath.Dir(gui.inputPath), "volumes")
		_, err := SplitIntoVolumes(ctx, gui.inputPath, volDir, n)
		if err != nil {
			gui.statusLabel.SetText("Error: " + err.Error())
		} else {
			gui.statusLabel.SetText(fmt.Sprintf("Split into %d volume(s) in %s", n, filepath.Base(volDir)))
		}
	}()
}

// reset clears the form and preview.
func (gui *GUI) reset() {
	gui.inputEntry.SetText("")
	gui.outputEntry.SetText("booklet.pdf")
	gui.pagesSelect.Selected = "1"
	gui.directionSelect.Selected = "RTL"
	gui.sectionsEntry.SetText("8")
	gui.blankSelect.Selected = "1"
	gui.pageNumCheck.SetChecked(false)
	gui.volumesEntry.SetText("0")
	gui.inputPath = ""
	gui.previewBox.RemoveAll()
	gui.previewTitle.SetText("PDF preview — select a file to begin")
	gui.statusLabel.SetText("Ready")
}

// renderPreview renders up to 30 pages of the input PDF as thumbnails using
// pdftoppm (poppler-utils) and displays them in the left preview panel.
func (gui *GUI) renderPreview() {
	if gui.inputPath == "" {
		gui.previewBox.RemoveAll()
		gui.previewTitle.SetText("PDF preview — select a file to begin")
		return
	}
	gui.previewTitle.SetText("Rendering preview...")
	go func() {
		tmpDir, err := os.MkdirTemp("", "bookit_preview_*")
		if err != nil {
			gui.previewTitle.SetText("Preview unavailable: " + err.Error())
			return
		}
		prefix := filepath.Join(tmpDir, "page")
		cmd := exec.Command("pdftoppm", "-png", "-r", "40", "-f", "1", "-l", "30", gui.inputPath, prefix)
		if err := cmd.Run(); err != nil {
			gui.previewTitle.SetText("Preview unavailable: " + err.Error())
			return
		}
		pngs, _ := filepath.Glob(filepath.Join(tmpDir, "page*.png"))
		naturalSortFiles(pngs)
		var images []fyne.CanvasObject
		for _, p := range pngs {
			images = append(images, canvas.NewImageFromFile(p))
		}
		gui.previewBox.RemoveAll()
		for _, im := range images {
			gui.previewBox.Add(im)
		}
		if len(pngs) == 0 {
			gui.previewTitle.SetText("No pages to preview")
		} else {
			gui.previewTitle.SetText(fmt.Sprintf("Preview — %d page(s) shown (scroll to browse)", len(pngs)))
		}
	}()
}

// naturalSortFiles sorts file paths by their base name using natural (numeric) ordering.
func naturalSortFiles(files []string) {
	sort.Slice(files, func(i, j int) bool {
		return naturalLess(filepath.Base(files[i]), filepath.Base(files[j]))
	})
}

// validateConfig validates the booklet configuration.
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
