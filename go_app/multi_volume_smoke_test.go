package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestCreateBookletWithOptionsMultiVolume exercises the full GUI-driven workflow:
// split the input into several volumes, stamp page numbers on each, and create a
// booklet for every volume in its own subfolder. It verifies the resulting
// directory structure (volumes/ raw splits + volume_01.. / with booklet.pdf,
// sections/ and print_ready/) rather than visual output.
func TestCreateBookletWithOptionsMultiVolume(t *testing.T) {
	input := findInputPDF(t)
	tmpDir, err := os.MkdirTemp("", "bookit_smoke_*")
	if err != nil {
		t.Fatalf("mkdir temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	src := filepath.Join(tmpDir, "input.pdf")
	data, err := os.ReadFile(input)
	if err != nil {
		t.Fatalf("reading input pdf: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("input pdf is empty")
	}
	if err := os.WriteFile(src, data, 0644); err != nil {
		t.Fatalf("writing input pdf: %v", err)
	}

	outDir := filepath.Join(tmpDir, "out")
	cfg := &BookletConfig{
		InputFile:        src,
		OutputFile:       filepath.Join(outDir, "booklet.pdf"),
		PagesPerSheet:    2,
		ReadingDirection: "LTR",
		Sections:         8,
		AddBlank:         1,
	}

	const numVolumes = 3
	ctx := context.Background()
	if err := createBookletWithOptions(ctx, cfg, true, numVolumes, nil); err != nil {
		t.Fatalf("createBookletWithOptions: %v", err)
	}

	for v := 1; v <= numVolumes; v++ {
		dir := filepath.Join(outDir, fmt.Sprintf("volume_%02d", v))
		if _, err := os.Stat(dir); err != nil {
			t.Fatalf("missing volume dir %s: %v", dir, err)
		}
		for _, p := range []string{"booklet.pdf", "sections", "print_ready"} {
			if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
				t.Fatalf("missing %s/%s: %v", dir, p, err)
			}
		}
		ready, _ := os.ReadDir(filepath.Join(dir, "print_ready"))
		if len(ready) == 0 {
			t.Fatalf("volume %d has an empty print_ready dir", v)
		}
	}

	if _, err := os.Stat(filepath.Join(outDir, "volumes")); err != nil {
		t.Fatalf("missing raw volume splits dir: %v", err)
	}
	t.Log("multi-volume booklet created successfully (numbering on, 3 volumes)")
}

// findInputPDF returns the path to a real multi-page PDF in the repo root so
// the smoke test works without any network access. Larger inputs are preferred
// so every split volume has content pages.
func findInputPDF(t *testing.T) string {
	candidates := []string{
		"../book.pdf",        // 200 pages
		"../sample.pdf",      // 447 pages
		"../no_nice_guy.pdf", // large
		"../simple_test.pdf", // 3 pages (last resort)
	}
	for _, c := range candidates {
		p, err := filepath.Abs(c)
		if err == nil {
			if _, e := os.Stat(p); e == nil {
				return p
			}
		}
	}
	t.Fatal("no input PDF found in repo root")
	return ""
}
