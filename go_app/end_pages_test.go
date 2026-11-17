package main

import "testing"

// TestComputeEndPages verifies the per-volume end-page math used by the Prepare
// tab in the GUI (and by SplitIntoVolumes). It covers the even-split fallback,
// explicit ends, the "to the end" ('l'/0) last-volume case, clamping, and all
// validation errors.
func TestComputeEndPages(t *testing.T) {
	type tc struct {
		name         string
		contentStart int
		contentEnd   int
		n            int
		explicit     []int // nil = even split

		want    []int
		wantErr bool
	}
	tests := []tc{
		{"even split: 10 content pages, 4 volumes", 2, 11, 4, nil, []int{4, 7, 9, 11}, false},
		{"even split: 9 content pages, 4 volumes (uneven)", 2, 10, 4, nil, []int{4, 6, 8, 10}, false},
		{"explicit ends", 2, 11, 3, []int{5, 8, 11}, []int{5, 8, 11}, false},
		{"explicit ends with 'l' (0) on last volume", 2, 10, 2, []int{4, 0}, []int{4, 10}, false},
		{"clamp high end to contentEnd", 2, 10, 2, []int{3, 15}, []int{3, 10}, false},
		{"non-increasing ends rejected", 2, 10, 2, []int{8, 5}, nil, true},
		{"equal ends rejected", 2, 10, 2, []int{5, 5}, nil, true},
		{"'l' on a non-last volume rejected", 2, 10, 3, []int{0, 5, 10}, nil, true},
		{"end before contentStart rejected", 2, 10, 2, []int{1, 10}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := computeEndPages(tt.contentStart, tt.contentEnd, tt.n, tt.explicit)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got none (ends=%v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("want %d ends, got %d: %v", len(tt.want), len(got), got)
				return
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("ends[%d]: want %d, got %d", i, tt.want[i], got[i])
				}
			}
		})
	}
}
