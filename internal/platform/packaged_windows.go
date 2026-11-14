//go:build windows

package platform

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var procGetCurrentPackageFullName = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetCurrentPackageFullName")

// Packaged reports whether AnyFile runs from an MSIX package, as
// Microsoft Store installs do. The package declares the right-click
// menu entry itself; registry writes from inside it never reach
// Explorer.
func Packaged() bool {
	if procGetCurrentPackageFullName.Find() != nil {
		return false // before Windows 8
	}
	var n uint32
	r, _, _ := procGetCurrentPackageFullName.Call(uintptr(unsafe.Pointer(&n)), 0)
	return windows.Errno(r) == windows.ERROR_INSUFFICIENT_BUFFER
}
