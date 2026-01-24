# TODO

- [x] Add Stations (swing marks)

  The recommended layout using percentages. These percentages refer to the
  **vertical position along the spine (fold line)**, starting from the top () to
  the bottom ().

  ---

  ### Recommended Sewing Point Placements

  | Format        | Number of Points | Placement (Percentage from Top)              |
  | ------------- | ---------------- | -------------------------------------------- |
  | **A7** (4-up) | 4 Points         | "10%,36.6%,63.3%,90%"                        |
  | **A6** (2-up) | 6 Points         | "8%,24.8%,41.6%,58.4%,75.2%,92%"             |
  | **A5** (1-up) | 8 Points         | "7%,19.3%,31.6%,43.9%,56.1%,68.4%,80.7%,93%" |

  ### Why these percentages?

  - **The Margins:** For A7, I’ve set the first and last holes at from the
    edges. As the paper size gets larger (A6 and A5), the percentage decreases
    slightly ( and ) because a margin on an A5 sheet would look unnecessarily
    large.
  - **Even Distribution:** The "internal" points are spaced perfectly evenly
    between those two outer marks to ensure the tension of the thread is
    consistent.

  ### Implementation Tip

  If you are coding this, you can use a simple formula to calculate the
  coordinate for any number of points (n):

  1. Pick your **Margin** (e.g., 10%).
  2. The **Available Space** is 100−(Margin×2).
  3. The **Gap** between points is AvailableSpace/(n−1).

  **Example for A7 (4 points):**

  - Point 1: 10%
  - Point 2: 10+26.6=36.6%
  - Point 3: 36.6+26.6=63.2%
  - Point 4: 63.2+26.6=89.8% (Rounded to 90%)

  ---

  ### How to implement :

  - command to get the width of the booklet
    `pdfcpu info booklet.pdf | grep "Page si" | awk '{ print $3 }'`
  - command to stamp
    `pdfcpu stamp add -mode text -- "."  "fontname:BigBlueTermPlusNFM ,pos:l,offset:x 6, points:2,scale:0.03, fillc:#000000, rot:0" booklet.pdf`
  - we use the width value of the page as 100% then we will calcualte the points
    to put as x in "offset:x 6"

- [x] add sections mark with folio number

  ### Where to add it

  - use the same method as adding swing points, it means you add it after
    creating 'booklet.pdf'

  ### How it work

  - add a stamp for section marking, and it will contain the number of the folio
    in that section like defined in the commend in the next section

  ### The Command to add and how it implement

  - command
    `pdfcpu stamp add -p Z1-Z2 -mode text -- "NN"  "fontname:BigBlueTermPlusNFM ,pos:r ,offset:X 0, points:2,scale:0.03, fillc:#000000,backgroundc:#808080, rot:-90, opacity: 0.4" booklet.pdf`
    - the "NN" in the command is the number of the folio in the section, for
      example if we have sections of 8 folio, we will get the first folio 01
      then the next one 02 till the last 08
    - the "X" in "offset:x 0" is position of the stamp mark, which define the
      section and it move by "-15" so for the fist section it start from "-10" ,
      then the second section will be "-25", and the third section will be "-40"
    - the "Z" is the page number, to stamp with, so lets say we have section
      with 8 folio, the first 16 folio. the next example will the process:
      > the fist section of 8 folio, depending of Z1-Z2 the NN and X will be
      - -p 1-2 => NN = 01, x = -10
      - -p 3-4 => NN = 02, x = -10
      - ...
      - -p 15-16 => NN = 08, x = -10
      > then the second section of 8 folio, the Z1-Z2, NN and X will be:
      - -p 17-18 => NN = 01, x = -25
      - -p 19-20 => NN = 02, x = -25
      - ...
      - -p 31-32 => NN = 08, x = -25
      > then so on.

  ### Update to script

  Now only apply the section mark to the back pages which are the even pages so
  the command will be

  `pdfcpu stamp add -p Z -mode text -- "NN"  "fontname:BigBlueTermPlusNFM ,pos:l ,offset:X 0, points:2,scale:0.03, fillc:#000000,backgroundc:#808080, rot:-90, opacity: 0.4" booklet.pdf`
  where the "pos:r" will be "pos:l" and for "X" value will take positive value
  so, X will be 10 instead of -10, and 25 instead of -25, and "rot:-90" to
  "rot:90"

  > the fist section of 8 folio, depending of Z the NN and X will be

  - -p 2 => NN = 01, x = 10
  - -p 4 => NN = 02, x = 10
  - ...
  - -p 16 => NN = 08, x = 10

  > then the second section of 8 folio, the Z, NN and X will be:

  - -p 18 => NN = 01, x = 25
  - -p 20 => NN = 02, x = 25
  - ...
  - -p 32 => NN = 08, x = 25

  > then so on.

- [ ] Update the script to be CLI app in pure Go lang
- [ ] add option to run the program in GUI where you can select options you
      want, and browse the pdf file you want to work with
