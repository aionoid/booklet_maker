# 📘 Booklet Maker

A simple tool for creating booklets from PDF files.

## ✨ Features

- 📄 Create booklets from PDF files
- ↔️ Support for both LTR and RTL reading directions
- 🖨️ Multiple page layouts (1-up, 2-up, 4-up, 8-up)
- 🔪 Section splitting for easy printing
- 📄 Automatic blank page insertion for proper booklet formatting and section filling
- 📌 Sewing points/stations for binding guidance
- 🏷️ Section marking with folio numbers for organization
- 📖 Optional page numbering on content pages (cover left blank)
- 📚 Split into multiple volumes — each volume gets its own subfolder, cover page and per-volume page numbering

## 🛠️ Installation

The following dependencies are required:

- 🐚 bash
- 📚 pdfcpu (PDF manipulation tool)

Install pdfcpu using your system's package manager or from [pdfcpu GitHub](https://github.com/pdfcpu/pdfcpu).

For Nix users, you can use the provided `shell.nix` file for easy setup:

```bash
nix-shell
```
This will create a development environment with all required dependencies pre-installed.

## 🚀 Usage

### 🖥️ Using the GUI (New!)

Run the GUI mode to get a visual interface for creating booklets:

```bash
cd go_app
go run main.go -gui
```

The GUI provides an easy way to select options and browse for PDF files. The left side
holds a single **Create** form: the booklet settings (input/output, pages per sheet,
reading direction, sections, blank pages) plus two checkboxes — **"Add page numbering"**
and **"Split into volumes"**. Checking *Split into volumes* reveals the volume options
inline (a "Number of volumes" field and a "Volume N ends at" field for every volume,
where the last one accepts `l` = "to the end"). A single **Create Booklet** button drives
the whole pipeline: it optionally stamps page numbers, splits the PDF into per-volume
subfolders (`volume_01/`, `volume_02/`, ...), and builds a booklet for each volume.
Below the inputs is a live **"What you'll get"** results panel showing sections, pages
per section, total booklet pages and total A4 sheets. On the right is a PDF preview
panel (page thumbnails) and an Info tab showing document metadata (page count, page
size, title, fonts, etc.).

### 🤖 Using the Helper Script (Recommended)

Run the helper script to start with presets:

```bash
./helper.sh
```

The helper script will present you with quick presets or allow you to configure custom settings.

### ⚙️ Using the Core Script Directly

You can also run the main script directly:

```bash
./bookit.sh [input_pdf] [output_pdf] [pages_per_sheet] [reading_direction] [sections] [add_blank]
```

### 🚀 Using the Go Application (CLI + GUI)

The Go application uses pdfcpu's Go API directly — no separate `pdfcpu` binary is needed.

Build and run the CLI:

```bash
cd go_app
go build .
./booklet-maker -input mybook.pdf -pages 2 -direction LTR
```

Or run the GUI mode (unified Create form with "Add page numbering" and "Split into
volumes" checkboxes, a live "what you'll get" results panel, a PDF preview panel and a
document info tab):

```bash
cd go_app
go run main.go -gui
```

The Go application accepts the same options as `bookit.sh`:

| Option | Shorthand | Description | Default |
|--------|-----------|-------------|---------|
| `-input` | `-i` | Input PDF file (required) | — |
| `-output` | `-o` | Output booklet file | `booklet.pdf` |
| `-pages` | `-p` | Pages per sheet (1, 2, 4, or 8) | 1 |
| `-direction` | `-d` | Reading direction (RTL or LTR) | RTL |
| `-sections` | `-s` | Number of sections | 8 |
| `-blank` | `-b` | Add blank pages (0 or 1) | 1 |
| `-page-numbers` | `-pnum` | Stamp page numbers on content pages (cover left blank) | off |
| `-volumes` | `-vol N` | Split the PDF into N volumes with covers | 0 (off) |
| `-gui` | `-g` | Run in GUI mode | off |

### 🧪 Testing the Application

To run the tests:

```bash
cd go_app
go test ./...
```

To test the GUI functionality:

```bash
cd go_app
go run main.go -gui
```

For more detailed testing instructions, see GUIDE_GUI_TESTING.md

## 📖 Help / Wiki

### 🛠️ Scripts Overview

#### 🤖 helper.sh - Interactive Assistant
Interactive script with 8 quick presets and custom configuration option:

- **0)** Standard RTL booklet (1-up) with blank pages
- **1)** Standard LTR booklet (1-up) with blank pages
- **2)** Standard RTL booklet (2-up) with blank pages
- **3)** Standard LTR booklet (2-up) with blank pages
- **4)** Compact RTL (4-up) with blank pages
- **5)** Compact LTR (4-up) with blank pages
- **6)** Custom settings (input PDF, output PDF, pages per sheet, reading direction, sections, add blank pages)
- **7)** Add page numbers to existing PDF
- **8)** Split PDF into volumes with cover pages and volume numbering

#### ⚙️ bookit.sh - Core Booklet Generator
Main script that creates booklets with the following parameters:
- **Input PDF**: Source PDF file (default: book.pdf)
- **Output PDF**: Output booklet file (default: booklet.pdf)
- **Pages per sheet**: 1, 2, 4, or 8 (default: 1, where 1 means 1-up booklet format)
- **Reading direction**: RTL (right-to-left) or LTR (left-to-right) (default: RTL)
- **Sections**: Number of sections for splitting (default: 8)
- **Add blank**: Whether to add blank pages (0: no, 1: yes) (default: 1)

### 📐 Pages Per Sheet Explained
- **1)** 1-up: Creates a traditional booklet format with 2 pages per sheet when folded
- **2)** 2-up: Places 2 content pages per sheet (like 2-up n-up layout)
- **4)** 4-up: Places 4 content pages per sheet (2x2 grid)
- **8)** 8-up: Places 8 content pages per sheet (4x2 grid)

### 🔄 Reading Direction
- **LTR)** Left-to-Right: Suitable for languages like English
- **RTL)** Right-to-Left: Suitable for languages like Arabic, Hebrew

### 📄 Blank Page Logic
The script automatically adds blank pages to ensure proper booklet printing:
1. **Booklet Format**: Adds 2 blank pages at the front and 2-3 at the end to make total pages a multiple of 4
2. **Section Filling**: Adds additional blank pages to ensure each section has the correct number of pages for complete filling
3. **Complete Sections**: Ensures all sections are properly filled for consistent printing

### 📌 Sewing Points/Stations
The script adds sewing points/stations to assist with binding:
- **Placement**: Stations are added to even pages (back sides) only using the `-p even` option
- **Configuration**: Different numbers of stations based on page layout:
  - **1-up (A5)**: 8 points at positions 7%, 19.3%, 31.6%, 43.9%, 56.1%, 68.4%, 80.7%, 93% from the top
  - **2-up (A6)**: 6 points at positions 8%, 24.8%, 41.6%, 58.4%, 75.2%, 92% from the top
  - **4-up/8-up (A7)**: 4 points at positions 10%, 36.6%, 63.3%, 90% from the top
- **Purpose**: Helps guide where to punch holes or sew for binding
- **Positioning**: Calculated based on percentage positions along the spine of the booklet
- **Visual Indicator**: Small dots (".") added using BigBlueTermPlusNFM font

### 🏷️ Section Marking with Folio Numbers
The script adds section marking to help organize printed sheets:
- **Placement**: Applied to even pages (back sides) only using the `-p even` option
- **Numbering**: Each section is numbered with folio numbers (01, 02, 03, etc.) following the signature structure
- **Positioning**: Marks are placed at calculated positions moving positively from the right edge (10, 25, 40, etc.) using "pos:r" positioning
- **Format**: Uses BigBlueTermPlusNFM font with right positioning, rotated 90 degrees, with gray background and 40% opacity
- **Configuration**: Number of folios per section matches the NSECTIONS parameter (e.g., if NSECTIONS=8, each section will have folios 01-08)
- **Purpose**: Helps identify and organize sections during assembly and binding
- **Page Ranges**: Each folio number is applied to corresponding even pages in the section (e.g., for 8 folios: page 2→01, page 4→02, ..., page 16→08 for the first section)

### ⚙️ Custom Settings

When prompted in helper.sh or when using bookit.sh directly, you can enter the following parameters:

- 📁 **Input PDF**: The source PDF file (default: book.pdf)
- 💾 **Output PDF**: The output booklet file (default: booklet.pdf)
- 📐 **Pages per sheet**: 1, 2, 4, or 8 (default: 1)
- ↪️ **Reading direction**: RTL or LTR (default: RTL)
- 🔢 **Sections**: Number of sections (default: 8)
- ➕ **Add blank pages**: Whether to add blank pages (0: no, 1: yes) (default: 1)

### 📚 Volume Splitting Feature

Option 8 in the helper script allows you to split a PDF into multiple volumes with the following features:

- **Smart Volume Naming**: Output volumes are now named using the pattern `input_book_name_volume_01.pdf`, `input_book_name_volume_02.pdf`, etc., making it clear which input file each volume originated from
- **Cover Pages**: Each volume includes the first page of the original PDF as a cover page with volume number stamp
- **Page Numbering**: Each volume (except the cover page) gets its own page numbering sequence
- **Custom Page Ranges**: You can specify custom page ranges for each volume

#### Example:
If you split `my_book.pdf` into 3 volumes, the output files will be:
- `my_book_volume_01.pdf`
- `my_book_volume_02.pdf`
- `my_book_volume_03.pdf`

### 🔧 Troubleshooting

- ✅ Make sure all required dependencies are installed
- 📂 Ensure the input PDF file exists and is accessible
- 🧪 Check that the parameters are valid (e.g., pages per sheet must be 1, 2, 4, or 8)
- 📄 Verify that the PDF is not password protected or corrupted
- 🖨️ For printing: Ensure your printer settings match the booklet format (double-sided, short-edge binding)
