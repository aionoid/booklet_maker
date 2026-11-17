package main

import (
	"strings"
	"testing"
)

// TestComputeBookletSummary verifies the live "What you'll get" math reproduces
// bookit.sh's result values: number of sections, pages per section, total
// booklet pages (with blank padding) and total A4 sheets to print.
func TestComputeBookletSummary(t *testing.T) {
	type tc struct {
		name          string
		totalPages    int // original PDF page count
		pagesPerSheet int
		sections      int
		addBlank      bool

		wantSections string // e.g. "1"
		wantPerPage  string // e.g. "16"
		wantTotal    string // e.g. "16"
		wantSheets   string // e.g. "4"
	}
	tests := []tc{
		{"10 pages, 2-up, 8 sections, blanks", 10, 2, 8, true, "1", "16", "16", "4"},
		{"1 page, 2-up, 8 sections, blanks", 1, 2, 8, true, "1", "16", "16", "4"},
		{"12 pages, 4-up, 8 sections, no blanks (partial last section)", 12, 4, 8, false, "1", "32", "12", "2"},
		{"30 pages, 2-up, 8 sections, blanks (multiple sections)", 30, 2, 8, true, "3", "16", "48", "12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blank := 0
			if tt.addBlank {
				blank = 1
			}
			lines, err := computeBookletSummary(tt.totalPages, tt.pagesPerSheet, tt.sections, blank)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			out := strings.Join(lines, "\n")

			wantSections := "Sections (signatures): " + tt.wantSections
			wantPerPage := "Pages per section: " + tt.wantPerPage
			wantTotal := "Total booklet pages: " + tt.wantTotal
			wantSheets := "Total A4 sheets to print: " + tt.wantSheets

			for _, w := range []string{wantSections, wantPerPage, wantTotal, wantSheets} {
				if !strings.Contains(out, w) {
					t.Errorf("expected %q in summary, got:\n%s", w, out)
				}
			}
		})
	}
}
