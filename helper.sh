#!/usr/bin/env bash

echo "=== Booklet Maker ==="
echo "Quick presets:"
echo "0) Standard RTL booklet (1-up)"
echo "1) Standard LTR booklet (1-up)"
echo "2) Standard RTL booklet (2-up)"
echo "3) Standard LTR booklet (2-up)"
echo "4) Compact RTL (4-up)"
echo "5) Compact LTR (4-up)"
echo "6) Custom settings"
echo ""

read -p "Choose option [0-6]: " choice

case $choice in
0) ./bookit.sh "book.pdf" "booklet.pdf" 1 RTL 4 ;;
1) ./bookit.sh "book.pdf" "booklet.pdf" 1 LTR 4 ;;
2) ./bookit.sh "book.pdf" "booklet.pdf" 2 RTL 4 ;;
3) ./bookit.sh "book.pdf" "booklet.pdf" 2 LTR 4 ;;
4) ./bookit.sh "book.pdf" "booklet.pdf" 4 RTL 4 ;;
5) ./bookit.sh "book.pdf" "booklet.pdf" 4 LTR 4 ;;
6)
        read -p "Input PDF [book.pdf]: " book
        read -p "Output PDF [booklet.pdf]: " out
        read -p "Pages per sheet (1|2|4|8) [2]: " pps
        read -p "Reading direction (RTL|LTR) [LTR]: " direction
        read -p "Sections [4]: " sections

        ./bookit.sh "${book:-book.pdf}" "${out:-booklet.pdf}" "${pps:-2}" "${direction:-LTR}" "${sections:-4}"
        ;;
*) echo "Invalid option" ;;
esac
