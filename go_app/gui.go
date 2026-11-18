package main

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/responses"
	"github.com/klippa-app/go-pdfium/webassembly"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// appVersion is the version shown in the GUI footer.
const appVersion = "v1.2.1"

// previewColWidth is the fixed width of the Preview/Info column (≈1/3 of the
// default 1200 px window). Every thumbnail is fitted to this width.
const previewColWidth = 400

// GUI handles the graphical user interface.
type GUI struct {
	app    fyne.App
	window fyne.Window

	// form widgets (kept so we can read values / reset them)
	inputEntry        *widget.Entry
	outputEntry       *widget.Entry
	pagesSelect       *widget.Select
	directionSelect   *widget.Select
	sectionsEntry     *widget.Entry
	blankSelect       *widget.Select
	pageNumCheck      *widget.Check
	splitVolumesCheck *widget.Check
	volumesEntry      *widget.Entry

	// left form panel: two tabs (create / prepare) + live "what you'll get" summary
	volumeEndEntries []*widget.Entry
	volumeEndsBox    *fyne.Container
	resultTitle      *widget.Label
	resultBox        *fyne.Container
	inputPageCount   int

	// left form panel: single "Create" form + live "what you'll get" summary
	createContent      *fyne.Container
	volumeOptionsBox   *fyne.Container
	volumeOptionsShown bool

	// left form tabs (Preview / Info) + the stack that holds the active one
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

	// PDFium preview state (page viewer with navigation).
	pdfPool         pdfium.Pool
	pdfInstance     pdfium.Pdfium
	pdfDoc          *responses.OpenDocument
	pdfPageCount    int
	pdfCurrentPage  int
	pdfGen          atomic.Int64
	pageEntry       *widget.Entry
	pageEntryLocked bool
}

// fixedWidthLayout always reserves a constant width for its content and fills
// it with the available size. This pins the Preview/Info column to exactly one
// third of the default window (1200 px / 3 = 400 px), independent of the PDF.
type fixedWidthLayout struct{ width float32 }

func (l *fixedWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(size)
	}
}

func (l *fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(l.width, 0)
}

// Run starts the GUI application.
func (gui *GUI) Run() {
	gui.app = app.NewWithID("booklet-maker")
	gui.window = gui.app.NewWindow("Booklet Maker " + appVersion)
	gui.window.Resize(fyne.NewSize(1200, 780))

	// --- form fields --------------------------------------------------------
	gui.volumeEndsBox = container.NewVBox() // created early so volumesEntry.OnChanged is safe

	gui.inputEntry = widget.NewEntry()
	gui.inputEntry.SetPlaceHolder("Input PDF file...")

	gui.outputEntry = widget.NewEntry()
	gui.outputEntry.SetText("output/booklet.pdf")
	gui.outputEntry.Validator = validation.NewRegexp(`^[^<>:"\\|?*]+$`, "Invalid output path")

	gui.pagesSelect = widget.NewSelect([]string{"1", "2", "4", "8"}, nil)
	gui.pagesSelect.Selected = "1"
	gui.pagesSelect.OnChanged = func(_ string) { gui.updateSummary() }

	gui.directionSelect = widget.NewSelect([]string{"RTL", "LTR"}, nil)
	gui.directionSelect.Selected = "RTL"

	gui.sectionsEntry = widget.NewEntry()
	gui.sectionsEntry.SetText("8")
	gui.sectionsEntry.Validator = validation.NewRegexp(`^\d+$`, "Must be a number")
	gui.sectionsEntry.OnChanged = func(_ string) { gui.updateSummary() }

	gui.blankSelect = widget.NewSelect([]string{"0", "1"}, nil)
	gui.blankSelect.Selected = "1"
	gui.blankSelect.OnChanged = func(_ string) { gui.updateSummary() }

	gui.pageNumCheck = widget.NewCheck("Add page numbering", func(b bool) {})
	gui.pageNumCheck.SetChecked(false)

	gui.splitVolumesCheck = widget.NewCheck("Split into volumes", func(b bool) {
		if b {
			n := 0
			fmt.Sscanf(gui.volumesEntry.Text, "%d", &n)
			if n < 2 {
				gui.volumesEntry.SetText("2")
			}
			gui.rebuildVolumeEnds()
			gui.setVolumeOptionsVisible(true)
		} else {
			gui.setVolumeOptionsVisible(false)
		}
	})

	gui.volumesEntry = widget.NewEntry()
	gui.volumesEntry.SetText("0")
	gui.volumesEntry.Validator = validation.NewRegexp(`^\d+$`, "Number of volumes (2 or more)")
	gui.volumesEntry.OnChanged = func(_ string) { gui.rebuildVolumeEnds() }

	// --- input file picker --------------------------------------------------
	filePickerButton := widget.NewButton("Browse PDF", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				path := strings.TrimPrefix(reader.URI().String(), "file://")
				gui.inputPath = path
				gui.inputEntry.SetText(path)
				gui.loadPdf(path)
				gui.refreshInfo()
			}
		}, gui.window)
		fd.SetFilter(&storage.ExtensionFileFilter{Extensions: []string{".pdf"}})
		fd.Show()
	})

	// --- action buttons ------------------------------------------------------
	processBtn := widget.NewButton("Create Booklet", func() { gui.processBooklet() })
	clearBtn := widget.NewButton("Clear Output", func() { gui.clearOutput() })
	resetBtn := widget.NewButton("Reset", func() { gui.reset() })

	gui.statusLabel = widget.NewLabel("Ready")
	gui.statusLabel.Wrapping = fyne.TextWrapWord
	gui.versionLabel = widget.NewLabel("Booklet Maker " + appVersion)

	// --- PDFium preview engine (WASM backend, no native library) -------------
	if err := gui.initPdfium(); err != nil {
		gui.statusLabel.SetText("PDF preview unavailable: " + err.Error())
	}
	gui.window.SetOnClosed(func() {
		gui.pdfInstance = nil
		if gui.pdfPool != nil {
			_ = gui.pdfPool.Close()
		}
	})

	// --- preview/info panel (right side) --------------------------------------
	tabPreviewBtn := widget.NewButton("Preview", func() { gui.showTab("preview") })
	tabInfoBtn := widget.NewButton("Info", func() { gui.showTab("info") })
	tabBar := container.NewHBox(tabPreviewBtn, tabInfoBtn)

	// preview content: title on top, a navigation row, then the rendered page in a
	// scrollable area (a full page is taller than the window).
	gui.previewTitle = widget.NewLabel("PDF preview — select a file to begin")
	gui.previewTitle.Wrapping = fyne.TextWrapWord

	gui.pageEntry = widget.NewEntry()
	gui.pageEntry.SetPlaceHolder("Page number")
	gui.pageEntry.OnChanged = gui.onPageEntryChanged

	prevBtn := widget.NewButton("‹", func() { gui.prevPage() })
	nextBtn := widget.NewButton("›", func() { gui.nextPage() })
	navRow := container.NewHBox(prevBtn, gui.pageEntry, nextBtn)

	gui.previewBox = container.NewVBox()
	previewScroll := container.NewVScroll(gui.previewBox)
	previewInner := container.NewBorder(navRow, nil, nil, nil, previewScroll)
	gui.previewCont = container.NewBorder(
		gui.previewTitle, nil, nil, nil, previewInner,
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
	// Pin the preview/info column to a fixed width (1/3 of the 1200 px window).
	previewCol := &fyne.Container{Layout: &fixedWidthLayout{previewColWidth}, Objects: []fyne.CanvasObject{previewPanel}}

	// --- left form panel: single "Create" form + live result ------------------
	gui.resultTitle = widget.NewLabel("What you'll get")
	gui.resultBox = container.NewVBox()

	// Volume options live inline and are only shown while "Split into volumes"
	// is checked.
	gui.volumeOptionsBox = container.NewVBox(
		gui.formRow("Number of volumes", gui.volumesEntry),
		gui.volumeEndsBox,
	)

	gui.createContent = container.NewVBox(
		gui.formRow("Input File", container.NewBorder(nil, nil, nil, filePickerButton, gui.inputEntry)),
		gui.formRow("Output File", gui.outputEntry),
		gui.formRow("Pages Per Sheet", gui.pagesSelect),
		gui.formRow("Reading Direction", gui.directionSelect),
		gui.formRow("Sections", gui.sectionsEntry),
		gui.formRow("Add Blank Pages", gui.blankSelect),
		gui.pageNumCheck,
		gui.splitVolumesCheck,
	)

	resultPanel := container.NewBorder(gui.resultTitle, nil, nil, nil, gui.resultBox)
	buttonsRow := container.NewHBox(processBtn, clearBtn, resetBtn)
	leftInner := container.NewVBox(gui.createContent, resultPanel)
	leftFormPanel := container.NewBorder(nil, buttonsRow, nil, nil, container.NewVScroll(leftInner))

	// --- main area: form fills the left, preview/info is a fixed 1/3 column on the right
	mainArea := container.NewBorder(nil, nil, nil, previewCol, leftFormPanel)

	// --- footer: status label fills the space up to the version on the right --
	footer := container.NewBorder(nil, nil, nil, gui.versionLabel, gui.statusLabel)

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

// labelColWidth is the fixed label column width in the form, so every input
// starts at the same x-position (aligned) with a consistent spacer after each label.
const labelColWidth = 230

// formRow builds one [left-aligned label | input] row. The label sits in a fixed
// width column (so all inputs align); the input fills the remaining width of the row.
func (gui *GUI) formRow(label string, input fyne.CanvasObject) *fyne.Container {
	lb := widget.NewLabel(label)
	lb.Wrapping = fyne.TextWrapWord
	labelCell := &fyne.Container{Layout: &fixedWidthLayout{labelColWidth}, Objects: []fyne.CanvasObject{lb}}
	return container.NewBorder(nil, nil, labelCell, nil, input)
}

// initPdfium starts the WASM-based PDFium engine used for the preview. It needs
// no native library. Called once at startup.
func (gui *GUI) initPdfium() error {
	pool, err := webassembly.Init(webassembly.Config{MinIdle: 1, MaxIdle: 1, MaxTotal: 2})
	if err != nil {
		return fmt.Errorf("init pdfium pool: %w", err)
	}
	gui.pdfPool = pool
	instance, err := pool.GetInstance(time.Second * 30)
	if err != nil {
		_ = pool.Close()
		gui.pdfPool = nil
		return fmt.Errorf("get pdfium instance: %w", err)
	}
	gui.pdfInstance = instance
	return nil
}

// loadPdf opens a new PDF in the preview engine and jumps to its first page.
func (gui *GUI) loadPdf(path string) {
	if gui.pdfInstance == nil {
		gui.setPreviewTitle("PDF preview unavailable")
		return
	}
	oldDoc := gui.pdfDoc
	gui.pdfDoc = nil
	gui.pdfPageCount = 0
	gui.pdfCurrentPage = 0
	if oldDoc != nil {
		go func() {
			_, _ = gui.pdfInstance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: oldDoc.Document})
		}()
	}
	go func() {
		data, err := os.ReadFile(path)
		if err != nil {
			fyne.Do(func() { gui.setPreviewTitle("Could not open PDF: " + err.Error()) })
			return
		}
		doc, err := gui.pdfInstance.OpenDocument(&requests.OpenDocument{File: &data})
		if err != nil {
			fyne.Do(func() { gui.setPreviewTitle("Could not open PDF: " + err.Error()) })
			return
		}
		pc, err := gui.pdfInstance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
		if err != nil {
			_, _ = gui.pdfInstance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})
			fyne.Do(func() { gui.setPreviewTitle("Could not read PDF: " + err.Error()) })
			return
		}
		fyne.Do(func() {
			gui.pdfDoc = doc
			gui.pdfPageCount = pc.PageCount
			gui.renderPage(0)
		})
	}()
}

// clearPdf releases the currently open preview document and resets the view to its
// empty state.
func (gui *GUI) clearPdf() {
	if gui.pdfInstance != nil && gui.pdfDoc != nil {
		doc := gui.pdfDoc
		go func() {
			_, _ = gui.pdfInstance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})
		}()
	}
	fyne.Do(func() {
		gui.pdfGen.Add(1) // invalidate any in-flight render
		gui.pdfDoc = nil
		gui.pdfPageCount = 0
		gui.pdfCurrentPage = 0
		gui.previewBox.RemoveAll()
		gui.setPreviewTitle("PDF preview — select a file to begin")
	})
}

// renderPage renders the given (0-indexed) page into the preview area. It is always
// invoked from the main thread.
func (gui *GUI) renderPage(page int) {
	if gui.pdfInstance == nil || gui.pdfDoc == nil || gui.pdfPageCount == 0 {
		return
	}
	doc := gui.pdfDoc.Document
	pageCount := gui.pdfPageCount
	if page < 0 {
		page = 0
	}
	if page >= pageCount {
		page = pageCount - 1
	}
	gui.pdfCurrentPage = page
	gen := gui.pdfGen.Add(1)
	go func() {
		resp, err := gui.pdfInstance.RenderToFile(&requests.RenderToFile{
			RenderPageInPixels: &requests.RenderPageInPixels{
				Page: requests.Page{ByIndex: &requests.PageByIndex{
					Document: doc,
					Index:    page,
				}},
				Width: 720,
			},
			OutputFormat: requests.RenderToFileOutputFormatPNG,
			OutputTarget: requests.RenderToFileOutputTargetBytes,
		})
		if err != nil || resp.ImageBytes == nil || len(*resp.ImageBytes) == 0 {
			fyne.Do(func() {
				if gui.pdfGen.Load() == gen {
					gui.setPreviewTitle("Could not render page: " + err.Error())
				}
			})
			return
		}
		png := *resp.ImageBytes
		fyne.Do(func() {
			if gui.pdfGen.Load() != gen {
				return // a newer render superseded this one
			}
			gui.renderPageImage(page, png)
		})
	}()
}

// renderPageImage swaps the rendered PNG into the preview scroll area and updates
// the title. Must be called from the main thread (it mutates Fyne widgets).
func (gui *GUI) renderPageImage(page int, png []byte) {
	r := bytes.NewReader(png)
	img := canvas.NewImageFromReader(r, "preview")
	img.FillMode = canvas.ImageFillContain
	gui.previewBox.RemoveAll()
	gui.previewBox.Add(img)
	gui.setPreviewTitle(fmt.Sprintf("Page %d of %d", page+1, gui.pdfPageCount))
}

// setPreviewTitle updates the preview title label (main thread).
func (gui *GUI) setPreviewTitle(s string) {
	gui.previewTitle.SetText(s)
}

// prevPage / nextPage move one page back/forward and re-render.
func (gui *GUI) prevPage() {
	if gui.pdfInstance == nil || gui.pdfPageCount == 0 {
		return
	}
	gui.setPageEntryLockedRender(gui.pdfCurrentPage - 1)
}

func (gui *GUI) nextPage() {
	if gui.pdfInstance == nil || gui.pdfPageCount == 0 {
		return
	}
	gui.setPageEntryLockedRender(gui.pdfCurrentPage + 1)
}

// setPageEntryLockedRender updates the entry (without triggering a re-render loop)
// and renders the requested page.
func (gui *GUI) setPageEntryLockedRender(page int) {
	if gui.pdfInstance == nil || gui.pdfPageCount == 0 {
		return
	}
	if page < 0 || page >= gui.pdfPageCount {
		return
	}
	gui.pageEntryLocked = true
	gui.pageEntry.SetText(fmt.Sprintf("%d", page+1))
	gui.pageEntryLocked = false
	gui.renderPage(page)
}

// onPageEntryChanged jumps to the page number the user typed in the entry.
func (gui *GUI) onPageEntryChanged(text string) {
	if gui.pageEntryLocked || gui.pdfInstance == nil || gui.pdfPageCount == 0 {
		return
	}
	var n int
	if _, err := fmt.Sscanf(text, "%d", &n); err != nil {
		return
	}
	p := n - 1
	if p < 0 || p >= gui.pdfPageCount {
		return
	}
	gui.renderPage(p)
}

// refreshInfo shows the selected PDF's metadata in the Info tab, read through
// pdfcpu's Go API — no external binary required.
func (gui *GUI) refreshInfo() {
	if gui.inputPath == "" {
		gui.inputPageCount = 0
		fyne.Do(func() {
			gui.setInfoContent("**PDF info**\n\n*Select a file to see details.*")
			gui.updateSummary()
		})
		return
	}
	go func() {
		ctx := context.Background()
		f, err := os.Open(gui.inputPath)
		var text string
		pageCount := 0
		if err != nil {
			text = "Could not open file: " + err.Error()
		} else {
			defer f.Close()
			info, e := api.PDFInfo(ctx, f, gui.inputPath, nil, true, nil)
			if e != nil {
				text = "Could not read PDF info: " + e.Error()
			} else {
				pageCount = info.PageCount
				text = buildPdfInfoText(info)
			}
		}
		gui.inputPageCount = pageCount
		fyne.Do(func() {
			gui.setInfoContent("**PDF info**\n\n" + text)
			gui.updateSummary()
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

// showErrorPopup shows a copyable error popup: the message is wrapped and
// selectable, with a button to copy the full text to the clipboard.
func (gui *GUI) showErrorPopup(message string) {
	content := container.NewVBox(
		widget.NewButton("Copy to clipboard", func() {
			if cb := fyne.CurrentApp().Clipboard(); cb != nil {
				cb.SetContent(message)
			}
		}),
		&widget.Label{
			Text:       message,
			Wrapping:   fyne.TextWrapWord,
			Selectable: true,
		},
	)
	dialog.ShowCustom("Error", "OK", content, gui.window)
}

// processBooklet runs the full creation workflow: it optionally stamps page
// numbers (cover left blank), optionally splits the input into volumes (each
// written to its own subfolder), and creates a booklet for each resulting
// document. The "Add page numbering" and "Split into volumes" checkboxes on the
// Create panel select which options apply.
func (gui *GUI) processBooklet() {
	if gui.inputPath == "" {
		gui.showErrorPopup("Please select an input PDF file")
		return
	}
	config := gui.buildConfig()
	if err := validateConfig(config); err != nil {
		gui.showErrorPopup(err.Error())
		return
	}

	var split int
	var endPages []int
	addNumbering := gui.pageNumCheck.Checked
	if gui.splitVolumesCheck.Checked {
		var n int
		fmt.Sscanf(gui.volumesEntry.Text, "%d", &n)
		if n < 2 {
			gui.showErrorPopup("Enter a number of volumes of 2 or more to split")
			return
		}
		split = n
		if len(gui.volumeEndEntries) != split {
			gui.rebuildVolumeEnds()
		}
		endPages = make([]int, len(gui.volumeEndEntries))
		for i, e := range gui.volumeEndEntries {
			t := strings.TrimSpace(e.Text)
			if t == "" || t == "l" || t == "L" {
				endPages[i] = 0 // 'l' or blank -> to the end (valid only for the last volume)
			} else {
				var v int
				if _, err := fmt.Sscanf(t, "%d", &v); err != nil {
					gui.showErrorPopup(fmt.Sprintf("Volume %d: enter a valid page number or 'l'", i+1))
					return
				}
				endPages[i] = v
			}
		}
	}

	gui.statusLabel.SetText("Creating booklet...")
	go func() {
		ctx := context.Background()
		err := createBookletWithOptions(ctx, config, addNumbering, split, endPages)
		fyne.Do(func() {
			if err != nil {
				gui.statusLabel.SetText("Error")
				gui.showErrorPopup(err.Error())
			} else if split > 0 {
				gui.statusLabel.SetText(fmt.Sprintf("Created %d volume(s) — see the volume_* folders under the output path.", split))
			} else {
				gui.statusLabel.SetText("Booklet created! See output, sections/ and print_ready/ folders.")
			}
		})
	}()
}

// reset clears the form and preview, returning to the Preview tab.
func (gui *GUI) reset() {
	gui.inputEntry.SetText("")
	gui.outputEntry.SetText("output/booklet.pdf")
	gui.pagesSelect.Selected = "1"
	gui.directionSelect.Selected = "RTL"
	gui.sectionsEntry.SetText("8")
	gui.blankSelect.Selected = "1"
	gui.pageNumCheck.SetChecked(false)
	gui.splitVolumesCheck.SetChecked(false)
	gui.volumesEntry.SetText("0")
	gui.setVolumeOptionsVisible(false)
	gui.rebuildVolumeEnds()
	gui.inputPath = ""
	gui.inputPageCount = 0
	gui.clearPdf()
	gui.statusLabel.SetText("Ready")
	gui.showTab("preview")
	gui.refreshInfo()
}

// clearOutput removes the folder that holds the generated files (the directory of
// the current output path) so the user can start fresh. It never deletes a path
// that does not exist or the process working directory.
func (gui *GUI) clearOutput() {
	out := gui.outputEntry.Text
	dir := filepath.Dir(out)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		gui.showErrorPopup("Invalid output path.")
		return
	}
	st, err := os.Stat(absDir)
	if err != nil || !st.IsDir() {
		gui.addResultLine("Nothing to clear (output folder does not exist yet).")
		gui.statusLabel.SetText("Ready")
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		gui.showErrorPopup("Could not determine working directory.")
		return
	}
	absCwd, _ := filepath.Abs(cwd)
	if absDir == absCwd {
		gui.addResultLine("Refusing to clear the current working directory. Use an output folder (e.g. output/).")
		gui.statusLabel.SetText("Ready")
		return
	}
	if err := os.RemoveAll(absDir); err != nil {
		gui.showErrorPopup("Could not clear output: " + err.Error())
		return
	}
	gui.addResultLine("Output cleared.")
	gui.statusLabel.SetText("Ready")
}

// setVolumeOptionsVisible adds or removes the inline volume-options box from the
// Create form. It is only shown while "Split into volumes" is checked.
func (gui *GUI) setVolumeOptionsVisible(show bool) {
	fyne.Do(func() {
		if show && !gui.volumeOptionsShown {
			gui.createContent.Add(gui.volumeOptionsBox)
			gui.volumeOptionsShown = true
		} else if !show && gui.volumeOptionsShown {
			gui.createContent.Remove(gui.volumeOptionsBox)
			gui.volumeOptionsShown = false
		}
	})
}

// rebuildVolumeEnds creates one "ends at" entry per requested volume. The last
// volume's entry accepts 'l' (last page); the others require a page number.
func (gui *GUI) rebuildVolumeEnds() {
	n := 0
	fmt.Sscanf(gui.volumesEntry.Text, "%d", &n)
	gui.volumeEndEntries = nil
	gui.volumeEndsBox.RemoveAll()
	for i := 1; i <= n; i++ {
		isLast := i == n
		ph := "Enter last page"
		if isLast {
			ph = "Enter last page (use 'l' for all remaining)"
		}
		e := widget.NewEntry()
		e.SetPlaceHolder(ph)
		e.Validator = validation.NewRegexp(`^[0-9]+$|^[lL]$`, "Page number or 'l' (last only)")
		gui.volumeEndEntries = append(gui.volumeEndEntries, e)
		gui.volumeEndsBox.Add(gui.formRow(fmt.Sprintf("Volume %d ends at", i), e))
	}
}

// updateSummary repopulates the live "What you'll get" panel under the inputs.
// It mirrors bookit.sh's result: pages per sheet, sections, pages per section,
// total booklet pages (with blanks/padding) and total A4 sheets to print.
func (gui *GUI) updateSummary() {
	fyne.Do(func() {
		gui.resultBox.RemoveAll()
		if gui.inputPageCount <= 0 {
			gui.addResultLine("Select a PDF file to see what you'll get.")
			return
		}
		pages := 1
		fmt.Sscanf(gui.pagesSelect.Selected, "%d", &pages)
		sections := 8
		fmt.Sscanf(gui.sectionsEntry.Text, "%d", &sections)
		addBlank := 1
		fmt.Sscanf(gui.blankSelect.Selected, "%d", &addBlank)

		lines, err := computeBookletSummary(gui.inputPageCount, pages, sections, addBlank)
		if err != nil {
			gui.addResultLine("Error: " + err.Error())
			return
		}
		for _, l := range lines {
			gui.addResultLine(l)
		}
	})
}

// addResultLine appends a wrapping label to the result panel.
func (gui *GUI) addResultLine(s string) {
	lb := widget.NewLabel(s)
	lb.Wrapping = fyne.TextWrapWord
	gui.resultBox.Add(lb)
}

// computeBookletSummary reproduces bookit.sh's "result" math: for a given page
// count, pages-per-sheet, sections and blank-page setting it reports the number
// of sections (signatures), pages per section, the total booklet page count (with
// blanks/padding) and the number of double-sided A4 sheets to print.
func computeBookletSummary(totalPages, pagesPerSheet, sections, addBlank int) ([]string, error) {
	if totalPages <= 0 {
		return nil, fmt.Errorf("no pages to summarize")
	}
	folioMult := folioMultiplier(pagesPerSheet)
	pagesPerSignature := sections * folioMult
	if pagesPerSignature <= 0 {
		return nil, fmt.Errorf("invalid sections/pages-per-sheet")
	}

	var finalTotal int
	if addBlank == 1 {
		afterFront := totalPages + 2
		endPages := 2
		switch afterFront % 4 {
		case 1, 3:
			endPages = 3
		}
		afterBlanks := afterFront + endPages
		remSig := afterBlanks % pagesPerSignature
		if remSig != 0 {
			finalTotal = afterBlanks + (pagesPerSignature - remSig)
		} else {
			finalTotal = afterBlanks
		}
	} else {
		finalTotal = totalPages
	}

	nSections := (finalTotal + pagesPerSignature - 1) / pagesPerSignature
	a4Sheets := (finalTotal + pagesPerSheet*2 - 1) / (pagesPerSheet * 2)

	lines := []string{
		fmt.Sprintf("Pages per sheet: %d A4 pages (double-sided)", pagesPerSheet*2),
		fmt.Sprintf("Sections (signatures): %d", nSections),
		fmt.Sprintf("Pages per section: %d", pagesPerSignature),
	}
	if addBlank == 0 && finalTotal%pagesPerSignature != 0 {
		lastSize := finalTotal - (nSections-1)*pagesPerSignature
		lines = append(lines, fmt.Sprintf("Last section: %d pages (partial)", lastSize))
	}
	lines = append(lines, fmt.Sprintf("Total booklet pages: %d", finalTotal))
	lines = append(lines, fmt.Sprintf("Total A4 sheets to print: %d", a4Sheets))
	return lines, nil
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
