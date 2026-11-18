package main

import (
	"fmt"
	"image"
	"os"
	"testing"
	"time"

	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"
)

// TestPdfiumRenderHeadless verifies the WASM PDFium backend can open a real
// PDF from this repo, report its page count, and render a page to self-contained
// PNG bytes (no native lib required). This guards the GUI preview pipeline.
func TestPdfiumRenderHeadless(t *testing.T) {
	pool, err := webassembly.Init(webassembly.Config{MinIdle: 1, MaxIdle: 1, MaxTotal: 2})
	if err != nil {
		t.Fatalf("init pdfium pool: %v", err)
	}
	defer func() {
		if e := pool.Close(); e != nil {
			t.Logf("pool close: %v", e)
		}
	}()

	instance, err := pool.GetInstance(time.Second * 30)
	if err != nil {
		t.Fatalf("get instance: %v", err)
	}
	defer func() {
		if e := instance.Close(); e != nil {
			t.Logf("instance close: %v", e)
		}
	}()

	data, err := os.ReadFile("../book.pdf")
	if err != nil {
		t.Fatalf("reading book.pdf: %v", err)
	}
	doc, err := instance.OpenDocument(&requests.OpenDocument{File: &data})
	if err != nil {
		t.Fatalf("open document: %v", err)
	}
	defer func() {
		instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})
	}()

	pc, err := instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		t.Fatalf("page count: %v", err)
	}
	fmt.Printf("book.pdf has %d page(s)\n", pc.PageCount)
	if pc.PageCount < 2 {
		t.Fatalf("expected at least 2 pages, got %d", pc.PageCount)
	}

	resp, err := instance.RenderToFile(&requests.RenderToFile{
		RenderPageInPixels: &requests.RenderPageInPixels{
			Page: requests.Page{
				ByIndex: &requests.PageByIndex{
					Document: doc.Document,
					Index:    0,
				},
			},
			Width: 720,
		},
		OutputFormat: requests.RenderToFileOutputFormatPNG,
		OutputTarget: requests.RenderToFileOutputTargetBytes,
	})
	if err != nil {
		t.Fatalf("render page 1 to PNG: %v", err)
	}
	if resp.ImageBytes == nil || len(*resp.ImageBytes) == 0 {
		t.Fatal("render returned no image bytes")
	}

	out := "/tmp/bookit_pdfium_page1.png"
	if err := os.WriteFile(out, *resp.ImageBytes, 0644); err != nil {
		t.Fatalf("writing png: %v", err)
	}
	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("opening png: %v", err)
	}
	defer f.Close()
	img, format, _ := image.Decode(f)
	if img == nil || format == "" {
		t.Fatal("rendered output is not a valid image")
	}
	fmt.Printf("rendered page 1 -> %s (%d bytes, format=%s, dims=%dx%d)\n",
		out, len(*resp.ImageBytes), format, img.Bounds().Dx(), img.Bounds().Dy())
}
