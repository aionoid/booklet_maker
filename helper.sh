#!/usr/bin/env bash

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
echo ""

read -p "Choose option [0-7]: " choice

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
        read -p "Input PDF [book.pdf]: " book
        read -p "Output PDF [booklet.pdf]: " out
        read -p "Pages per sheet (1|2|4|8) [2]: " pps
        read -p "Reading direction (RTL|LTR) [LTR]: " direction
        read -p "Sections [8]: " sections
        read -p "Add blank pages to front and back ?(0|1) [1]: " blank

        ./bookit.sh "${book:-book.pdf}" "${out:-booklet.pdf}" "${pps:-2}" "${direction:-LTR}" "${sections:-8}" "${blank:-1}"

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
*) echo "Invalid option" ;;
esac
