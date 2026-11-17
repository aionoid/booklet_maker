# Booklet Maker - Go Edition

A pure Go CLI application for creating booklets from PDF files, using the pdfcpu library directly.

## ✨ Features

- 📄 Create booklets from PDF files
- ↔️ Support for both LTR and RTL reading directions
- 🖨️ Multiple page layouts (1-up, 2-up, 4-up, 8-up)
- 🔪 Section splitting for easy printing
- 📄 Automatic blank page insertion for proper booklet formatting and section filling
- 📌 Sewing points/stations for binding guidance
- 🏷️ Section marking with folio numbers for organization
- 🔢 Page numbering on content pages (cover left blank)
- 📚 Splitting into volumes with cover pages and per-volume numbering
- 🖼️ GUI PDF preview panel — browse the input PDF's pages as scrollable thumbnails (first 30)
- 📄 GUI PDF info tab — page count, page size, metadata, and embedded fonts (via pdfcpu)
- 🗂️ GUI two-tab left side — "Create" (booklet settings) and "Prepare" (volume splitting with per-volume end-page fields)
- 📊 Live "What you'll get" results panel under the inputs — sections, pages per section, total booklet pages and total A4 sheets, updated as settings change
- 🏷️ App version shown in the GUI footer

## 📌 Sewing Points/Stations

The application adds sewing points/stations to assist with binding:
- **Placement**: Stations are added to all pages of the booklet
- **Configuration**: Different numbers of stations based on page layout:
  - **1-up (A5)**: 8 points at positions 7%, 19.3%, 31.6%, 43.9%, 56.1%, 68.4%, 80.7%, 93% of the page width
  - **2-up (A6)**: 6 points at positions 8%, 24.8%, 41.6%, 58.4%, 75.2%, 92% of the page width
  - **4-up/8-up (A7)**: 4 points at positions 10%, 36.6%, 63.3%, 90% of the page width
- **Purpose**: Helps guide where to punch holes or sew for binding
- **Positioning**: Each station's horizontal offset is calculated as a percentage of the page width; placed near the bottom edge with a small fixed vertical offset

## 🏷️ Section Marking with Folio Numbers

The application adds section marking to help organize printed sheets:
- **Placement**: Applied to front (odd) pages of each section
- **Numbering**: Each section is numbered with folio numbers (01, 02, 03, etc.) following the signature structure
- **Positioning**: Marks are placed at calculated positions moving from the right edge (10, 25, 40, etc.) using "pos:r" positioning with a negative x offset
- **Format**: Uses Courier-Bold font with right positioning, rotated -90 degrees, with gray background and 90% opacity
- **Configuration**: Number of folios per section matches the NSECTIONS parameter (e.g., if NSECTIONS=8, each section will have folios 01-08)
- **Purpose**: Helps identify and organize sections during assembly and binding
- **Page Ranges**: Each folio number is applied to the corresponding odd page in the section (e.g., for 8 folios: page 1→01, page 3→02, ..., page 15→08 for the first section)

## 🛠️ Prerequisites

- Go 1.21 or higher
- pdfcpu library (v0.16.4)
- GUI: Fyne (v2.7.2), X11 libraries (`libx11`, `libxxf86vm`), and an X display at runtime
- GUI preview: `poppler-utils` (provides `pdftoppm`)

## 🚀 Building

```bash
# Clone the repository
git clone <repository-url>
cd booklet-maker/go_app

# Build the application
make build

# Or run directly
go run . -input mybook.pdf
```

## 📖 Usage

### Basic Usage

```bash
./bin/booklet-maker -input mybook.pdf
```

### Advanced Usage

```bash
# Create a 2-up booklet with LTR reading direction
./bin/booklet-maker -input mybook.pdf -output output.pdf -pages 2 -direction LTR

# Create a booklet with 6 sections and no blank pages
./bin/booklet-maker -input mybook.pdf -sections 6 -blank 0
```

### Command Line Options

```
-input, -i          Input PDF file (required)
-output, -o         Output PDF file (default: booklet.pdf)
-pages, -p          Pages per sheet (1, 2, 4, or 8) (default: 1)
-direction, -d      Reading direction (RTL or LTR) (default: RTL)
-sections, -s       Number of sections (default: 8)
-blank, -b          Add blank pages (0 or 1) (default: 1)
-page-numbers, -pnum  Stamp page numbers on content pages (cover left blank)
-volumes, -vol N   Split the PDF into N volumes with covers (0 = off)
-gui, -g           Run in GUI mode
```

### Page Numbering (helper.sh option 7)

Stamp a page number on every content page, leaving the first page (cover) unnumbered:

```bash
./bin/booklet-maker -input mybook.pdf --page-numbers -o numbered.pdf
# or: ./bin/booklet-maker -i mybook.pdf -pnum
```

Content pages are stamped with their actual in-document page number using pdfcpu's `%p` placeholder.

### Splitting Into Volumes (helper.sh option 8)

Split a PDF into N volumes, each with its own cover page (the first page of the source, stamped "VOL k") followed by a contiguous range of content pages (evenly distributed):

```bash
./bin/booklet-maker -input mybook.pdf -vol 3
# or: ./bin/booklet-maker -i mybook.pdf --volumes 3
```

Output volumes are written to a `volumes/` directory next to the input file, named
`vol_01.pdf`, `vol_02.pdf`, ... Each volume's cover shows "VOL k" and its content pages carry sequential page numbers.

### 🖥️ GUI (Fyne) Mode

Run `./bin/booklet-maker -gui` to open a windowed interface. It mirrors the CLI with a two-column layout: configuration controls on the left and a Preview/Info panel on the right.

The left side has two tabs:

- **Create** — the booklet settings: input/output file, pages per sheet, reading direction, sections and blank-page options, plus the page-numbering toggle.
- **Prepare** — the volume-splitting controls: how many volumes to split into, and one "Volume N ends at" field per volume (the last volume accepts `l` for "to the end"). This is the per-volume end-page option from `helper.sh`.

Under the inputs (on whichever tab is active) a live **"What you'll get"** panel shows what the current settings will produce — number of sections, pages per section, total booklet pages (with blank padding) and total double-sided A4 sheets to print. It updates immediately as you change pages-per-sheet, sections or blank pages, and after picking a PDF file.

The right side has:

- **Preview tab** — renders up to 30 pages of the input PDF as scrollable thumbnails (via `pdftoppm`).
- **Info tab** — shows document metadata read through pdfcpu's Go API: page count, page size (and paper name), title, author, subject, creator, producer, creation/modification dates, encryption/linearized/tagged/bookmarks flags, and embedded fonts.

The app version is shown in the footer. Errors are shown as copyable popups rather than in the footer.

Build it with the X11 libraries available:

```bash
nix-shell -p go pkg-config cairo glib gtk3 libx11 libxxf86vm --command "cd go_app && make build"
./go_app/bin/booklet-maker -gui
```

## 🏗️ Architecture

The application is structured as follows:

- `main.go` - Entry point of the application
- `cli.go` - Command-line interface handler
- `booklet.go` - Core booklet processing logic
- `Makefile` - Build and deployment scripts

## 🎯 Future Enhancements

- [x] GUI application using a Go GUI framework (Fyne) — done, includes a PDF preview panel
- [ ] More advanced booklet layouts
- [ ] Batch processing capabilities
- [ ] Configuration file support