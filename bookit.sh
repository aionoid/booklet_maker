#!/usr/bin/env bash
#TODO:
# 1. FIXME: on page reverse calculate the total so to add blank pages to the front on RTL

# Configuration
BOOK="${1:-book.pdf}"
OUT="${2:-booklet.pdf}"
PAGES_PER_SHEET="${3:-1}"     #1:booklet, 2, 4, or 8 pages per sheet
READING_DIRECTION="${4:-RTL}" # RTL or LTR
NSECTIONS="${5:-4}"
DIR_SECTION="sections"
DIR_READY="print_ready"
DISPLAY_CMD="figlet"                      # or toilet
DISPLAY_FONT="./fonts/figlet/ANSI_Shadow" # Font for DISPLAY_CMD

# Validation
validate_input() {
        if [[ ! -f "$BOOK" ]]; then
                echo "Error: Input file '$BOOK' not found!"
                exit 1
        fi

        if [[ ! "$PAGES_PER_SHEET" =~ ^(1|2|4|8)$ ]]; then
                echo "Error: Pages per sheet must be 1:booklet, 2, 4, or 8"
                exit 1
        fi

        if [[ ! "$READING_DIRECTION" =~ ^(RTL|LTR)$ ]]; then
                echo "Error: Reading direction must be RTL or LTR"
                exit 1
        fi
}

# Calculate derived values
calculate_values() {
        case $PAGES_PER_SHEET in
        1) FOLIO_MULTIPLIER=2 ;;
        2) FOLIO_MULTIPLIER=2 ;;
        4) FOLIO_MULTIPLIER=4 ;;
        8) FOLIO_MULTIPLIER=8 ;;
        esac

        PAGES_PER_SIGNATURE=$((NSECTIONS * FOLIO_MULTIPLIER))

        echo "=== Booklet Configuration ==="
        echo "Input: $BOOK"
        echo "Output: $OUT"
        echo "Pages per sheet: $((PAGES_PER_SHEET * 2)) "
        echo "Reading direction: $READING_DIRECTION "
        echo "Sections: $NSECTIONS"
        echo "Pages per signature: $((PAGES_PER_SIGNATURE * 2))"
        echo "=============================="
}

# Cleanup function
cleanup() {
        $DISPLAY_CMD -t -f $DISPLAY_FONT "1: Clean"
        echo "Cleaning previous runs..."
        rm -f "$OUT"
        rm -rf "$DIR_SECTION" "$DIR_READY"
}

# Preparation function
prepare() {
        $DISPLAY_CMD -t -f $DISPLAY_FONT "2: Prepare"
        mkdir -p "$DIR_SECTION" "$DIR_READY"
}

# Handle RTL reading direction by reversing pages using collect
handle_reading_direction() {
        if [[ "$READING_DIRECTION" == "RTL" ]]; then
                $DISPLAY_CMD -t -f $DISPLAY_FONT "RTL: Reverse"
                echo "Reversing pages for RTL reading direction..."

                # Get total page count
                local total_pages=$(pdfcpu info "$BOOK" | grep "Page count:" | awk '{print $3}')
                echo "Total pages: $total_pages"

                # Generate reverse page sequence (last page to first page)
                local rev_pages=$(generate_sequence "$total_pages" -1 1)

                # Create reversed PDF
                local temp_book="${BOOK%.pdf}_reversed.pdf"
                pdfcpu collect -q -p "$rev_pages" "$BOOK" "$temp_book"

                # Use the reversed file for booklet creation
                BOOK="$temp_book"
        fi
}

# Create booklet - Always use 2-up for initial booklet creation
create_booklet() {
        $DISPLAY_CMD -t -f $DISPLAY_FONT "3: Create Booklet"

        # Handle reading direction first
        handle_reading_direction

        # Create booklet with 2-up layout first
        local booklet_cmd="multifolio:on, foliosize:$NSECTIONS, g:on, or:ld"

        echo "Creating booklet with: $booklet_cmd"
        pdfcpu booklet -- "$booklet_cmd" "$OUT" 2 "$BOOK"

        # Clean up temporary reversed file if it exists
        if [[ "$READING_DIRECTION" == "RTL" && -f "${BOOK%.pdf}_reversed.pdf" ]]; then
                rm "${BOOK%.pdf}_reversed.pdf"
        fi
}

# Split into sections
split_sections() {
        $DISPLAY_CMD -t -f $DISPLAY_FONT "4: Split Sections"
        echo "Splitting into $PAGES_PER_SIGNATURE-page sections..."
        pdfcpu split -q "$OUT" "$DIR_SECTION" "$PAGES_PER_SIGNATURE"
}

# Generate front/back pages
generate_print_pages() {
        $DISPLAY_CMD -t -f $DISPLAY_FONT "5: Print Pages"

        local IND=0
        local total_files=$(ls -1 "$DIR_SECTION"/*.pdf 2>/dev/null | wc -l)

        for FILE in $(ls -1tr "$DIR_SECTION"/*.pdf 2>/dev/null); do
                ((IND++))
                echo "Processing section $IND/$total_files: $(basename "$FILE")"

                local FOUT=$(basename "$FILE")

                # Always generate 1-up pages first
                generate_1up_pages "$FILE" "$IND" "$FOUT"

                # Apply n-up layout for 4 or 8 pages per sheet
                if [[ "$PAGES_PER_SHEET" =~ ^(2|4|8)$ ]]; then
                        apply_nup_layout "$IND" "$FOUT"
                fi
        done
}

# 2 pages per sheet (original logic)
generate_1up_pages() {
        local file="$1" ind="$2" fout="$3"

        # Front pages (odd numbers)
        local front_pages=$(generate_sequence 1 2 "$PAGES_PER_SIGNATURE")
        pdfcpu collect -q -p "$front_pages" "$file" "$DIR_READY/${ind}_F_$fout"

        if [[ $PAGES_PER_SHEET == 1 ]]; then
                # back pages (even numbers)
                local back_pages=$(generate_sequence "$PAGES_PER_SIGNATURE" -2 2)
                # local back_pages=$(generate_sequence 2 2 "$pages_per_signature")
                pdfcpu collect -q -p "$back_pages" "$file" "$DIR_READY/${ind}_B_$fout"
        else
                # back pages (even numbers)
                local back_pages=$(generate_sequence 2 2 "$PAGES_PER_SIGNATURE")
                pdfcpu collect -q -p "$back_pages" "$file" "$DIR_READY/${ind}_B_$fout"

        fi
}

# Apply n-up layout for 4 or 8 pages per sheet
apply_nup_layout() {
        local ind="$1" fout="$2"

        local front_file="$DIR_READY/${ind}_F_$fout"
        local back_file="$DIR_READY/${ind}_B_$fout"

        $DISPLAY_CMD -t -f $DISPLAY_FONT "6: N-up Layout"
        echo "Applying $PAGES_PER_SHEET-up layout..."

        case $PAGES_PER_SHEET in
        2)
                # For 2-up (1x2 layout)
                apply_2up_layout "$front_file" "$back_file" "$ind" "$fout"
                ;;
        4)
                # For 4-up (2x2 layout)
                apply_4up_layout "$front_file" "$back_file" "$ind" "$fout"
                ;;
        8)
                # For 8-up (4x2 layout)
                apply_8up_layout "$front_file" "$back_file" "$ind" "$fout"
                ;;
        esac
}

# Apply 2-up layout (1x2)
apply_2up_layout() {
        local front_file="$1" back_file="$2" ind="$3" fout="$4"

        local temp_front="${front_file%.pdf}_temp.pdf"
        local temp_back="${back_file%.pdf}_temp.pdf"

        # Apply nup to front pages (right-to-left orientation)
        pdfcpu nup -- "form:A4, guides:on, orientation:rd" "$temp_front" 2 "$front_file"

        # rotate 180 for 2up booklet in section
        pdfcpu rotate "$back_file" 180
        # Apply nup to back pages (left-to-right orientation)
        pdfcpu nup -- "form:A4, guides:on, orientation:dr" "$temp_back" 2 "$back_file"

        # Replace original files
        mv "$temp_front" "$front_file"
        mv "$temp_back" "$back_file"

        echo "Applied 2-up layout to section $ind"
}

# Apply 4-up layout (2x2)
apply_4up_layout() {
        local front_file="$1" back_file="$2" ind="$3" fout="$4"

        local temp_front="${front_file%.pdf}_temp.pdf"
        local temp_back="${back_file%.pdf}_temp.pdf"

        # Apply nup to front pages (right-to-left orientation)
        pdfcpu nup -- "form:A4, guides:on, orientation:rd" "$temp_front" 4 "$front_file"

        # Apply nup to back pages (left-to-right orientation)
        pdfcpu nup -- "form:A4, guides:on, orientation:ld" "$temp_back" 4 "$back_file"

        # Replace original files
        mv "$temp_front" "$front_file"
        mv "$temp_back" "$back_file"

        echo "Applied 4-up layout to section $ind"
}

# Apply 8-up layout (4x2)
apply_8up_layout() {
        local front_file="$1" back_file="$2" ind="$3" fout="$4"

        local temp_front="${front_file%.pdf}_temp.pdf"
        local temp_back="${back_file%.pdf}_temp.pdf"

        # Apply nup to front pages (right-to-left orientation)
        pdfcpu nup -- "form:A4, guides:on, orientation:rd" "$temp_front" 8 "$front_file"

        # Apply nup to back pages (left-to-right orientation)
        pdfcpu nup -- "form:A4, guides:on, orientation:ld" "$temp_back" 8 "$back_file"

        # Replace original files
        mv "$temp_front" "$front_file"
        mv "$temp_back" "$back_file"

        echo "Applied 8-up layout to section $ind"
}

# Helper function to generate number sequences
generate_sequence() {
        local start="$1" step="$2" end="$3"

        # Handle reverse sequences (negative step)
        if [[ $step -lt 0 ]]; then
                if [[ $start -ge $end ]]; then
                        seq -s "," "$start" "$step" "$end"
                else
                        # If start < end, generate empty sequence (shouldn't happen for reverse)
                        echo ""
                fi
        else
                # Forward sequence
                seq -s "," "$start" "$step" "$end"
        fi
}

# Main execution
main() {
        validate_input
        calculate_values
        cleanup
        prepare
        create_booklet
        split_sections
        generate_print_pages
        echo "=== Processing Complete ==="
        echo "Generated files in: $DIR_READY/"
}

# Run main function
main "$@"
