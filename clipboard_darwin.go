package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

static int pasteboardChangeCount() {
    return (int)[[NSPasteboard generalPasteboard] changeCount];
}
*/
import "C"

// getPasteboardChangeCount returns the current change count of the general pasteboard.
// This is used to detect when the clipboard content has changed.
func getPasteboardChangeCount() int {
	return int(C.pasteboardChangeCount())
}
