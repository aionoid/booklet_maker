#!/usr/bin/env bash

echo "=== Booklet Maker ==="
echo "Quick presets:"
echo "1) Standard RTL booklet (2-up)"
echo "2) Standard LTR booklet (2-up)"
echo "3) Compact RTL (4-up)"
echo "4) Compact LTR (4-up)"
echo "5) Custom settings"
echo ""

read -p "Choose option [1-5]: " choice

case $choice in
1) ./bookit.sh "book.pdf" "booklet.pdf" 2 RTL 8 ;;
2) ./bookit.sh "book.pdf" "booklet.pdf" 2 LTR 8 ;;
3) ./bookit.sh "book.pdf" "booklet.pdf" 4 RTL 8 ;;
4) ./bookit.sh "book.pdf" "booklet.pdf" 4 LTR 8 ;;
5)
        read -p "Input PDF [book.pdf]: " book
        read -p "Output PDF [booklet.pdf]: " out
        read -p "Pages per sheet (2|4|8) [2]: " pps
        read -p "Reading direction (RTL|LTR) [RTL]: " direction
        read -p "Sections [8]: " sections

        ./bookit.sh "${book:-book.pdf}" "${out:-booklet.pdf}" "${pps:-2}" "${direction:-RTL}" "${sections:-8}"
        ;;
*) echo "Invalid option" ;;
esac
