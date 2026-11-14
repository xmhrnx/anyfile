//go:build darwin

package platform

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit

#include <stdlib.h>
#import <AppKit/AppKit.h>

// anyfileOpen opens a file as double-clicking it in Finder would.
static int anyfileOpen(const char *path) {
	@autoreleasepool {
		NSURL *u = [NSURL fileURLWithPath:@(path)];
		return [[NSWorkspace sharedWorkspace] openURL:u] ? 1 : 0;
	}
}

// anyfileReveal shows a file selected in a Finder window.
static void anyfileReveal(const char *path) {
	@autoreleasepool {
		NSURL *u = [NSURL fileURLWithPath:@(path)];
		[[NSWorkspace sharedWorkspace] activateFileViewerSelectingURLs:@[u]];
	}
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

// NSWorkspace rather than running open(1): the App Sandbox of Mac App
// Store builds allows it, and Launch Services then lets the viewer read
// the file.

// Open launches the user's default application for the file.
func Open(path string) error {
	c := C.CString(path)
	defer C.free(unsafe.Pointer(c))
	if C.anyfileOpen(c) == 0 {
		return errors.New("no application opens " + path)
	}
	return nil
}

// OpenWith shows the file in Finder, the closest macOS has to a
// chooser: its Open With menu lists the applications.
func OpenWith(path string) error { return Reveal(path) }

// Reveal opens the file manager at the file.
func Reveal(path string) error {
	c := C.CString(path)
	defer C.free(unsafe.Pointer(c))
	C.anyfileReveal(c)
	return nil
}

// CanOpenWith reports whether OpenWith shows a chooser here.
func CanOpenWith() bool { return false }
