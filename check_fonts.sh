#!/usr/bin/env bash

# Function to check if a font is installed in pdfcpu
check_font_installed() {
    local font_name="$1"
    local font_file="$2"
    
    # Check if the font is in the list of installed fonts
    if pdfcpu fonts list 2>/dev/null | grep -q "$font_name"; then
        echo "✓ Font '$font_name' is already installed"
        return 0
    else
        echo "✗ Font '$font_name' is not installed"
        return 1
    fi
}

# Function to install a font in pdfcpu
install_font() {
    local font_file="$1"
    
    if [ ! -f "$font_file" ]; then
        echo "✗ Font file '$font_file' does not exist"
        return 1
    fi
    
    echo "Installing font: $font_file"
    if pdfcpu fonts install "$font_file"; then
        echo "✓ Font installed successfully"
        return 0
    else
        echo "✗ Failed to install font: $font_file"
        return 1
    fi
}

# Main function to check and install fonts if needed
check_and_install_fonts() {
    echo "Checking for required fonts..."
    
    # Check for BigBlueTermPlusNerdFontMono-Regular.ttf
    # The font might be installed under a different name, so we'll check for variations
    if ! pdfcpu fonts list 2>/dev/null | grep -E -q "BigBlueTerm|NerdFont|BigBlueTermPlus"; then
        echo "Installing BigBlueTermPlusNerdFontMono-Regular.ttf..."
        install_font "./fonts/BigBlueTermPlusNerdFontMono-Regular.ttf"
    else
        echo "✓ BigBlueTermPlusNerdFontMono font is already installed"
    fi
    
    # You can add more fonts here as needed
    # Example:
    # if ! check_font_installed "SomeOtherFont" "./fonts/some_other_font.ttf"; then
    #     install_font "./fonts/some_other_font.ttf"
    # fi
}

# Run the check if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    # Make sure we're in the nix shell environment
    if ! command -v pdfcpu &> /dev/null; then
        echo "pdfcpu is not available. Please run this script within the nix shell environment:"
        echo "  nix-shell --run './check_fonts.sh'"
        exit 1
    fi
    
    check_and_install_fonts
fi