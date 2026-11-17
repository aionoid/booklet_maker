# Changelog

All notable changes to Booklet Maker will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
[1.0.0]: https://github.com/aionoid/booklet_maker/releases/tag/v1.0.0
