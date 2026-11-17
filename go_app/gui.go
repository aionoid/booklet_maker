package main

import (
	"context"
	"fmt"
	"math"
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

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// appVersion is the version shown in the GUI footer.
const appVersion = "v1.1.0"

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

	// left panel: tab bar + switchable content (preview or info)
	leftContent  *fyne.Container // Stack holding exactly the active content, which fills it
	previewCont  *fyne.Container
	previewBox   *fyne.Container
	previewTitle *widget.Label

	infoCont  *fyne.Container
	infoRows  *fyne.Container // VBox of one Label per property line
	infoTitle *widget.Label

	statusLabel  *widget.Label
	versionLabel *widget.Label
	inputPath    string
}

// Run starts the GUI application.
func (gui *GUI) Run() {
	gui.app = app.NewWithID("booklet-maker")
	gui.window = gui.app.NewWindow("Booklet Maker " + appVersion)
	gui.window.Resize(fyne.NewSize(1200, 780))

	// --- form fields --------------------------------------------------------
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

	// --- input file picker --------------------------------------------------
	filePickerButton := widget.NewButton("Browse PDF", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				path := strings.TrimPrefix(reader.URI().String(), "file://")
				gui.inputPath = path
				gui.inputEntry.SetText(path)
				gui.renderPreview()
				gui.refreshInfo()
			}
		}, gui.window)
		fd.SetFilter(&storage.ExtensionFileFilter{Extensions: []string{".pdf"}})
		fd.Show()
	})

	// --- action buttons ------------------------------------------------------
	processBtn := widget.NewButton("Create Booklet", func() { gui.processBooklet() })
	pnumBtn := widget.NewButton("Add Page Numbers", func() { gui.addPageNumbers() })
	volBtn := widget.NewButton("Split Volumes", func() { gui.splitVolumes() })
	resetBtn := widget.NewButton("Reset", func() { gui.reset() })

	gui.statusLabel = widget.NewLabel("Ready")
	gui.versionLabel = widget.NewLabel("Booklet Maker " + appVersion)

	// --- preview/info panel (right side) --------------------------------------
	tabPreviewBtn := widget.NewButton("Preview", func() { gui.showTab("preview") })
	tabInfoBtn := widget.NewButton("Info", func() { gui.showTab("info") })
	tabBar := container.NewHBox(tabPreviewBtn, tabInfoBtn)

	// preview content: title on top, scroll filling the rest
	gui.previewTitle = widget.NewLabel("PDF preview — select a file to begin")
	gui.previewBox = container.NewVBox()
	previewScroll := container.NewVScroll(gui.previewBox)
	gui.previewCont = container.NewBorder(
		gui.previewTitle, nil, nil, nil, previewScroll,
	)

	// info content: title on top, scroll (per-line labels) filling the rest
	gui.infoTitle = widget.NewLabel("PDF information")
	gui.infoRows = container.NewVBox()
	infoScroll := container.NewVScroll(gui.infoRows)
	gui.infoCont = container.NewBorder(
		gui.infoTitle, nil, nil, nil, infoScroll,
	)

	// The preview/info column's middle is a Stack: whichever tab is active fills it.
	gui.leftContent = container.NewStack()
	gui.leftContent.Add(gui.previewCont) // start on Preview
	previewPanel := container.NewBorder(tabBar, nil, nil, nil, gui.leftContent)

	// --- form panel (left side): manual rows so labels are left-aligned --------
	buttonsRow := container.NewHBox(processBtn, pnumBtn, volBtn, resetBtn)
	formRows := container.NewVBox(
		gui.formRow("Input File", container.NewBorder(nil, nil, nil, filePickerButton, gui.inputEntry)),
		gui.formRow("Output File", gui.outputEntry),
		gui.formRow("Pages Per Sheet", gui.pagesSelect),
		gui.formRow("Reading Direction", gui.directionSelect),
		gui.formRow("Sections", gui.sectionsEntry),
		gui.formRow("Add Blank Pages", gui.blankSelect),
		gui.formRow("Page Numbers", gui.pageNumCheck),
		gui.formRow("Split Into Volumes (N)", gui.volumesEntry),
	)
	formPanel := container.NewBorder(nil, buttonsRow, nil, nil, formRows)

	// --- main area: form fills the left, preview/info is a fixed column on the right
	mainArea := container.NewBorder(nil, nil, nil, previewPanel, formPanel)

	// --- footer (status left, version right) ---------------------------------
	footer := container.NewHBox(gui.statusLabel, layout.NewSpacer(), gui.versionLabel)

	// --- assemble: title top, main area fills the middle, footer bottom ------
	content := container.NewBorder(
		widget.NewRichTextFromMarkdown("# Booklet Maker"), footer, nil, nil, mainArea,
	)
	gui.window.SetContent(content)
	gui.showTab("preview")
	gui.refreshInfo()
	gui.window.ShowAndRun()
}

// showTab swaps the left panel content between the preview and info tabs.
func (gui *GUI) showTab(name string) {
	fyne.Do(func() {
		gui.leftContent.RemoveAll()
		switch name {
		case "info":
			gui.leftContent.Add(gui.infoCont)
		default:
			gui.leftContent.Add(gui.previewCont)
		}
	})
}

// formRow builds one [left-aligned label | input] row. The label sits at the left
// edge at its natural width; the input fills the remaining width of the row.
func (gui *GUI) formRow(label string, input fyne.CanvasObject) *fyne.Container {
	lb := widget.NewLabel(label)
	return container.NewBorder(nil, nil, lb, nil, input)
}

// renderPreview renders up to 30 pages of the input PDF as thumbnails using
// pdftoppm (poppler-utils) and displays them in the preview tab.
func (gui *GUI) renderPreview() {
	if gui.inputPath == "" {
		gui.previewBox.RemoveAll()
		gui.previewTitle.SetText("PDF preview — select a file to begin")
		return
	}
	gui.previewTitle.SetText("Rendering preview...")
	go func() {
		var pngs []string
		var images []fyne.CanvasObject
		var errMsg string
		tmpDir, err := os.MkdirTemp("", "bookit_preview_*")
		if err != nil {
			errMsg = "Preview unavailable: " + err.Error()
		} else {
			prefix := filepath.Join(tmpDir, "page")
			// ~48 DPI => A4 thumbnail ≈ 400 px wide, so the Preview/Info column is about 1/3 of the window width.
			cmd := exec.Command("pdftoppm", "-png", "-r", "48", "-f", "1", "-l", "30", gui.inputPath, prefix)
			if err := cmd.Run(); err != nil {
				errMsg = "Preview unavailable: " + err.Error()
			} else {
				pngs, _ = filepath.Glob(filepath.Join(tmpDir, "page*.png"))
				naturalSortFiles(pngs)
				for _, p := range pngs {
					img := canvas.NewImageFromFile(p)
					// ImageFillOriginal makes the layout allocate the image's true
					// pixel dimensions (the default Stretch leaves a 1x1 MinSize,
					// squishing each page into a thin line).
					img.FillMode = canvas.ImageFillOriginal
					images = append(images, img)
				}
			}
		}
		fyne.Do(func() {
			if errMsg != "" {
				gui.previewTitle.SetText(errMsg)
				return
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
		})
	}()
}

// refreshInfo shows the selected PDF's metadata in the Info tab, read through
// pdfcpu's Go API — no external binary required.
func (gui *GUI) refreshInfo() {
	if gui.inputPath == "" {
		fyne.Do(func() {
			gui.setInfoContent("**PDF info**\n\n*Select a file to see details.*")
		})
		return
	}
	go func() {
		ctx := context.Background()
		f, err := os.Open(gui.inputPath)
		var text string
		if err != nil {
			text = "Could not open file: " + err.Error()
		} else {
			defer f.Close()
			info, e := api.PDFInfo(ctx, f, gui.inputPath, nil, true, nil)
			if e != nil {
				text = "Could not read PDF info: " + e.Error()
			} else {
				text = buildPdfInfoText(info)
			}
		}
		fyne.Do(func() {
			gui.setInfoContent("**PDF info**\n\n" + text)
		})
	}()
}

// setInfoContent renders a string as one wrapping label per line so every property
// sits on its own row and long lines wrap instead of overflowing the window.
func (gui *GUI) setInfoContent(s string) {
	gui.infoRows.RemoveAll()
	for _, line := range strings.Split(s, "\n") {
		lb := widget.NewLabel(line)
		lb.Wrapping = fyne.TextWrapWord
		gui.infoRows.Add(lb)
	}
}

// buildPdfInfoText formats a PDFInfo struct into a human-readable block.
func buildPdfInfoText(info *pdfcpu.PDFInfo) string {
	lines := []string{
		"File: " + info.FileName,
		"PDF version: " + info.Version,
		fmt.Sprintf("Page count: %d", info.PageCount),
	}
	if len(info.Dimensions) > 0 {
		d := info.Dimensions[0]
		lines = append(lines, fmt.Sprintf("Page size: %.0f x %.0f pt (%s)", d.Width, d.Height, paperName(d.Width, d.Height)))
	}
	if info.Title != "" {
		lines = append(lines, "Title: "+info.Title)
	}
	if info.Author != "" {
		lines = append(lines, "Author: "+info.Author)
	}
	if info.Subject != "" {
		lines = append(lines, "Subject: "+info.Subject)
	}
	if info.Creator != "" {
		lines = append(lines, "Creator: "+info.Creator)
	}
	if info.Producer != "" {
		lines = append(lines, "Producer: "+info.Producer)
	}
	if info.CreationDate != "" {
		lines = append(lines, "Created: "+info.CreationDate)
	}
	if info.ModificationDate != "" {
		lines = append(lines, "Modified: "+info.ModificationDate)
	}
	if info.Encrypted {
		lines = append(lines, "Encrypted: yes")
	} else {
		lines = append(lines, "Encrypted: no")
	}
	if info.Linearized {
		lines = append(lines, "Linearized: yes")
	}
	if info.Tagged {
		lines = append(lines, "Tagged (accessible): yes")
	}
	if info.Outlines {
		lines = append(lines, "Bookmarks: yes")
	}
	if len(info.Fonts) > 0 {
		var sb strings.Builder
		for _, ft := range info.Fonts {
			name := ft.Name
			if name == "" {
				name = ft.Prefix
			}
			if ft.Embedded {
				name += " (embedded)"
			}
			sb.WriteString(name + ", ")
		}
		lines = append(lines, "Fonts: "+strings.TrimRight(sb.String(), ", "))
	}
	return strings.Join(lines, "\n")
}

// paperName guesses a common paper name from width/height in points.
func paperName(w, h float64) string {
	da4 := math.Abs(w-595.28) + math.Abs(h-841.89)
	dlet := math.Abs(w-612.0) + math.Abs(h-792.0)
	if da4 <= dlet {
		return "A4"
	}
	if dlet < 30 {
		return "Letter"
	}
	return "custom"
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
		fyne.Do(func() {
			if err != nil {
				gui.statusLabel.SetText("Error: " + err.Error())
			} else {
				gui.statusLabel.SetText("Booklet created! See output, sections/ and print_ready/ folders.")
			}
		})
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
		fyne.Do(func() {
			if err != nil {
				gui.statusLabel.SetText("Error: " + err.Error())
			} else {
				gui.statusLabel.SetText("Page numbers added to " + outFile)
			}
		})
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
		fyne.Do(func() {
			if err != nil {
				gui.statusLabel.SetText("Error: " + err.Error())
			} else {
				gui.statusLabel.SetText(fmt.Sprintf("Split into %d volume(s) in %s", n, filepath.Base(volDir)))
			}
		})
	}()
}

// reset clears the form and preview, returning to the Preview tab.
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
	gui.showTab("preview")
	gui.refreshInfo()
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
