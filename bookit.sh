#!/usr/bin/env bash

# Function to check if a font is installed in pdfcpu
check_font_installed() {
        local font_name="$1"

        # Check if the font is in the list of installed fonts
        if command -v pdfcpu &>/dev/null && pdfcpu fonts list 2>/dev/null | grep -E -q "$font_name"; then
                return 0
        else
                return 1
        fi
}

# Function to install a font in pdfcpu
install_font_if_needed() {
        local font_file="$1"
        local font_pattern="$2"

        if [ ! -f "$font_file" ]; then
                echo "Warning: Font file '$font_file' does not exist"
                return 1
        fi

        # Check if font is already installed (using pattern matching)
        if check_font_installed "$font_pattern"; then
                echo "✓ Font for PDF operations is already installed"
        else
                echo "Installing font: $font_file for PDF operations..."
                if pdfcpu fonts install "$font_file" 2>/dev/null; then
                        echo "✓ Font installed successfully"
                else
                        echo "⚠️  Warning: Could not install font: $font_file"
                        echo "   This may affect PDF stamping/watermarking features."
                fi
        fi
}

# Check and install required fonts if needed
check_and_install_fonts() {
        # Check for BigBlueTermPlusNerdFontMono-Regular.ttf or similar
        if ! check_font_installed "BigBlueTerm|NerdFont|BigBlueTermPlus"; then
                install_font_if_needed "./fonts/BigBlueTermPlusNerdFontMono-Regular.ttf" "BigBlueTerm|NerdFont|BigBlueTermPlus"
        fi
}

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
        echo "Pages per signature: $PAGES_PER_SIGNATURE"
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

                # Calculate total pages after adding initial blank pages
                TOTAL_AFTER_BLANKS=$((TOTAL_AFTER_FRONT + END_PAGES))

                # Calculate how many pages we need to add to fill the last section
                # Each signature needs to be a multiple of PAGES_PER_SIGNATURE pages
                REMAINDER_SIGNATURE=$((TOTAL_AFTER_BLANKS % PAGES_PER_SIGNATURE))
                if [ $REMAINDER_SIGNATURE -ne 0 ]; then
                        # Add blank pages to make it a multiple of PAGES_PER_SIGNATURE
                        PAGES_TO_ADD=$((PAGES_PER_SIGNATURE - REMAINDER_SIGNATURE))
                        FINAL_TOTAL=$((TOTAL_AFTER_BLANKS + PAGES_TO_ADD))
                        echo "Adding $PAGES_TO_ADD blank pages to fill the last section..."
                        echo "Final booklet will have $FINAL_TOTAL pages (multiple of $PAGES_PER_SIGNATURE: $((FINAL_TOTAL % PAGES_PER_SIGNATURE == 0)))"
                else
                        # Total pages already divides evenly into signatures
                        FINAL_TOTAL=$TOTAL_AFTER_BLANKS
                        PAGES_TO_ADD=0
                        echo "Total pages ($TOTAL_AFTER_BLANKS) already divides evenly into signatures of $PAGES_PER_SIGNATURE pages"
                fi

                echo "Adding 2 blank pages at front..."
                echo "Adding $END_PAGES blank pages at end for booklet format..."
                echo "Adding $PAGES_TO_ADD blank pages to fill last section..."

                # Create temporary copy
                cp "$input_pdf" "$BOOK_PRE"

                # Step 1: Add 2 blank pages at the beginning
                echo "Step 1/4: Adding 2 blank pages at front..."
                pdfcpu pages insert -m before -p 1 -- "$BOOK_PRE"
                pdfcpu pages insert -m before -p 1 -- "$BOOK_PRE"

                # Step 2: Add calculated blank pages at the end for booklet format
                echo "Step 2/4: Adding $END_PAGES blank pages at end for booklet format..."
                if [ $END_PAGES -eq 2 ]; then
                        pdfcpu pages insert -m after -p l -- "$BOOK_PRE"
                        pdfcpu pages insert -m after -p l -- "$BOOK_PRE"
                else
                        pdfcpu pages insert -m after -p l -- "$BOOK_PRE"
                        pdfcpu pages insert -m after -p l -- "$BOOK_PRE"
                        pdfcpu pages insert -m after -p l -- "$BOOK_PRE"
                fi

                # Step 3: Add pages to fill the last section
                echo "Step 3/4: Adding $PAGES_TO_ADD blank pages to fill last section..."
                for ((i = 1; i <= PAGES_TO_ADD; i++)); do
                        pdfcpu pages insert -m after -p l -- "$BOOK_PRE"
                done

                echo "Page preparation completed:"
                echo "  - 2 blank pages at front"
                echo "  - $END_PAGES blank pages at end for booklet format"
                echo "  - $PAGES_TO_ADD blank pages to fill last section"
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

# Function to add stations (sewing points) based on the x-up format
add_stations() {
        local input_pdf="$1"
        local output_pdf="$2"
        local pages_per_sheet="$3"

        # Define station configurations based on the x-up format
        case $pages_per_sheet in
        1) # A5 (1-up) - 8 points
                STATIONS_CONFIG="7%,19.3%,31.6%,43.9%,56.1%,68.4%,80.7%,93%"
                ;;
        2) # A6 (2-up) - 6 points
                STATIONS_CONFIG="8%,24.8%,41.6%,58.4%,75.2%,92%"
                ;;
        4) # A7 (4-up) - 4 points
                STATIONS_CONFIG="10%,36.6%,63.3%,90%"
                ;;
        8) # For 8-up, we'll use the same as 4-up (A7) since it's similar density
                STATIONS_CONFIG="10%,36.6%,63.3%,90%"
                ;;
        *)
                # Default to 4 points if unknown format
                STATIONS_CONFIG="10%,36.6%,63.3%,90%"
                ;;
        esac

        echo "Adding stations to $input_pdf with config: $STATIONS_CONFIG"
        echo "Using pages per sheet: $pages_per_sheet"

        # Get the page width from the PDF info (following the exact instruction from stations.md)
        local page_width=$(pdfcpu info "$input_pdf" 2>/dev/null | grep "Page si" | awk '{ print $3 }')
        if [ -z "$page_width" ]; then
                echo "Warning: Could not get page width from PDF, using default approach"
                # Fallback to a reasonable default
                page_width=500 # Default page width in points
        fi

        echo "Page width: $page_width"

        # Calculate the x position for the spine (center of the page)
        # According to stations.md: "we use the width value of the page as 100% then we will calculate the points to put as x in "offset:x 6""
        # For the spine position, we want the center of the page
        local spine_x=$(echo "$page_width" | awk '{print $1 / 2}')

        # Convert percentage string to array
        IFS=',' read -ra PERCENTAGES <<<"$STATIONS_CONFIG"

        # Process each percentage to create station stamps
        local stamp_specs=()
        for percent in "${PERCENTAGES[@]}"; do
                # Remove the % sign to get the numeric percentage
                clean_percent=$(echo "$percent" | sed 's/%//')

                # Calculate the y position based on percentage
                # The example in stations.md uses "offset:x 6" where 6 might be a fixed vertical position
                # But for proper station placement along the spine, we need to vary the y position based on percentage
                # So we'll interpret "offset:x 6" as x being the spine position and 6 being the y position
                # But we need to vary the y position based on the percentage

                # Get page height to calculate vertical positions
                local page_height=$(pdfcpu info "$input_pdf" 2>/dev/null | grep "Page si" | awk '{ print $4 }')
                if [ -z "$page_height" ]; then
                        echo "Warning: Could not get page height from PDF, using default"
                        page_height=700 # Default page height in points
                fi

                # Calculate y position based on percentage of page height
                # The percentage is from top (0%) to bottom (100%)
                y_percentage_pos=$(echo "$clean_percent $page_height" | awk '{print ($1/100) * $2}')

                # For the offset, we need to convert this to pdfcpu's coordinate system
                # The example in stations.md uses pos:l, but for spine stations, pos:c (center) might be more appropriate
                # However, to follow the example exactly, we'll use pos:l with offset
                # With pos:l, the anchor is at the left edge, so to position at spine_x we use that as x offset
                # For y offset, we want to position from the bottom up, so y_offset = y_percentage_pos
                # Actually, let's reconsider: if pos:l means left position anchor, and we want to place at center horizontally
                # and at a specific vertical position, we need to think about this differently

                # The example shows pos:l,offset:0 6 - this places the stamp at x=0 from left edge, y=6 from bottom
                # To place at spine (center), we want x = page_width/2 from left edge
                # For vertical position, if we want percentage from top, that's different from offset from bottom
                # If we want the station at Y% from top, that's (100-Y)% from bottom
                # So y_offset_from_bottom = page_height - y_percentage_pos (from top)
                y_offset_from_bottom=$(echo "$page_height $y_percentage_pos" | awk '{print $1 - $2}')

                # Create stamp specification for this station
                # Using pos:l as specified in stations.md, with calculated x (spine position) and y (vertical position from bottom)
                # Using a small dot as the station marker
                stamp_spec="fontname:BigBlueTermPlusNFM,pos:l,offset:$spine_x $y_offset_from_bottom,points:2,scale:0.03,fillc:#000000,rot:0"
                stamp_specs+=("$stamp_spec")
        done

        # Apply all station stamps to the PDF
        # Since pdfcpu stamp add can only add one stamp at a time, we'll loop through each
        local temp_pdf="$input_pdf.tmp"
        cp "$input_pdf" "$temp_pdf"

        for i in "${!stamp_specs[@]}"; do
                local stamp_spec="${stamp_specs[$i]}"
                echo "Adding station $((i + 1)): ${PERCENTAGES[$i]}"

                # Create a temporary output file for this iteration
                local temp_out="${temp_pdf%.tmp}_temp_$i.pdf"

                # Apply the station stamp
                if pdfcpu stamp add -mode text -- "." "$stamp_spec" "$temp_pdf" "$temp_out" 2>/dev/null; then
                        # Replace temp_pdf with the new version
                        mv "$temp_out" "$temp_pdf"
                else
                        echo "Warning: Could not add station $((i + 1)) to PDF"
                        # If the stamp failed, remove temp_out and continue with temp_pdf
                        [ -f "$temp_out" ] && rm "$temp_out"
                fi
        done

        # Move the final stamped PDF to the output
        mv "$temp_pdf" "$output_pdf"
        echo "Stations added successfully to $output_pdf"
}

# Function to add stations (sewing points) directly to the booklet PDF based on the x-up format
add_stations_direct() {
        local input_pdf="$1"
        local pages_per_sheet="$2"

        # Define station configurations based on the x-up format
        case $pages_per_sheet in
        1) # A5 (1-up) - 8 points
                STATIONS_CONFIG="7%,19.3%,31.6%,43.9%,56.1%,68.4%,80.7%,93%"
                ;;
        2) # A6 (2-up) - 6 points
                STATIONS_CONFIG="8%,24.8%,41.6%,58.4%,75.2%,92%"
                ;;
        4) # A7 (4-up) - 4 points
                STATIONS_CONFIG="10%,36.6%,63.3%,90%"
                ;;
        8) # For 8-up, we'll use the same as 4-up (A7) since it's similar density
                STATIONS_CONFIG="10%,36.6%,63.3%,90%"
                ;;
        *)
                # Default to 4 points if unknown format
                STATIONS_CONFIG="10%,36.6%,63.3%,90%"
                ;;
        esac

        echo "Adding stations to $input_pdf with config: $STATIONS_CONFIG"
        echo "Using pages per sheet: $pages_per_sheet"

        # Get the page width from the PDF info (following the exact instruction from stations.md)
        local page_width=$(pdfcpu info "$input_pdf" 2>/dev/null | grep "Page si" | awk '{ print $3 }')
        if [ -z "$page_width" ]; then
                echo "Warning: Could not get page width from PDF, using default approach"
                # Fallback to a reasonable default
                page_width=500 # Default page width in points
        fi

        echo "Page width: $page_width"

        # Convert percentage string to array
        IFS=',' read -ra PERCENTAGES <<<"$STATIONS_CONFIG"

        # Process each percentage to add station stamps directly to the PDF
        for percent in "${PERCENTAGES[@]}"; do
                # Remove the % sign to get the numeric percentage
                clean_percent=$(echo "$percent" | sed 's/%//')

                # Calculate the x position based on page width and percentage
                # Width value as 100%, so x = (percentage/100) * page_width
                x_pos=$(echo "$clean_percent $page_width" | awk '{print ($1/100) * $2}')

                echo "Adding station at percentage: $percent (x_position: $x_pos)"

                # Apply the station stamp directly to the PDF using the exact command format from stations.md
                # Add stations only to even pages (back sides) using -p even
                # pdfcpu stamp add -p even -mode text -- "."  "fontname:BigBlueTermPlusNFM ,pos:l,offset:x 6, points:2,scale:0.03, fillc:#000000, rot:0" input.pdf
                # Replace 'x' with the calculated x_pos
                if pdfcpu stamp add -p even -mode text -- "." "fontname:BigBlueTermPlusNFM,pos:l,offset:$x_pos 6,points:2,scale:0.03,fillc:#000000,rot:0" "$input_pdf" 2>/dev/null; then
                        echo "  Station added successfully at x=$x_pos to all even pages"
                else
                        echo "  Warning: Could not add station at x=$x_pos to even pages"
                fi
        done

        echo "All stations added successfully to $input_pdf"
}

# Function to add section marking to the booklet PDF based on the x-up format
add_section_marking() {
        local input_pdf="$1"
        local nsections="$2"
        local pages_per_sheet="$3"

        echo "Adding section marking to $input_pdf"
        echo "Number of sections: $nsections"
        echo "Using pages per sheet: $pages_per_sheet"

        # Get total page count to determine page ranges for each section
        local total_pages=$(pdfcpu info "$input_pdf" | grep "Page count:" | awk '{print $3}')
        if [ -z "$total_pages" ]; then
                echo "Warning: Could not get page count from PDF, skipping section marking"
                return 1
        fi

        echo "Total pages in booklet: $total_pages"
        echo "Folios per signature (nsections param): $nsections"

        # For section marking, we'll use the nsections parameter to determine
        # how many folios are in each marking section
        local folios_per_marking_section=$nsections  # Use the nsections parameter
        local pages_per_marking_section=$((folios_per_marking_section * 2))  # 2 pages per folio

        echo "Folios per marking section: $folios_per_marking_section"
        echo "Pages per marking section: $pages_per_marking_section"

        # Calculate how many marking sections we have
        local total_marking_sections=$((total_pages / pages_per_marking_section))
        if [ $((total_pages % pages_per_marking_section)) -ne 0 ]; then
            # If there are remaining pages, we still need to mark them
            ((total_marking_sections++))
        fi

        echo "Total marking sections: $total_marking_sections"

        # Calculate the x position for the section marking based on the section number
        # Position moves by +15 for each section: first section at +10, second at +25, third at +40, etc.
        # Based on the updated requirements in sections_mark.md
        local base_position=10
        local position_increment=15

        # Add section marking for each marking section
        for ((section_num = 1; section_num <= total_marking_sections; section_num++)); do
                # Calculate the x position for this marking section
                local section_position=$((base_position + (section_num - 1) * position_increment))

                # Calculate page range for this marking section
                local start_page=$(( (section_num - 1) * pages_per_marking_section + 1 ))
                local end_page=$((section_num * pages_per_marking_section))

                # Make sure we don't exceed the total pages
                if [ $end_page -gt $total_pages ]; then
                    end_page=$total_pages
                fi

                # For each even page in the marking section, apply the folio marking
                # According to the updated requirements in sections_mark.md, only apply to back pages (even pages)
                # For the first section with 8 folios:
                # -p 2 => NN = 01, x = 10 (folio 01 in section)
                # -p 4 => NN = 02, x = 10 (folio 02 in section)
                # ...
                # -p 16 => NN = 08, x = 10 (folio 08 in section)

                # Find the first even page in this section
                local first_even_page=$start_page
                if [ $((first_even_page % 2)) -eq 1 ]; then
                    # If start page is odd, first even page is the next one
                    first_even_page=$((start_page + 1))
                fi

                local folio_num=1
                local even_page=$first_even_page
                while [ $even_page -le $end_page ] && [ $folio_num -le $folios_per_marking_section ] && [ $even_page -le $total_pages ]; do
                    # Use the folio number within the marking section for the stamp
                    local folio_number_in_section=$(printf "%02d" $folio_num)

                    echo "Adding folio marking $folio_number_in_section to page $even_page at position: $section_position (marking section $section_num)"

                    # Apply the section marking stamp to the specific page using the updated command format from sections_mark.md
                    # pdfcpu stamp add -p Z -mode text -- "NN"  "fontname:BigBlueTermPlusNFM ,pos:l ,offset:X 0, points:2,scale:0.03, fillc:#000000,backgroundc:#808080, rot:90, opacity: 0.4" booklet.pdf
                    # Replace 'NN' with the folio number within the section, 'X' with the calculated section_position, and 'Z' with the even page number
                    if pdfcpu stamp add -p "$even_page" -mode text -- "$folio_number_in_section" "fontname:BigBlueTermPlusNFM,pos:l,offset:$section_position 0,points:2,scale:0.03,fillc:#000000,backgroundc:#808080,rot:90,opacity:0.4" "$input_pdf" 2>/dev/null; then
                            echo "  Folio marking $folio_number_in_section added successfully on page $even_page at position $section_position (marking section $section_num)"
                    else
                            echo "  Warning: Could not add folio marking $folio_number_in_section on page $even_page at position $section_position (marking section $section_num)"
                    fi

                    # Increment folio number and move to next even page
                    ((folio_num++))
                    even_page=$((even_page + 2))
                done
        done

        echo "All section markings added successfully to $input_pdf"
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

        # Add stations (sewing points) to the booklet based on the x-up format
        # This modifies the main booklet.pdf file directly after it's created
        add_stations_direct "$OUT" "$PAGES_PER_SHEET"
        echo "Stations added to final booklet: $OUT"

        # Add section marking to the booklet based on the number of sections
        # This modifies the main booklet.pdf file directly after stations are added
        add_section_marking "$OUT" "$NSECTIONS" "$PAGES_PER_SHEET"
        echo "Section markings added to final booklet: $OUT"

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

        # reversing back pages for direct printing in place[apply the changes in place]
        local back_pages_count=$(pdfcpu info "$temp_back" 2>/dev/null | grep "Page count:" | awk '{print $3}')
        local back_pages=$(generate_sequence "$back_pages_count" -1 1)
        pdfcpu collect -q -p "$back_pages" "$temp_back"

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

        # reversing back pages for direct printing in place[apply the changes in place]
        local back_pages_count=$(pdfcpu info "$temp_back" 2>/dev/null | grep "Page count:" | awk '{print $3}')
        local back_pages=$(generate_sequence "$back_pages_count" -1 1)
        pdfcpu collect -q -p "$back_pages" "$temp_back"

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

        # reversing back pages for direct printing in place[apply the changes in place]
        local back_pages_count=$(pdfcpu info "$temp_back" 2>/dev/null | grep "Page count:" | awk '{print $3}')
        local back_pages=$(generate_sequence "$back_pages_count" -1 1)
        pdfcpu collect -q -p "$back_pages" "$temp_back"

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
        # Check and install required fonts
        check_and_install_fonts

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
