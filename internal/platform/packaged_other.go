//go:build !windows

package platform

// Packaged reports whether AnyFile runs from an MSIX package, which
// only Windows has.
func Packaged() bool { return false }
