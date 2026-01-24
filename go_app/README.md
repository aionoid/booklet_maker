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