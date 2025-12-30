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
echo ""

read -p "Choose option [0-6]: " choice

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
*) echo "Invalid option" ;;
esac
