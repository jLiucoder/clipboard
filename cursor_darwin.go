package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

static void mousePosAndScreen(int *mx, int *my, int *sw, int *sh, int *sc) {
    NSPoint p = [NSEvent mouseLocation];
    NSScreen *screen = [NSScreen mainScreen];
    CGFloat scale = [screen backingScaleFactor];
    // All values in scaled pixels (what Wails SetPosition expects).
    *mx = (int)(p.x * scale);
    *my = (int)((screen.frame.size.height - p.y) * scale);
    *sw = (int)(screen.frame.size.width * scale);
    *sh = (int)(screen.frame.size.height * scale);
    *sc = (int)scale;
}
*/
import "C"

// cursorAndScreen returns the mouse position, screen size, and scale factor,
// all in the same scaled-pixel coordinate space used by Wails SetPosition.
func cursorAndScreen() (cx, cy, sw, sh, scale int) {
	var mx, my, screenW, screenH, sc C.int
	C.mousePosAndScreen(&mx, &my, &screenW, &screenH, &sc)
	return int(mx), int(my), int(screenW), int(screenH), int(sc)
}
