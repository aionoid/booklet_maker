# 📘 Booklet Maker

A simple tool for creating booklets from PDF files.

## ✨ Features

- 📄 Create booklets from PDF files
- ↔️ Support for both LTR and RTL reading directions
- 🖨️ Multiple page layouts (1-up, 2-up, 4-up, 8-up)
- 🔪 Section splitting for easy printing
- 📄 Automatic blank page insertion for proper booklet formatting and section filling

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
