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

echo "=== Booklet Maker ==="
echo "Quick presets:"
echo "0) Standard RTL booklet (1-up) with blank pages"
echo "1) Standard LTR booklet (1-up) with blank pages"
echo "2) Standard RTL booklet (2-up) with blank pages"
echo "3) Standard LTR booklet (2-up) with blank pages"
echo "4) Compact RTL (4-up) with blank pages"
echo "5) Compact LTR (4-up) with blank pages"
echo "6) Custom settings"
echo "7) Add page numbers to existing PDF"
echo "8) Split PDF into volumes with cover pages"
echo ""

# Check and install required fonts
check_and_install_fonts

read -p "Choose option [0-8]: " choice

# Function to select PDF file
select_pdf_file() {
        local pdf_files=()
        local i=0
        for file in *.pdf; do
                if [[ -f "$file" ]]; then
                        pdf_files+=("$file")
                        ((i++))
                fi
        done

        if [[ ${#pdf_files[@]} -eq 0 ]]; then
                echo "No PDF files found in current directory!" >&2
                return 1
        fi

        # Display the PDF files before asking for selection (to stderr to avoid capture)
        echo "Available PDF files:" >&2
        for j in "${!pdf_files[@]}"; do
                echo "$j) ${pdf_files[$j]}" >&2
        done

        read -r -p "Select PDF file by number [0]: " pdf_choice
        pdf_choice=${pdf_choice:-0}

        # Validate the selection and return the selected file
        if [[ $pdf_choice -ge 0 && $pdf_choice -lt ${#pdf_files[@]} ]]; then
                echo "${pdf_files[$pdf_choice]}"
        else
                # If invalid selection, return first PDF
                echo "${pdf_files[0]}"
        fi
}

# Function to add page numbers starting from page 2 (numbered as 1)
add_page_numbers() {
        local input_pdf="$1"
        local output_pdf="$2"

        echo "Adding page numbers to $input_pdf..."
        echo "Output will be saved as $output_pdf"
        echo "Page numbering will start from page 2 (numbered as 1), skipping the first page (cover)."

        local total_pages=$(pdfcpu info "$input_pdf" 2>/dev/null | grep "Page count:" | awk '{print $3}')

        if [ -z "$total_pages" ] || [ "$total_pages" -lt 1 ]; then
                echo "Error: Could not read page count from PDF or PDF has no pages."
                return 1
        fi

        if [ "$total_pages" -eq 1 ]; then
                # If there's only one page, just copy it without numbering
                cp "$input_pdf" "$output_pdf"
                echo "PDF has only one page, copying without numbering."
                return 0
        fi

        # Create temporary files
        local temp_numbered="${output_pdf%.pdf}_numbered_temp.pdf"
        local temp_cover="${output_pdf%.pdf}_cover_temp.pdf"
        local temp_numbered_final="${output_pdf%.pdf}_numbered_final.pdf"

        # Extract the first page (cover) to a temporary file
        pdfcpu collect -q -p 1 "$input_pdf" "$temp_cover"

        # Extract pages 2 to end to another temporary file
        if [ "$total_pages" -gt 1 ]; then
                local remaining_pages=$(seq -s "," 2 "$total_pages")
                pdfcpu collect -q -p "$remaining_pages" "$input_pdf" "$temp_numbered"

                # Add page numbers to the remaining pages
                # The %p will number them as 1, 2, 3... based on their position in this extracted document
                pdfcpu stamp add -mode text "%p" "scale:1.0 abs, pos:bc, rot:0" "$temp_numbered" "$temp_numbered_final"

                # Now combine the cover page with the numbered pages using pdfcpu merge
                pdfcpu merge "$output_pdf" "$temp_cover" "$temp_numbered_final"

                # Clean up temporary files
                rm -f "$temp_cover" "$temp_numbered" "$temp_numbered_final"
        fi

        echo "Page numbering completed. Output saved as $output_pdf"
}

# Function to add page numbers to all pages except the first (cover) page
number_volume_pages() {
        local input_pdf="$1"
        local output_pdf="$2"

        echo "Adding page numbers to $input_pdf..."
        echo "Output will be saved as $output_pdf"
        echo "Page numbering will start from page 2 (numbered as 1), skipping the first page (cover)."

        local total_pages=$(pdfcpu info "$input_pdf" 2>/dev/null | grep "Page count:" | awk '{print $3}')

        if [ -z "$total_pages" ] || [ "$total_pages" -lt 1 ]; then
                echo "Error: Could not read page count from PDF or PDF has no pages."
                return 1
        fi

        if [ "$total_pages" -eq 1 ]; then
                # If there's only one page, just copy it without numbering
                cp "$input_pdf" "$output_pdf"
                echo "PDF has only one page, copying without numbering."
                return 0
        fi

        # Create temporary files
        local temp_numbered="${output_pdf%.pdf}_numbered_temp.pdf"
        local temp_cover="${output_pdf%.pdf}_cover_temp.pdf"
        local temp_numbered_final="${output_pdf%.pdf}_numbered_final.pdf"

        # Extract the first page (cover) to a temporary file
        pdfcpu collect -q -p 1 "$input_pdf" "$temp_cover"

        # Extract pages 2 to end to another temporary file
        if [ "$total_pages" -gt 1 ]; then
                local remaining_pages=$(seq -s "," 2 "$total_pages")
                pdfcpu collect -q -p "$remaining_pages" "$input_pdf" "$temp_numbered"

                # Add page numbers to the remaining pages
                # The %p will number them as 1, 2, 3... based on their position in this extracted document
                pdfcpu stamp add -mode text "%p" "fontname:BigBlueTermPlusNFM,scale:1.0 abs, pos:bc, rot:0" "$temp_numbered" "$temp_numbered_final"

                # Now combine the cover page with the numbered pages using pdfcpu merge
                pdfcpu merge "$output_pdf" "$temp_cover" "$temp_numbered_final"

                # Clean up temporary files
                rm -f "$temp_cover" "$temp_numbered" "$temp_numbered_final"
        fi

        echo "Page numbering completed. Output saved as $output_pdf"
}

# Function to split PDF into volumes with cover pages and volume numbering
split_into_volumes() {
        local input_pdf="$1"
        local num_volumes="$2"
        local page_ranges=("${@:3}") # Array of page ranges

        # Extract the base name of the input PDF (without extension)
        local input_base_name=$(basename "${input_pdf%.pdf}")

        echo "Splitting $input_pdf into $num_volumes volumes..."
        echo "Page ranges: ${page_ranges[*]}"

        # Get total pages in the original PDF
        local total_pages=$(pdfcpu info "$input_pdf" 2>/dev/null | grep "Page count:" | awk '{print $3}')

        if [ -z "$total_pages" ] || [ "$total_pages" -lt 1 ]; then
                echo "Error: Could not read page count from PDF or PDF has no pages."
                return 1
        fi

        echo "Total pages in original PDF: $total_pages"

        # Extract the first page (cover) to use for all volumes
        local temp_cover="temp_cover.pdf"
        pdfcpu collect -q -p 1 "$input_pdf" "$temp_cover"

        # Process each volume
        for ((i = 0; i < num_volumes; i++)); do
                local vol_num=$((i + 1))
                # Format volume number with leading zeros (e.g., 01, 02, etc.)
                local vol_num_formatted=$(printf "%02d" $vol_num)
                local vol_name="${input_base_name}_volume_${vol_num_formatted}.pdf"

                echo "Creating volume $vol_num: $vol_name"

                # Determine the page range for this volume
                local start_page
                local end_page

                if [ $i -eq 0 ]; then
                        # First volume: starts from page 2 (after cover) to the specified end page
                        start_page=2
                        end_page=${page_ranges[0]}

                        # Check if user entered 'l' for the last page
                        if [ "$end_page" = "l" ] || [ "$end_page" = "L" ]; then
                                end_page=$total_pages
                        fi
                else
                        # Subsequent volumes: start from the page after the previous volume's end
                        start_page=$((page_ranges[i - 1] + 1))
                        end_page=${page_ranges[i]}

                        # Check if user entered 'l' for the last page (only valid for the last volume)
                        if ([ "$end_page" = "l" ] || [ "$end_page" = "L" ]) && [ $i -eq $((num_volumes - 1)) ]; then
                                end_page=$total_pages
                        fi
                fi

                echo "  Page range: $start_page to $end_page"

                # Validate page range
                if [ "$start_page" -gt "$total_pages" ] || [ "$end_page" -gt "$total_pages" ]; then
                        echo "  Warning: Page range exceeds total pages in PDF. Adjusting..."
                        if [ "$start_page" -gt "$total_pages" ]; then
                                echo "  Warning: No more pages to extract for volume $vol_num"
                                continue
                        fi
                        if [ "$end_page" -gt "$total_pages" ]; then
                                end_page=$total_pages
                        fi
                fi

                # Extract pages for this volume
                local temp_volume="temp_volume_${vol_num}.pdf"

                if [ "$start_page" -le "$end_page" ]; then
                        local page_sequence=$(seq -s "," $start_page $end_page)
                        pdfcpu collect -q -p "$page_sequence" "$input_pdf" "$temp_volume"

                        # Add volume number to the cover page
                        local temp_cover_with_vol="temp_cover_vol_${vol_num}.pdf"
                        cp "$temp_cover" "$temp_cover_with_vol"

                        # Add volume number stamp to the cover page using the exact command format
                        pdfcpu stamp add -p 1 -mode text -- "Vol. $vol_num" "fontname:BigBlueTermPlusNFM,pos:bc,ma:10 270,offset:0 20,points:60,fillc:#000000,bgcol:#808080,rot:0,opacity:0.8" "$temp_cover_with_vol"

                        # Merge the cover with volume pages
                        pdfcpu merge "$vol_name" "$temp_cover_with_vol" "$temp_volume"

                        # Clean up temporary files
                        rm -f "$temp_volume" "$temp_cover_with_vol"

                        # Add page numbering to the volume (excluding the cover page)
                        local numbered_vol_name="numbered_${vol_name}"
                        number_volume_pages "$vol_name" "$numbered_vol_name"

                        # Replace the original volume with the numbered one
                        mv "$numbered_vol_name" "$vol_name"

                        echo "  Volume $vol_num created: $vol_name (with page numbering)"
                else
                        echo "  Warning: Invalid page range for volume $vol_num (start: $start_page, end: $end_page)"
                fi
        done

        # Clean up the temporary cover file
        rm -f "$temp_cover"

        echo "Volume splitting completed!"
}

case $choice in
0)
        selected_pdf=$(select_pdf_file)
        if [[ -n "$selected_pdf" ]]; then
                ./bookit.sh "$selected_pdf" "booklet.pdf" 1 RTL 8 1
        else
                echo "No PDF file selected, exiting."
                exit 1
        fi
        ;;
1)
        selected_pdf=$(select_pdf_file)
        if [[ -n "$selected_pdf" ]]; then
                ./bookit.sh "$selected_pdf" "booklet.pdf" 1 LTR 8 1
        else
                echo "No PDF file selected, exiting."
                exit 1
        fi
        ;;
2)
        selected_pdf=$(select_pdf_file)
        if [[ -n "$selected_pdf" ]]; then
                ./bookit.sh "$selected_pdf" "booklet.pdf" 2 RTL 8 1
        else
                echo "No PDF file selected, exiting."
                exit 1
        fi
        ;;
3)
        selected_pdf=$(select_pdf_file)
        if [[ -n "$selected_pdf" ]]; then
                ./bookit.sh "$selected_pdf" "booklet.pdf" 2 LTR 8 1
        else
                echo "No PDF file selected, exiting."
                exit 1
        fi
        ;;
4)
        selected_pdf=$(select_pdf_file)
        if [[ -n "$selected_pdf" ]]; then
                ./bookit.sh "$selected_pdf" "booklet.pdf" 4 RTL 8 1
        else
                echo "No PDF file selected, exiting."
                exit 1
        fi
        ;;
5)
        selected_pdf=$(select_pdf_file)
        if [[ -n "$selected_pdf" ]]; then
                ./bookit.sh "$selected_pdf" "booklet.pdf" 4 LTR 8 1
        else
                echo "No PDF file selected, exiting."
                exit 1
        fi
        ;;
6)
        selected_pdf=$(select_pdf_file)
        if [[ -n "$selected_pdf" ]]; then

                read -p "Output PDF [booklet.pdf]: " out
                read -p "Pages per sheet (1|2|4|8) [2]: " pps
                read -p "Reading direction (RTL|LTR) [LTR]: " direction
                read -p "Sections [8]: " sections
                read -p "Add blank pages to front and back ?(0|1) [1]: " blank

                ./bookit.sh "$selected_pdf" "${out:-booklet.pdf}" "${pps:-2}" "${direction:-LTR}" "${sections:-8}" "${blank:-1}"
        else
                echo "No PDF file selected, exiting."
                exit 1
        fi

        ;;
7)
        selected_pdf=$(select_pdf_file)
        if [[ -n "$selected_pdf" ]]; then
                read -p "Output PDF name [numbered_$(basename "$selected_pdf")]: " output_pdf
                output_pdf="${output_pdf:-numbered_$(basename "$selected_pdf")}"
                add_page_numbers "$selected_pdf" "$output_pdf"
        else
                echo "No PDF file selected, exiting."
                exit 1
        fi
        ;;
8)
        selected_pdf=$(select_pdf_file)
        if [[ -n "$selected_pdf" ]]; then
                read -p "Number of volumes to create: " num_volumes

                # Validate number of volumes
                if ! [[ "$num_volumes" =~ ^[0-9]+$ ]] || [ "$num_volumes" -lt 1 ]; then
                        echo "Error: Please enter a valid number of volumes (at least 1)"
                        exit 1
                fi

                # Get total pages in the original PDF to validate 'l' input
                total_pages=$(pdfcpu info "$selected_pdf" 2>/dev/null | grep "Page count:" | awk '{print $3}')
                if [ -z "$total_pages" ] || [ "$total_pages" -lt 1 ]; then
                        echo "Error: Could not read page count from PDF or PDF has no pages."
                        exit 1
                fi

                # Get page ranges for each volume
                page_ranges=()
                for ((i = 0; i < num_volumes; i++)); do
                        vol_num=$((i + 1))
                        read -p "Last page for volume $vol_num (use 'l' for last page): " last_page

                        # Check if user entered 'l' or 'L' for the last page
                        if [ "$last_page" = "l" ] || [ "$last_page" = "L" ]; then
                                # Only allow 'l' for the last volume
                                if [ $i -eq $((num_volumes - 1)) ]; then
                                        page_ranges+=("$last_page")
                                else
                                        echo "Error: 'l' can only be used for the last volume"
                                        exit 1
                                fi
                        elif ! [[ "$last_page" =~ ^[0-9]+$ ]] || [ "$last_page" -lt 1 ]; then
                                echo "Error: Please enter a valid page number or 'l' for the last volume"
                                exit 1
                        else
                                # Validate that the page number doesn't exceed total pages
                                if [ "$last_page" -gt "$total_pages" ]; then
                                        echo "Error: Page number ($last_page) exceeds total pages in PDF ($total_pages)"
                                        exit 1
                                fi
                                page_ranges+=("$last_page")
                        fi
                done

                # Call the split function
                split_into_volumes "$selected_pdf" "$num_volumes" "${page_ranges[@]}"
        else
                echo "No PDF file selected, exiting."
                exit 1
        fi
        ;;
*) echo "Invalid option" ;;
esac
