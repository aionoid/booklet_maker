package main

import (
	//"io"
	// "time"
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/log"
	// "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	// "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func BookletFile(inFiles []string, outFile string, selectedPages []string, nup *model.NUp, conf *model.Configuration) (err error) {

	var f1, f2 *os.File

	// booklet from a PDF
	if f1, err = os.Open(inFiles[0]); err != nil {
		return err
	}

	if f2, err = os.Create(outFile); err != nil {
		f1.Close()
		return err
	}
	log.CLI.Printf("writing %s...\n", outFile)

	defer func() {
		if err != nil {
			f2.Close()
			f1.Close()
			return
		}
		if err = f2.Close(); err != nil {
			return
		}
		err = f1.Close()
	}()

	err = api.Booklet(f1, f2, inFiles, selectedPages, nup, conf)
	// return api.Booklet(f1, f2, inFiles, selectedPages, nup, conf)

	return err
}

// FOR ONE SECTION
// pdfcpu pages insert -p 1 -m before book_.pdf out.pdf
// pdfcpu pages insert -p 1 -m before out.pdf
// pdfcpu pages insert -p l -m after out.pdf
// pdfcpu pages insert -p l -m after out.pdf
// pdfcpu split book_.pdf out 32
// pdfcpu booklet -- g:on, ma:10, bgcol:#beded9 booklet.pdf 2 out.pdf
// pdfcpu collect -p 1,3,5,7,9,11,13,15,17,19 booklet.pdf book_f_1.pdf
// pdfcpu collect -p 2,4,6,8,10,12,14,16,18,20 booklet.pdf book_b_1.pdf

func main() {
	fmt.Printf("hello")
	log.CLI.Printf("writing %s...\n")
}
