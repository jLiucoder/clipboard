package main

import "testing"

func TestCalcWindowPosition(t *testing.T) {
	// Simulated screen: 2880x1800 (Retina 2x of 1440x900)
	// Window: 760x840 (380x420 * 2)
	const sw, sh = 2880, 1800
	const ww, wh = 760, 840

	tests := []struct {
		name       string
		cx, cy     int
		wantX      int
		wantY      int
	}{
		{
			name:  "normal — top-left at cursor",
			cx:    500, cy: 200,
			wantX: 500, wantY: 200,
		},
		{
			name:  "right edge — flip horizontal",
			cx:    2800, cy: 200,
			wantX: 2800 - ww, wantY: 200,
		},
		{
			name:  "bottom edge — flip vertical",
			cx:    500, cy: 1500,
			wantX: 500, wantY: 1500 - wh,
		},
		{
			name:  "bottom-right corner — flip both",
			cx:    2800, cy: 1500,
			wantX: 2800 - ww, wantY: 1500 - wh,
		},
		{
			name:  "top-left corner — no flip, no clamp needed",
			cx:    0, cy: 0,
			wantX: 0, wantY: 0,
		},
		{
			name:  "exact fit right edge — no flip needed",
			cx:    sw - ww, cy: 100,
			wantX: sw - ww, wantY: 100,
		},
		{
			name:  "one pixel over right — flips",
			cx:    sw - ww + 1, cy: 100,
			wantX: sw - ww + 1 - ww, wantY: 100,
		},
		{
			name:  "exact fit bottom edge — no flip needed",
			cx:    100, cy: sh - wh,
			wantX: 100, wantY: sh - wh,
		},
		{
			name:  "one pixel over bottom — flips",
			cx:    100, cy: sh - wh + 1,
			wantX: 100, wantY: sh - wh + 1 - wh,
		},
		{
			name:  "cursor near top-right — flip horizontal, clamp prevents negative Y",
			cx:    2800, cy: 10,
			wantX: 2800 - ww, wantY: 10,
		},
		{
			name:  "cursor near bottom-left — flip vertical, clamp prevents negative X",
			cx:    10, cy: 1500,
			wantX: 10, wantY: 1500 - wh,
		},
		{
			name:  "tiny screen where flip goes negative — clamped to 0",
			cx:    100, cy: 100,
			// Window bigger than screen half, flip goes negative
			wantX: 0, wantY: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screenW, screenH := sw, sh
			winW, winH := ww, wh
			// Special case: override for "tiny screen" test
			if tt.name == "tiny screen where flip goes negative — clamped to 0" {
				screenW, screenH = 150, 150
				winW, winH = 200, 200
			}

			gotX, gotY := calcWindowPosition(tt.cx, tt.cy, winW, winH, screenW, screenH)
			if gotX != tt.wantX || gotY != tt.wantY {
				t.Errorf("calcWindowPosition(%d, %d, %d, %d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.cx, tt.cy, winW, winH, screenW, screenH,
					gotX, gotY, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestCalcWindowPosition_AlwaysInBounds(t *testing.T) {
	const sw, sh = 2880, 1800
	const ww, wh = 760, 840

	// Sweep cursor across the entire screen and verify result is always in bounds.
	for cx := 0; cx <= sw; cx += 100 {
		for cy := 0; cy <= sh; cy += 100 {
			wx, wy := calcWindowPosition(cx, cy, ww, wh, sw, sh)
			if wx < 0 || wy < 0 || wx+ww > sw || wy+wh > sh {
				t.Errorf("out of bounds: cursor=(%d,%d) → window=(%d,%d), window end=(%d,%d), screen=(%d,%d)",
					cx, cy, wx, wy, wx+ww, wy+wh, sw, sh)
			}
		}
	}
}
