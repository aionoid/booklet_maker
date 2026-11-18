package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// BookletConfig holds the configuration for booklet creation.
type BookletConfig struct {
	InputFile        string
	OutputFile       string
	PagesPerSheet    int
	ReadingDirection string // "RTL" or "LTR"
	Sections         int
	AddBlank         int // 0 or 1
}

// copyConfig returns a copy of the given configuration so callers can override
// individual fields without mutating the original.
func copyConfig(c *BookletConfig) *BookletConfig {
	return &BookletConfig{
		InputFile:        c.InputFile,
		OutputFile:       c.OutputFile,
		PagesPerSheet:    c.PagesPerSheet,
		ReadingDirection: c.ReadingDirection,
		Sections:         c.Sections,
		AddBlank:         c.AddBlank,
	}
}

const (
	nupDesc     = "form:A4, g:off, border:on, margin:0, bgcol:#beded9"
	stationFont = "BigBlueTermPlusNFM"
	folioFont   = "Courier-Bold"
	// pageNumbersDesc is the stamp used for content page numbers — it matches the
	// shell scripts (helper.sh): a tiny Courier-Bold number at bottom-center with
	// 80% opacity, so it reads as a subtle folio rather than a big watermark.
	pageNumbersDesc = "fontname:Courier-Bold,scalefactor:0.9 rel,position:bc,margins:1 10,offset:0 1,points:1,fillcolor:#000000,rotation:0,opacity:0.8"
)

// folioMultiplier returns the folio multiplier for a given pages-per-sheet value.
func folioMultiplier(pagesPerSheet int) int {
	switch pagesPerSheet {
	case 1, 2:
		return 2
	case 4:
		return 4
	case 8:
		return 8
	default:
		return 2
	}
}

// ProcessBooklet processes a PDF file to create a booklet and print-ready pages.
func ProcessBooklet(config *BookletConfig) error {
	ctx := context.Background()

	fmt.Printf("Processing booklet: %s -> %s\n", config.InputFile, config.OutputFile)
	fmt.Printf("Config: pagesPerSheet=%d, direction=%s, sections=%d, addBlank=%d\n",
		config.PagesPerSheet, config.ReadingDirection, config.Sections, config.AddBlank)

	// Verify input file exists.
	if _, err := os.Stat(config.InputFile); err != nil {
		return fmt.Errorf("input file does not exist: %w", err)
	}

	pagesPerSignature := config.Sections * folioMultiplier(config.PagesPerSheet)

	outDir := filepath.Dir(config.OutputFile)
	sectionDir := filepath.Join(outDir, "sections")
	readyDir := filepath.Join(outDir, "print_ready")

	// Cleanup previous runs.
	os.Remove(config.OutputFile)
	if err := os.RemoveAll(sectionDir); err != nil {
		return fmt.Errorf("removing section dir: %w", err)
	}
	if err := os.RemoveAll(readyDir); err != nil {
		return fmt.Errorf("removing print-ready dir: %w", err)
	}
	if err := os.MkdirAll(sectionDir, 0755); err != nil {
		return fmt.Errorf("creating section dir: %w", err)
	}
	if err := os.MkdirAll(readyDir, 0755); err != nil {
		return fmt.Errorf("creating print-ready dir: %w", err)
	}

	// Prepare PDF with blank pages (temp file, or the original when AddBlank is 0).
	prepared, err := prepareBookletPages(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to prepare booklet pages: %w", err)
	}
	if prepared != config.InputFile {
		defer os.Remove(prepared)
	}

	// Handle reading direction by reversing pages if RTL.
	source := prepared
	var reversed string
	if config.ReadingDirection == "RTL" {
		reversed = prepared + ".rev"
		if err := reversePages(ctx, prepared, reversed); err != nil {
			return fmt.Errorf("failed to handle reading direction: %w", err)
		}
		source = reversed
		defer os.Remove(reversed)
	}

	// Create the booklet (2-up, foliosize=Sections).
	if err := createBooklet(ctx, source, config.OutputFile, config.Sections); err != nil {
		return fmt.Errorf("failed to create booklet: %w", err)
	}

	// Add stations (sewing points) to the booklet.
	if err := addStations(ctx, config.OutputFile, config.PagesPerSheet); err != nil {
		return fmt.Errorf("failed to add stations: %w", err)
	}

	// Add section/folio marking to the booklet.
	if err := addSectionMarking(ctx, config.OutputFile, config.Sections); err != nil {
		return fmt.Errorf("failed to add section marking: %w", err)
	}

	// Split into sections of pagesPerSignature pages.
	if err := api.SplitFile(ctx, config.OutputFile, sectionDir, pagesPerSignature, nil); err != nil {
		return fmt.Errorf("failed to split sections: %w", err)
	}

	// Generate front/back print pages with n-up layout.
	if err := generatePrintPages(ctx, config.PagesPerSheet, pagesPerSignature, sectionDir, readyDir); err != nil {
		return fmt.Errorf("failed to generate print pages: %w", err)
	}

	fmt.Printf("Booklet created successfully: %s\n", config.OutputFile)
	return nil
}

// prepareBookletPages copies the input and inserts the required blank pages.
// It returns the path of the file to use for booklet creation.
func prepareBookletPages(ctx context.Context, config *BookletConfig) (string, error) {
	if config.AddBlank == 0 {
		return config.InputFile, nil
	}

	dir := filepath.Dir(config.InputFile)
	base := filepath.Join(dir, strings.TrimSuffix(filepath.Base(config.InputFile), ".pdf")+"_prepared.pdf")

	if err := copyFile(config.InputFile, base); err != nil {
		return "", fmt.Errorf("copying input: %w", err)
	}

	info, err := pdfInfo(ctx, config.InputFile)
	if err != nil {
		os.Remove(base)
		return "", fmt.Errorf("reading page count from PDF: %w", err)
	}
	totalPages := info.PageCount
	fmt.Printf("Original PDF has %d pages\n", totalPages)

	totalAfterFront := totalPages + 2
	endPages := 2
	switch r := totalAfterFront % 4; {
	case r == 1, r == 3:
		endPages = 3
	}
	totalAfterBlanks := totalAfterFront + endPages

	pagesPerSignature := config.Sections * folioMultiplier(config.PagesPerSheet)
	remSig := totalAfterBlanks % pagesPerSignature
	pagesToAdd := 0
	if remSig != 0 {
		pagesToAdd = pagesPerSignature - remSig
	}

	fmt.Printf("Adding 2 blank pages at front...\n")
	fmt.Printf("Adding %d blank pages at end for booklet format...\n", endPages)
	fmt.Printf("Adding %d blank pages to fill last section...\n", pagesToAdd)

	for i := 0; i < 2; i++ {
		if err := insertBlank(ctx, base, []string{"1"}, true); err != nil {
			os.Remove(base)
			return "", fmt.Errorf("failed to add blank page at front: %w", err)
		}
	}
	for i := 0; i < endPages+pagesToAdd; i++ {
		if err := insertBlank(ctx, base, []string{"l"}, false); err != nil {
			os.Remove(base)
			return "", fmt.Errorf("failed to add blank page at end: %w", err)
		}
	}

	return base, nil
}

// insertBlank inserts a blank page before or after the selected pages of a PDF.
func insertBlank(ctx context.Context, file string, selectedPages []string, before bool) error {
	return api.InsertPagesFile(ctx, file, file, selectedPages, before, nil, nil)
}

// reversePages writes a new PDF with the pages of inFile in reverse order.
func reversePages(ctx context.Context, in, out string) error {
	info, err := pdfInfo(ctx, in)
	if err != nil {
		return fmt.Errorf("reading page count: %w", err)
	}
	pages := make([]string, 0, info.PageCount)
	for i := info.PageCount; i >= 1; i-- {
		pages = append(pages, strconv.Itoa(i))
	}
	return api.CollectFile(ctx, in, out, pages, nil)
}

// createBooklet creates the booklet using a 2-up multifolio layout.
func createBooklet(ctx context.Context, in, out string, nsections int) error {
	desc := fmt.Sprintf("multifolio:on, foliosize:%d, g:off, ma:5, border:on, bgcol:#beded9, or:ld", nsections)
	fmt.Printf("Creating booklet with: %s\n", desc)
	nup, err := api.PDFBookletConfig(2, desc, nil)
	if err != nil {
		return fmt.Errorf("booklet config: %w", err)
	}
	return api.BookletFile(ctx, []string{in}, out, nil, nup, nil)
}

// addStations stamps sewing points along the spine, one per configured percentage.
func addStations(ctx context.Context, pdfFile string, pagesPerSheet int) error {
	var stationsConfig []string
	switch pagesPerSheet {
	case 1: // A5 (1-up) - 8 points
		stationsConfig = []string{"7%", "19.3%", "31.6%", "43.9%", "56.1%", "68.4%", "80.7%", "93%"}
	case 2: // A6 (2-up) - 6 points
		stationsConfig = []string{"8%", "24.8%", "41.6%", "58.4%", "75.2%", "92%"}
	default: // A7 (4-up/8-up) - 4 points
		stationsConfig = []string{"10%", "36.6%", "63.3%", "90%"}
	}

	info, err := pdfInfo(ctx, pdfFile)
	if err != nil {
		return fmt.Errorf("stations: %w", err)
	}
	pageWidth := pageWidthFromInfo(info)
	fmt.Printf("Adding stations to %s, configuration: %v\n", pdfFile, stationsConfig)
	fmt.Printf("Page width: %s\n", offsetStr(pageWidth))

	for _, percentStr := range stationsConfig {
		var percent float64
		if _, err := fmt.Sscanf(strings.Trim(percentStr, "%"), "%f", &percent); err != nil {
			fmt.Printf("Warning: skipping invalid percentage %s: %v\n", percentStr, err)
			continue
		}
		xPos := percent / 100 * pageWidth
		desc := fmt.Sprintf("fontname:%s,pos:l,offset:%s 6,points:2,scale:0.03,fillc:#000000,rot:0", stationFont, offsetStr(xPos))
		if err := api.AddTextWatermarksFile(ctx, pdfFile, pdfFile, nil, true, ".", desc, nil); err != nil {
			fmt.Printf("Warning: could not add station at %s: %v\n", percentStr, err)
			continue
		}
		fmt.Printf("  Station added successfully at x=%s to all pages\n", offsetStr(xPos))
	}

	return nil
}

// addSectionMarking stamps per-section folio numbers on the front (odd) pages.
func addSectionMarking(ctx context.Context, pdfFile string, nsections int) error {
	info, err := pdfInfo(ctx, pdfFile)
	if err != nil {
		return fmt.Errorf("section marking: %w", err)
	}
	totalPages := info.PageCount
	fmt.Printf("Adding section marking to %s\n", pdfFile)
	fmt.Printf("Total pages in booklet: %d\n", totalPages)

	foliosPerMarkingSection := nsections
	pagesPerMarkingSection := foliosPerMarkingSection * 2
	totalMarkingSections := totalPages / pagesPerMarkingSection
	if totalPages%pagesPerMarkingSection != 0 {
		totalMarkingSections++
	}

	basePosition := 10
	positionIncrement := 15

	for sectionNum := 1; sectionNum <= totalMarkingSections; sectionNum++ {
		sectionPosition := basePosition + (sectionNum-1)*positionIncrement
		startPage := (sectionNum-1)*pagesPerMarkingSection + 1
		endPage := sectionNum * pagesPerMarkingSection
		if endPage > totalPages {
			endPage = totalPages
		}

		firstOddPage := startPage
		if firstOddPage%2 == 0 {
			firstOddPage++
		}

		folioNum := 1
		oddPage := firstOddPage
		for oddPage <= endPage && folioNum <= foliosPerMarkingSection && oddPage <= totalPages {
			folioNumberInSection := fmt.Sprintf("%02d", folioNum)
			desc := fmt.Sprintf("fontname:%s,pos:r,offset:%d 0,points:2,scale:0.04,fillc:#000000,backgroundc:#808080,rot:-90,opacity:0.9", folioFont, -sectionPosition)
			pages := []string{strconv.Itoa(oddPage)}
			fmt.Printf("Adding folio marking %s to page %d at position: %d (marking section %d)\n",
				folioNumberInSection, oddPage, sectionPosition, sectionNum)
			if err := api.AddTextWatermarksFile(ctx, pdfFile, pdfFile, pages, true, folioNumberInSection, desc, nil); err != nil {
				fmt.Printf("Warning: could not add folio marking %s on page %d: %v\n", folioNumberInSection, oddPage, err)
			} else {
				fmt.Printf("  Folio marking %s added successfully on page %d at position %d\n", folioNumberInSection, oddPage, sectionPosition)
			}
			folioNum++
			oddPage += 2
		}
	}

	fmt.Println("All section markings added successfully to " + pdfFile)
	return nil
}

// generatePrintPages splits the booklet into front/back print files per section
// and applies the n-up layout.
func generatePrintPages(ctx context.Context, pagesPerSheet int, pagesPerSignature int, sectionDir, readyDir string) error {
	entries, err := os.ReadDir(sectionDir)
	if err != nil {
		return fmt.Errorf("reading section dir: %w", err)
	}
	var sectionFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".pdf") {
			sectionFiles = append(sectionFiles, filepath.Join(sectionDir, e.Name()))
		}
	}
	sort.Slice(sectionFiles, func(i, j int) bool {
		return naturalLess(filepath.Base(sectionFiles[i]), filepath.Base(sectionFiles[j]))
	})

	totalFiles := len(sectionFiles)
	for i, file := range sectionFiles {
		ind := i + 1
		fout := filepath.Base(file)
		fmt.Printf("Processing section %d/%d: %s\n", ind, totalFiles, fout)

		if err := generate1upPages(ctx, file, ind, fout, pagesPerSignature, pagesPerSheet, readyDir); err != nil {
			return fmt.Errorf("section %d: %w", ind, err)
		}
		if pagesPerSheet != 1 {
			if err := applyNupLayout(ctx, ind, fout, pagesPerSheet, readyDir); err != nil {
				return fmt.Errorf("section %d nup layout: %w", ind, err)
			}
		}
	}

	fmt.Printf("Generated files in: %s/\n", readyDir)
	return nil
}

// generate1upPages collects the front (odd) and back (even) pages of a section.
func generate1upPages(ctx context.Context, file string, ind int, fout string, pagesPerSignature, pagesPerSheet int, readyDir string) error {
	frontPages := seq(1, pagesPerSignature, 2)
	frontFile := filepath.Join(readyDir, fmt.Sprintf("%d_F_%s", ind, fout))
	if err := api.CollectFile(ctx, file, frontFile, frontPages, nil); err != nil {
		return fmt.Errorf("collecting front pages: %w", err)
	}

	var backPages []string
	if pagesPerSheet == 1 {
		backPages = seqRev(pagesPerSignature, 2, 2) // even pages in reverse order
	} else {
		backPages = seq(2, pagesPerSignature, 2) // even pages in forward order
	}
	backFile := filepath.Join(readyDir, fmt.Sprintf("%d_B_%s", ind, fout))
	if err := api.CollectFile(ctx, file, backFile, backPages, nil); err != nil {
		return fmt.Errorf("collecting back pages: %w", err)
	}

	return nil
}

// applyNupLayout imposes the front and back pages of a section into an n-up layout.
func applyNupLayout(ctx context.Context, ind int, fout string, pps int, readyDir string) error {
	frontFile := filepath.Join(readyDir, fmt.Sprintf("%d_F_%s", ind, fout))
	backFile := filepath.Join(readyDir, fmt.Sprintf("%d_B_%s", ind, fout))

	tempFront := strings.TrimSuffix(frontFile, ".pdf") + "_temp.pdf"
	tempBack := strings.TrimSuffix(backFile, ".pdf") + "_temp.pdf"

	// Nup the front pages (right-to-left orientation).
	if err := nup(ctx, []string{frontFile}, tempFront, nil, pps, nupDesc+", orientation:rd"); err != nil {
		return fmt.Errorf("nup front pages: %w", err)
	}

	// Rotate the back file 180 degrees for a 2-up section booklet.
	if pps == 2 {
		if err := api.RotateFile(ctx, backFile, backFile, 180, nil, nil); err != nil {
			return fmt.Errorf("rotating back pages: %w", err)
		}
	}

	// Nup the back pages (dr for 2-up, otherwise ld).
	backOrient := "ld"
	if pps == 2 {
		backOrient = "dr"
	}
	if err := nup(ctx, []string{backFile}, tempBack, nil, pps, nupDesc+", orientation:"+backOrient); err != nil {
		return fmt.Errorf("nup back pages: %w", err)
	}

	// Reverse the back pages for direct printing in place.
	info, err := pdfInfo(ctx, tempBack)
	if err != nil {
		return fmt.Errorf("reading back page count: %w", err)
	}
	backPages := seqRev(info.PageCount, 1, 1)
	if err := api.CollectFile(ctx, tempBack, tempBack, backPages, nil); err != nil {
		return fmt.Errorf("reversing back pages: %w", err)
	}

	if err := os.Rename(tempFront, frontFile); err != nil {
		return fmt.Errorf("renaming temp front: %w", err)
	}
	if err := os.Rename(tempBack, backFile); err != nil {
		return fmt.Errorf("renaming temp back: %w", err)
	}

	fmt.Printf("Applied %d-up layout to section %d\n", pps, ind)
	return nil
}

// nup imposes the given number of pages per sheet using the provided description.
func nup(ctx context.Context, inFiles []string, outFile string, selectedPages []string, pps int, desc string) error {
	nup, err := api.PDFNUpConfig(pps, desc, nil)
	if err != nil {
		return fmt.Errorf("nup config: %w", err)
	}
	return api.NUpFile(ctx, inFiles, outFile, selectedPages, nup, nil)
}

// pdfInfo opens a PDF and returns its metadata.
func pdfInfo(ctx context.Context, fileName string) (*pdfcpu.PDFInfo, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return api.PDFInfo(ctx, f, fileName, nil, false, nil)
}

// pageWidthFromInfo returns the first page's width in points, falling back to A4.
func pageWidthFromInfo(info *pdfcpu.PDFInfo) float64 {
	if len(info.PageBoundaries) > 0 {
		pb := info.PageBoundaries[0]
		if pb.Media != nil && pb.Media.Rect != nil {
			return pb.Media.Rect.UR.X - pb.Media.Rect.LL.X
		}
	}
	if len(info.Dimensions) > 0 {
		return info.Dimensions[0].Width
	}
	return 595 // A4 width fallback in points
}

// copyFile copies src to dst.
func copyFile(src, dst string) error {
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()
	dstF, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstF.Close()
	if _, err := io.Copy(dstF, srcF); err != nil {
		return err
	}
	return nil
}

// AddPageNumbers stamps page numbers onto content pages of a PDF, leaving page 1
// (the cover) unnumbered. Content pages are stamped with their actual page
// number using pdfcpu's "%p" placeholder (one in-place stamp call).
func AddPageNumbers(ctx context.Context, input, output string) error {
	info, err := pdfInfo(ctx, input)
	if err != nil {
		return fmt.Errorf("reading input for page numbers: %w", err)
	}
	totalPages := info.PageCount
	if totalPages < 2 {
		// Nothing to number; just copy the file.
		return copyFile(input, output)
	}
	fmt.Printf("Adding page numbers to content pages 2-%d of %s (cover left blank)\n", totalPages, input)
	return api.AddTextWatermarksFile(ctx, input, output, seq(2, totalPages, 1), true, "%p", pageNumbersDesc, nil)
}

// SplitIntoVolumes splits a PDF into n volumes. Each volume contains the cover
// (original page 1, stamped "VOL k") followed by a contiguous range of content
// pages. Returns the volume paths.
//
// endPages (len == n) gives the last content page of each volume (0 or "l"
// means "to the end", only valid for the last volume). When endPages is empty
// the pages are distributed evenly (backward compatible with the CLI).
//
// numberContent, when true, stamps page numbers (cover left blank) on each
// volume's content pages. When false the volumes are returned unnumbered so a
// caller can number them itself (e.g. via AddPageNumbers) before imposing.
func SplitIntoVolumes(ctx context.Context, input, outDir string, n int, endPages []int, numberContent bool) ([]string, error) {
	info, err := pdfInfo(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("reading input for volumes: %w", err)
	}
	totalPages := info.PageCount
	if totalPages < 2 || n <= 0 {
		return nil, fmt.Errorf("cannot split into %d volume(s): input has %d pages", n, totalPages)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("creating volume dir: %w", err)
	}

	contentStart := 2
	contentEnd := totalPages

	ends, err := computeEndPages(contentStart, contentEnd, n, endPages)
	if err != nil {
		return nil, err
	}

	nextStart := contentStart
	var vols []string
	for v := 1; v <= n; v++ {
		end := ends[v-1]
		start := nextStart
		nextStart = end + 1
		if start > end {
			continue
		}

		volName := fmt.Sprintf("vol_%02d.pdf", v)
		base := strings.TrimSuffix(volName, ".pdf")
		tmp := filepath.Join(outDir, base+"_tmp.pdf")

		// Combined file: cover (page 1) + this volume's content pages.
		pages := []string{"1"}
		pages = append(pages, seq(start, end, 1)...)
		if err := api.CollectFile(ctx, input, tmp, pages, nil); err != nil {
			os.Remove(tmp)
			return nil, fmt.Errorf("collecting pages for volume %d: %w", v, err)
		}

		// Stamp "VOL k" on the cover (page 1).
		coverDesc := fmt.Sprintf("fontname:%s,scalefactor:1 rel,position:bc,margins:10 10,offset:0 20,points:1,fillcolor:#000000,bgcolor:#808080,rotation:0,opacity:0.8", folioFont)
		if err := api.AddTextWatermarksFile(ctx, tmp, tmp, []string{"1"}, true, fmt.Sprintf("VOL %d", v), coverDesc, nil); err != nil {
			os.Remove(tmp)
			return nil, fmt.Errorf("stamping cover for volume %d: %w", v, err)
		}

		// Number content pages (2..len) with their in-volume page number — only when
		// numberContent is requested, using the same tiny stamp as AddPageNumbers.
		volInfo, err := pdfInfo(ctx, tmp)
		if err != nil {
			os.Remove(tmp)
			return nil, fmt.Errorf("reading volume %d: %w", v, err)
		}
		numPages := volInfo.PageCount
		if numPages > 1 && numberContent {
			if err := api.AddTextWatermarksFile(ctx, tmp, tmp, seq(2, numPages, 1), true, "%p", pageNumbersDesc, nil); err != nil {
				os.Remove(tmp)
				return nil, fmt.Errorf("numbering volume %d pages: %w", v, err)
			}
		}

		outPath := filepath.Join(outDir, volName)
		if err := os.Rename(tmp, outPath); err != nil {
			os.Remove(tmp)
			return nil, fmt.Errorf("writing volume %d: %w", v, err)
		}
		vols = append(vols, outPath)
		fmt.Printf("Volume %d/%d: %s (content pages %d-%d)\n", v, n, outPath, start, end)
	}

	if len(vols) == 0 {
		return nil, fmt.Errorf("no volumes produced")
	}
	fmt.Printf("Split into %d volume(s) in %s\n", len(vols), outDir)
	return vols, nil
}

// createBookletWithOptions runs the full creation workflow used by the GUI's
// "Create Booklet" action. It optionally stamps page numbers (cover left blank),
// optionally splits the input into `split` volumes (each written to its own
// subfolder named volume_01, volume_02, ...), and creates a booklet for each
// resulting document. When split is 0 it behaves like ProcessBooklet on the
// (optionally numbered) single input.
func createBookletWithOptions(ctx context.Context, config *BookletConfig, addNumbering bool, split int, endPages []int) error {
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Single-document path.
	if split <= 0 {
		cfg := copyConfig(config)
		return createSingle(ctx, cfg.InputFile, cfg, addNumbering)
	}

	// Multi-volume path: split first (unnumbered), then booklet each volume into
	// its own subfolder, numbering first if requested.
	outDir := filepath.Dir(config.OutputFile)
	volDir := filepath.Join(outDir, "volumes")
	vols, err := SplitIntoVolumes(ctx, config.InputFile, volDir, split, endPages, false)
	if err != nil {
		return fmt.Errorf("splitting into volumes: %w", err)
	}
	for i, vol := range vols {
		v := i + 1
		volOutDir := filepath.Join(outDir, fmt.Sprintf("volume_%02d", v))
		if err := os.MkdirAll(volOutDir, 0755); err != nil {
			return fmt.Errorf("creating volume %d dir: %w", v, err)
		}
		cfg := copyConfig(config)
		cfg.OutputFile = filepath.Join(volOutDir, "booklet.pdf")
		if err := createSingle(ctx, vol, cfg, addNumbering); err != nil {
			return fmt.Errorf("creating booklet for volume %d: %w", v, err)
		}
	}
	return nil
}

// createSingle creates a booklet from one document (config.OutputFile is the
// output location), optionally stamping page numbers onto the content (cover
// left blank) before imposing.
func createSingle(ctx context.Context, input string, config *BookletConfig, addNumbering bool) error {
	source := input
	if addNumbering {
		base := filepath.Join(filepath.Dir(input), filepath.Base(input)+"_numbered.pdf")
		if err := AddPageNumbers(ctx, input, base); err != nil {
			return fmt.Errorf("adding page numbers: %w", err)
		}
		source = base
		defer os.Remove(base)
	}
	cfg := copyConfig(config)
	cfg.InputFile = source
	return ProcessBooklet(cfg)
}

// computeEndPages returns the last content page of each volume. With no explicit
// end pages it distributes the content pages evenly across n volumes (giving the
// extra pages to the leading volumes). When explicitEndPages has n entries it
// uses them as the per-volume ends; a 0 or "l" entry means "to the end" and is
// only valid for the last volume. Returns an error if any end is out of range
// or the ends are not strictly increasing.
func computeEndPages(contentStart, contentEnd int, n int, explicitEndPages []int) ([]int, error) {
	var ends []int
	if len(explicitEndPages) == n {
		ends = make([]int, n)
		copy(ends, explicitEndPages)
	} else {
		// Even split: distribute (contentEnd-contentStart+1) pages across n
		// volumes, giving the extra pages to the first `rem` volumes.
		contentCount := contentEnd - contentStart + 1
		base := contentCount / n
		rem := contentCount % n
		prev := contentStart - 1
		for i := 0; i < n; i++ {
			size := base
			if i < rem {
				size++
			}
			prev += size
			ends = append(ends, prev)
		}
	}
	// Normalize "l"/0 (last page) and validate the range.
	for i := range ends {
		e := ends[i]
		if e <= 0 {
			if i == n-1 {
				e = contentEnd // 'l' or blank on the last volume -> use all remaining pages
			} else {
				return nil, fmt.Errorf("volume %d: end page must be a positive number ('l' is only allowed for the last volume)", i+1)
			}
		}
		if e < contentStart {
			return nil, fmt.Errorf("volume %d: end page %d is before the first content page (%d)", i+1, e, contentStart)
		}
		if e > contentEnd {
			e = contentEnd // clamp to the last page
		}
		ends[i] = e
	}
	for i := 1; i < n; i++ {
		if ends[i] <= ends[i-1] {
			return nil, fmt.Errorf("volume end pages must strictly increase: volume %d ends at %d (previous volume %d ends at %d)", i+1, ends[i], i, ends[i-1])
		}
	}
	return ends, nil
}

// seq returns [start, start+step, ..., <=end] as page-number strings.
func seq(start, end, step int) []string {
	var out []string
	for i := start; i <= end; i += step {
		out = append(out, strconv.Itoa(i))
	}
	return out
}

// seqRev returns [start, start-step, ..., >=end] as page-number strings.
func seqRev(start, end, step int) []string {
	var out []string
	for i := start; i >= end; i -= step {
		out = append(out, strconv.Itoa(i))
	}
	return out
}

// offsetStr formats a float for use in a stamp offset, avoiding scientific notation.
func offsetStr(f float64) string {
	return strconv.FormatFloat(f, 'f', -6, 64)
}

// naturalLess reports whether baseA sorts before baseB using natural (numeric-aware)
// ordering, so e.g. booklet_9-16.pdf comes before booklet_17-32.pdf.
func naturalLess(a, b string) bool {
	ta := intTokens(a)
	tb := intTokens(b)
	for i := range ta {
		if i >= len(tb) {
			return false
		}
		if ta[i] != tb[i] {
			return ta[i] < tb[i]
		}
	}
	return len(ta) < len(tb)
}

// intTokens splits a name into its numeric tokens for natural sorting.
func intTokens(s string) []int64 {
	var out []int64
	var num int64
	inNum := false
	for _, r := range s {
		if r >= '0' && r <= '9' {
			num = num*10 + int64(r-'0')
			inNum = true
		} else if inNum {
			out = append(out, num)
			num = 0
			inNum = false
		}
	}
	if inNum {
		out = append(out, num)
	}
	return out
}
