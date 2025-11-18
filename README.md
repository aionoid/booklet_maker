# 📘 Booklet Maker

A simple tool for creating booklets from PDF files.

## ✨ Features

- 📄 Create booklets from PDF files
- ↔️ Support for both LTR and RTL reading directions
- 🖨️ Multiple page layouts (1-up, 2-up, 4-up, 8-up)
- 🔪 Section splitting for easy printing

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

Run the helper script to start:

```bash
./helper.sh
```

The helper script will present you with options for quick presets or you can choose to configure custom settings.

## 📖 Help / Wiki

### 🖥️ Basic Commands

- 📌 Use `./helper.sh` to start the booklet creation process
- 🎛️ The helper script will guide you through the process with various presets

### ⚙️ Custom Settings

When prompted, you can enter the following parameters:

- 📁 Input PDF: The source PDF file (default: book.pdf)
- 💾 Output PDF: The output booklet file (default: booklet.pdf)
- 📐 Pages per sheet: 1, 2, 4, or 8 (default: 2)
- ↪️ Reading direction: RTL or LTR (default: LTR)
- 🔢 Sections: Number of sections (default: 4)

### 🔧 Troubleshooting

- ✅ Make sure all required dependencies are installed
- 📂 Ensure the input PDF file exists and is accessible
- 🧪 Check that the parameters are valid (e.g., pages per sheet must be 1, 2, 4, or 8)
