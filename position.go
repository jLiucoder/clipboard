package main

// calcWindowPosition computes the window origin so that the panel appears
// at the cursor but flips to stay within screen bounds.
//
// All arguments must be in the same coordinate space (scaled pixels).
//   - cx, cy: cursor position
//   - ww, wh: window size
//   - sw, sh: screen size
//
// Default: top-left corner at cursor.
// If the window would overflow right, it flips so the top-right is at the cursor.
// If the window would overflow bottom, it flips so the bottom edge is at the cursor.
// A final clamp ensures the result is never negative.
func calcWindowPosition(cx, cy, ww, wh, sw, sh int) (wx, wy int) {
	wx = cx
	wy = cy

	if cx+ww > sw {
		wx = cx - ww
	}
	if cy+wh > sh {
		wy = cy - wh
	}

	if wx < 0 {
		wx = 0
	}
	if wy < 0 {
		wy = 0
	}
	return wx, wy
}
