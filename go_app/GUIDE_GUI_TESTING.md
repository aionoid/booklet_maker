# Testing the Booklet Maker GUI

## Prerequisites

Before testing the GUI functionality, ensure you have:

1. Go programming language (version 1.19 or later)
2. Fyne toolkit dependencies for your platform:
   - Linux: GTK3, pkg-config, libcairo2-dev, libglib2.0-dev
   - macOS: Xcode command line tools
   - Windows: MSYS2 or Visual Studio build tools

## Building and Testing the GUI

### 1. Install Dependencies

First, make sure you have the required dependencies:

```bash
# On Ubuntu/Debian
sudo apt-get install libgtk-3-dev libcairo2-dev pkg-config

# On Fedora/RHEL
sudo dnf install gtk3-devel cairo-devel pkgconfig

# On macOS
xcode-select --install

# On Windows with MSYS2
pacman -S mingw-w64-x86_64-gtk3 mingw-w64-x86_64-cairo
```

### 2. Navigate to the Go Application Directory

```bash
cd go_app
```

### 3. Download Dependencies

```bash
go mod tidy
```

### 4. Test the Build

```bash
go build .
```

### 5. Run the GUI

```bash
go run main.go -gui
```

Or alternatively:

```bash
./go_app -gui
```

## Expected Behavior

When running the GUI mode:

1. A window titled "Booklet Maker" should appear
2. The window should contain form fields for:
   - Input File (with Browse button)
   - Output File
   - Pages Per Sheet (dropdown: 1, 2, 4, 8)
   - Reading Direction (dropdown: RTL, LTR)
   - Sections (numeric input)
   - Add Blank Pages (dropdown: 0, 1)
3. The Browse button should open a file dialog filtered for PDF files
4. The Submit button should process the booklet with the provided parameters
5. Progress updates should appear in the status label

## Unit Tests

To run the unit tests:

```bash
go test ./...
```

## Integration Tests

To run integration tests:

```bash
go test -tags=integration ./...
```

## Troubleshooting

### Common Issues:

1. **Missing GUI dependencies**: Install the platform-specific GUI libraries mentioned above
2. **Build errors**: Run `go mod tidy` to ensure all dependencies are downloaded
3. **File dialog not working**: Ensure proper file system permissions

### Verifying Installation:

If the GUI launches successfully, you should see the main window with all form elements properly displayed.