#!/usr/bin/env bash

# Configuration
BOOK="${1:-book.pdf}"
OUT="${2:-booklet.pdf}"
PAGES_PER_SHEET="${3:-1}"     #1:booklet, 2, 4, or 8 pages per sheet
READING_DIRECTION="${4:-RTL}" # RTL or LTR
NSECTIONS="${5:-8}"
ADD_BLANK="${6:-1}"
DIR_SECTION="sections"
DIR_READY="print_ready"
DISPLAY_CMD="figlet"                      # or toilet
DISPLAY_FONT="./fonts/figlet/ANSI_Shadow" # Font for DISPLAY_CMD
BOOK_REV="pdf"
BOOK_PRE="pdf"
PDFCPU_NUP="form:A4, g:off, border:on, margin:0, bgcol:#beded9"
BOOKLET_CMD="multifolio:on, foliosize:$NSECTIONS, g:off, ma:5, border:on, bgcol:#beded9, or:ld"

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

# Prepare PDF with blank pages for booklet
prepare_booklet_pages() {
        if [[ "$ADD_BLANK" != 0 ]]; then

                $DISPLAY_CMD -t -f $DISPLAY_FONT "Prep Pages"
                echo "Preparing PDF with blank pages for booklet..."

                local input_pdf="$BOOK"
                # local temp_prepared="${input_pdf%.pdf}_prepared.pdf"
                BOOK_PRE="${input_pdf%.pdf}_prepared.pdf"

                # Get total pages from original PDF
                TOTAL_PAGES=$(pdfcpu info "$input_pdf" 2>/dev/null | grep "Page count:" | awk '{print $3}')

                if [ -z "$TOTAL_PAGES" ]; then
                        echo "Error: Could not read page count from PDF."
                        return 1
                fi

                echo "Original PDF has $TOTAL_PAGES pages"

                # Calculate total pages after adding 2 front pages
                TOTAL_AFTER_FRONT=$((TOTAL_PAGES + 2))

                # Calculate how many end pages to add (2 or 3)
                # We want the final total to be a multiple of 4 for proper booklet printing
                REMAINDER=$((TOTAL_AFTER_FRONT % 4))
                case $REMAINDER in
                0) END_PAGES=2 ;; # Need 2 more to make it multiple of 4
                1) END_PAGES=3 ;; # Need 3 more to reach multiple of 4
                2) END_PAGES=2 ;; # Need 2 more to reach multiple of 4
                3) END_PAGES=3 ;; # Need 1 more, but we'll add 3 to maintain pattern
                esac

                # For consistency, let's use this logic:
                # Always add 2 at front, then add enough at end to make total multiple of 4
                # But minimum 2 at end, maximum 3
                if [ $END_PAGES -lt 2 ]; then
                        END_PAGES=2
                elif [ $END_PAGES -gt 3 ]; then
                        END_PAGES=3
                fi

                FINAL_TOTAL=$((TOTAL_AFTER_FRONT + END_PAGES))

                echo "Adding 2 blank pages at front..."
                echo "Adding $END_PAGES blank pages at end..."
                echo "Final booklet will have $FINAL_TOTAL pages (multiple of 4: $((FINAL_TOTAL % 4 == 0)))"

                # Create temporary copy
                cp "$input_pdf" "$BOOK_PRE"

                # Step 1: Add 2 blank pages at the beginning
                echo "Step 1/3: Adding 2 blank pages at front..."
                pdfcpu pages insert -m before -p 1 -- "$BOOK_PRE"
                pdfcpu pages insert -m before -p 1 -- "$BOOK_PRE"

                # Step 2: Add calculated blank pages at the end
                echo "Step 2/3: Adding $END_PAGES blank pages at end..."
                if [ $END_PAGES -eq 2 ]; then
                        pdfcpu pages insert -m after -p l -- "$BOOK_PRE"
                        pdfcpu pages insert -m after -p l -- "$BOOK_PRE"
                else
                        pdfcpu pages insert -m after -p l -- "$BOOK_PRE"
                        pdfcpu pages insert -m after -p l -- "$BOOK_PRE"
                        pdfcpu pages insert -m after -p l -- "$BOOK_PRE"
                fi

                echo "Page preparation completed:"
                echo "  - 2 blank pages at front"
                echo "  - $END_PAGES blank pages at end"
                echo "  - Total pages: $FINAL_TOTAL"

                # echo "$temp_prepared"
                # Use the prepared file for booklet creation
                BOOK="$BOOK_PRE"
        fi
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
                # local temp_book="${BOOK%.pdf}_reversed.pdf"
                BOOK_REV="${BOOK%.pdf}_reversed.pdf"
                # pdfcpu collect -q -p "$rev_pages" "$BOOK" "$temp_book"
                pdfcpu collect -q -p "$rev_pages" "$BOOK" "$BOOK_REV"

                # Use the reversed file for booklet creation
                BOOK="$BOOK_REV"
        fi
}

# Create booklet - Always use 2-up for initial booklet creation
create_booklet() {
        $DISPLAY_CMD -t -f $DISPLAY_FONT "3: Create Booklet"

        # Prepare PDF with blank pages for booklet
        prepare_booklet_pages

        # Handle reading direction first
        handle_reading_direction

        # Create booklet with 2-up layout using the prepared PDF
        echo "Creating booklet with: $BOOKLET_CMD"
        pdfcpu booklet -- "$BOOKLET_CMD" "$OUT" 2 "$BOOK"

        # Clean up temporary files
        if [[ "$ADD_BLANK" != 0 && -f "$BOOK_PRE" ]]; then
                echo "CLEAN $BOOK_PRE"
                rm "$BOOK_PRE"
        fi

        if [[ "$READING_DIRECTION" == "RTL" && -f "$BOOK_REV" ]]; then
                echo "CLEAN $BOOK_REV"
                rm "$BOOK_REV"
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
        pdfcpu nup -- "${PDFCPU_NUP}, orientation:rd" "$temp_front" 2 "$front_file"

        # rotate 180 for 2up booklet in section
        pdfcpu rotate "$back_file" 180
        # Apply nup to back pages (left-to-right orientation)
        pdfcpu nup -- "${PDFCPU_NUP}, orientation:dr" "$temp_back" 2 "$back_file"

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
        pdfcpu nup -- "${PDFCPU_NUP}, orientation:rd" "$temp_front" 4 "$front_file"

        # Apply nup to back pages (left-to-right orientation)
        pdfcpu nup -- "${PDFCPU_NUP}, orientation:ld" "$temp_back" 4 "$back_file"

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
        pdfcpu nup -- "${PDFCPU_NUP}, orientation:rd" "$temp_front" 8 "$front_file"

        # Apply nup to back pages (left-to-right orientation)
        pdfcpu nup -- "${PDFCPU_NUP}, orientation:ld" "$temp_back" 8 "$back_file"

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
