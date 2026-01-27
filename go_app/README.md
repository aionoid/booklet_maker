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

## 📌 Sewing Points/Stations

The application adds sewing points/stations to assist with binding:
- **Placement**: Stations are added to even pages (back sides) only
- **Configuration**: Different numbers of stations based on page layout:
  - **1-up (A5)**: 8 points at positions 7%, 19.3%, 31.6%, 43.9%, 56.1%, 68.4%, 80.7%, 93% from the top
  - **2-up (A6)**: 6 points at positions 8%, 24.8%, 41.6%, 58.4%, 75.2%, 92% from the top
  - **4-up/8-up (A7)**: 4 points at positions 10%, 36.6%, 63.3%, 90% from the top
- **Purpose**: Helps guide where to punch holes or sew for binding
- **Positioning**: Calculated based on percentage positions along the spine of the booklet

## 🏷️ Section Marking with Folio Numbers

The application adds section marking to help organize printed sheets:
- **Placement**: Applied to even pages (back sides) only
- **Numbering**: Each section is numbered with folio numbers (01, 02, 03, etc.) following the signature structure
- **Positioning**: Marks are placed at calculated positions moving positively from the right edge (10, 25, 40, etc.) using "pos:r" positioning
- **Format**: Uses BigBlueTermPlusNFM font with right positioning, rotated 90 degrees, with gray background and 40% opacity
- **Configuration**: Number of folios per section matches the NSECTIONS parameter (e.g., if NSECTIONS=8, each section will have folios 01-08)
- **Purpose**: Helps identify and organize sections during assembly and binding
- **Page Ranges**: Each folio number is applied to corresponding even pages in the section (e.g., for 8 folios: page 2→01, page 4→02, ..., page 16→08 for the first section)

## 🛠️ Prerequisites

- Go 1.21 or higher
- pdfcpu library

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
-input, -i        Input PDF file (required)
-output, -o       Output PDF file (default: booklet.pdf)
-pages, -p        Pages per sheet (1, 2, 4, or 8) (default: 1)
-direction, -d    Reading direction (RTL or LTR) (default: RTL)
-sections, -s     Number of sections (default: 8)
-blank, -b        Add blank pages (0 or 1) (default: 1)
```

## 🏗️ Architecture

The application is structured as follows:

- `main.go` - Entry point of the application
- `cli.go` - Command-line interface handler
- `booklet.go` - Core booklet processing logic
- `Makefile` - Build and deployment scripts

## 🎯 Future Enhancements

- [ ] GUI application using a Go GUI framework
- [ ] More advanced booklet layouts
- [ ] Batch processing capabilities
- [ ] Configuration file support