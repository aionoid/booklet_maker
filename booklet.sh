BOOK="book.pdf"
OUT="booklet.pdf"
NSECTIONS=8
DIR_SECTION=sections
DIR_READY=print_ready

# CLEAN
toilet -t -f banner3-D "1: clean"
rm $OUT
rm -rd $DIR_SECTION
rm -rd $DIR_READY

# PREPAIR
toilet -t -f banner3-D "2: prepair"
mkdir $DIR_SECTION
mkdir $DIR_READY

# make a booklet
toilet -t -f banner3-D "3: make a booklet"
pdfcpu booklet -- "or:rd, p:A4, multifolio:on, foliosize:$NSECTIONS, g:on" $OUT 2 $BOOK
#pdfcpu booklet -- "p:A4, multifolio:on, foliosize:$NSECTIONS, g:on, ma:10, bgcol:#beded9 " $OUT 2 $BOOK

# split for each section
toilet -t -f banner3-D "4: split for $DIR_SECTION p sections"
pdfcpu split -q $OUT $DIR_SECTION $(($NSECTIONS * 2))

# collect for each section Front/Back pages
toilet -t -f banner3-D "5: export Front/Back pages"
IND=0
for FILE in $(ls -1tr $DIR_SECTION/*.pdf); do
	echo $FILE
	((IND += 1))
	FOUT=$(echo $FILE | sed "s/$DIR_SECTION\///g")
	#echo $FOUT
	pdfcpu collect -q -p $(seq -s "," 1 2 $(($NSECTIONS * 2))) $FILE $DIR_READY/$IND"_F_"$FOUT
	# reverse order for printing directly
	pdfcpu collect -q -p $(seq -s "," $(($NSECTIONS * 2)) -2 2) $FILE $DIR_READY/$IND"_B_"$FOUT

	# pdfcpu collect -p 1,3,5,7,9,11,13,15,17,19 $FILE $DIR_READY/$FOUT"_F.pdf"
	# pdfcpu collect -p 2,4,6,8,10,12,14,16,18,20 $FILE $DIR_READY/$FOUT"_B.pdf"
done

# pdfcpu pages insert -p 1 -m before book_.pdf out.pdf
# pdfcpu pages insert -p 1 -m before out.pdf
# pdfcpu pages insert -p l -m after out.pdf
# pdfcpu pages insert -p l -m after out.pdf
# pdfcpu booklet -- "g:on, ma:10, bgcol:#beded9" booklet.pdf 2 out.pdf
# pdfcpu collect -p 1,3,5,7,9,11,13,15,17,19 booklet.pdf book_f_1.pdf
# pdfcpu collect -p 2,4,6,8,10,12,14,16,18,20 booklet.pdf book_b_1.pdf
# for doing a booklet with sections with 8 papaers
#pdfcpu booklet -- "p:A4, multifolio:on, foliosize:8" hardbackbook.pdf 2 book_.pdf
# for spliting pdf
#pdfcpu split out.pdf out 32
