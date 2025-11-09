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

case $choice in
0) ./bookit.sh "book.pdf" "booklet.pdf" 1 RTL 8 1 ;;
1) ./bookit.sh "book.pdf" "booklet.pdf" 1 LTR 8 1 ;;
2) ./bookit.sh "book.pdf" "booklet.pdf" 2 RTL 8 1 ;;
3) ./bookit.sh "book.pdf" "booklet.pdf" 2 LTR 8 1 ;;
4) ./bookit.sh "book.pdf" "booklet.pdf" 4 RTL 8 1 ;;
5) ./bookit.sh "book.pdf" "booklet.pdf" 4 LTR 8 1 ;;
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
