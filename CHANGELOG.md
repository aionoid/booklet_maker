# Changelog

All notable changes to Booklet Maker will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2026-11-17

### Added
- **GUI PDF info tab**: shows document metadata read via pdfcpu's Go API — page count, page size (and paper name), title, author, subject, creator, producer, creation/modification dates, encryption/linearized/tagged/bookmarks flags, and embedded fonts.
- **App version** displayed in the GUI footer.
- **GUI two-tab left side**: the configuration controls now sit under two tabs — **Create** (input/output, pages per sheet, reading direction, sections, blank pages, page numbers) and **Prepare** (volume splitting). The Prepare tab adds a per-volume **"Volume N ends at"** field for every requested volume (the last one accepts `l` = "to the end") — the missing per-volume end-page option from `helper.sh`/`bookit.sh`.
- **Live "What you'll get" results panel**: under the inputs, shows number of sections, pages per section, total booklet pages (with blank padding) and total double-sided A4 sheets to print. Updates immediately as pages-per-sheet / sections / blank change or after picking a PDF file, mirroring `bookit.sh`'s printed result.
- **Test coverage**: `computeEndPages` unit tests (even split, explicit ends, "to the end" last volume, clamping, validation errors) and `computeBookletSummary` unit tests (sections / pages-per-section / total pages / A4 sheets across blank and no-blank layouts).

### Changed
- **GUI layout**: Preview/Info panel pinned to a fixed 400 px column on the right; configuration controls moved to the left as a scrollable stack — [Create/Prepare tab bar] → [active tab inputs] → [live results panel] → [action buttons] — with left-aligned labels in a fixed-width column so all inputs align. The results panel sits directly under the inputs.
- **Preview**: thumbnails rendered to fit the column (Fyne `ImageFillContain` + fixed cell) so every page size fits; the preview title wraps so it never widens the panel.
- **Errors**: pipeline and validation errors now appear in a copyable popup dialog (wrapped, selectable text + a "Copy to clipboard" button via the Fyne clipboard API) instead of the footer status label, which no longer expands the window; the footer status label wraps long text and fills the space up to the version.
- `booklet.go`: `SplitIntoVolumes` now accepts an optional per-volume end-page list (`computeEndPages`) in addition to even distribution; CLI keeps even-split behaviour (backward compatible).

### Fixed
- GUI preview thread errors: wrapped all goroutine UI mutations in `fyne.Do` and gave the app a unique ID (`app.NewWithID("booklet-maker")`).
- Squished preview thumbnails (Fyne default stretch + lazy image → 1×1 min size): set `canvas.ImageFillOriginal`, later `ImageFillContain`, so pages show correctly.
- Fyne `OnChanged` wiring for `Select`/`Entry` updated to the v2.7.2 field API (`func(string)`) so live updates fire on change.

## [Unreleased]

### Fixed
- **Page numbers / volume stamps rendered too large in the Go app** (relative to the shell scripts). The Go pipeline stamped page numbers at `points:8` (absolute scale) and volume covers at `points:24`, producing oversized numerals. Both now use the exact stamp parameters from `helper.sh`:
  - Page numbers: `fontname:Courier-Bold, points:1, scalefactor:0.9 rel, position:bc, margins:1 10, offset:0 1, fillcolor:#000000, rotation:0, opacity:0.8` (bottom-center).
  - Volume covers: `fontname:Courier-Bold, points:1, scalefactor:1 rel, position:bc, margins:10 10, offset:0 20, fillcolor:#000000, bgcolor:#808080 (gray badge), rotation:0, opacity:0.8` (bottom-center).
  - Key names use unambiguous long forms (`margins:` / `bgcolor:`) because short prefixes such as `ma` are ambiguous in the pdfcpu parser (`ma` matches both `margins` and `maxWidth`).

## [1.1.0] - 2026-11-17

### Added
- **Page numbering** (helper.sh option 7): `-page-numbers` / `-pnum` stamps a page number on every content page, leaving the first page (cover) unnumbered. Uses pdfcpu's `%p` placeholder in a single watermark pass.
- **Volume splitting** (helper.sh option 8): `-volumes N` / `-vol N` splits the PDF into N volumes, each with its own cover page (the source's first page, stamped "VOL k") followed by an evenly distributed contiguous range of content pages with per-volume page numbers. Output goes to a `volumes/` directory next to the input file (`vol_01.pdf`, `vol_02.pdf`, ...).
- **GUI PDF preview panel**: the Fyne GUI now shows the selected input PDF's pages as a scrollable thumbnail list on the left, rendered via `pdftoppm` (poppler-utils); refreshes when a new file is chosen.

### Changed
- `booklet.go`: added `AddPageNumbers` and `SplitIntoVolumes` pipeline functions; fixed volume range calculation so content pages are split evenly with no overlap or gaps.
- `cli.go`: added `-page-numbers`/`-pnum` and `-volumes`/`-vol N` flags (alternative modes that operate on the raw input).
- `gui.go`: full rewrite — two-column layout (left preview + right options form), page-numbering and volume-splitting buttons, and a live PDF preview.

### Dependencies
- `shell.nix` now includes `poppler-utils` (provides `pdftoppm`) for the GUI preview.

## [1.0.0] - 2026-11-17

### Added
- **Go application** (`go_app/`) — a pure Go CLI + Fyne GUI built on pdfcpu's **Go API**
  (`github.com/pdfcpu/pdfcpu/pkg/api`) directly, instead of shelling out to the `pdfcpu` binary.
  - `main.go` entry point (CLI vs `-gui`), `booklet.go` (full pipeline), `cli.go` (argument parsing), `gui.go` (Fyne GUI), `Makefile` (build/test targets).
  - Same six options as `bookit.sh`: `-input`, `-output`, `-pages` (1/2/4/8), `-direction` (RTL/LTR), `-sections` (default 8), `-blank` (0/1).
  - Reproduces the full pipeline output: `booklet.pdf`, `sections/*.pdf`, and `print_ready/*_F.pdf` / `print_ready/*_B.pdf`.
- Go test suite in `go_app/` (`go test ./...`) covering CLI parsing, module build, vet, and fmt.
- `pdfcpu-go-api` skill (software-development) documenting pdfcpu Go API signatures, type locations, NUp/stamp config formats, and gotchas.

### Changed
- `booklet.go`: pipeline rewritten to call the pdfcpu Go API (`PDFInfo`, `InsertPagesFile`, `CollectFile`, `BookletFile`/`PDFBookletConfig`, `AddTextWatermarksFile`, `RotateFile`, `SplitFile`, `NUpFile`/`PDFNUpConfig`).
  - Stations now stamped on **all** pages with per-page width-based x-positions (previously a single spine-X on even pages).
  - Folio numbers stamped on front (odd) pages; booklet `foliosize` now set to the sections count (previously hardcoded 8).
- Output subdirectories (`sections/`, `print_ready/`) now created relative to the output file's directory instead of the process CWD.
- Section files processed in natural (numeric-aware) order, fixing alphabetical mis-ordering (e.g. `booklet_17-…` before `booklet_9-…`).
- Dropped the broken `ma:1 1` token from the folio stamp description — pdfcpu rejects it as an "ambiguous parameter prefix" — so folio markings now actually render.

### Fixed
- Import of the `PDFInfo` type corrected to its real package path (`pkg/pdfcpu`, not the module root).
- A return statement missing its string value (type mismatch in the blank-page loop).
- Unused `"os"` import in `gui_test.go`; `Label.Wrapping` set to `fyne.TextWrapWord`.

### Removed
- Dead root Go module (`main.go`, `go.mod`, `go.sum`) — `go_app/` is now the single Go module.
- Stale `.qwen/PROJECT_SUMMARY.md` AI artifact and `test_section_logic.sh` scratch script.

### Dependencies
- `pdfcpu` pinned to **v0.16.4** (Go API).
- Fyne **v2.7.2** for the GUI.
- `shell.nix` now includes `libx11` and `libxxf86vm` for the GUI build.

[Unreleased]: https://github.com/aionoid/booklet_maker/commits/main
[1.2.0]: https://github.com/aionoid/booklet_maker/releases/tag/v1.2.0
[1.1.0]: https://github.com/aionoid/booklet_maker/releases/tag/v1.1.0
[1.0.0]: https://github.com/aionoid/booklet_maker/releases/tag/v1.0.0
