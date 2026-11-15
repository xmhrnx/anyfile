// Package platform hands files to the operating system: opening them
// in the user's own viewers (AnyFile never renders content itself) and
// revealing them in the file manager. It also tells how AnyFile was
// installed, since the stores' packages limit what it may do.
package platform

import (
	"os"
	"runtime"
	"strings"
)

// FileDropSupported reports whether files can be dropped on the window
// here. giode takes drops on Windows and macOS, and on Linux when the
// giode CLI builds the app.
func FileDropSupported() bool {
	switch runtime.GOOS {
	case "windows", "darwin", "linux":
		return true
	}
	return false
}

// Sandboxed reports whether AnyFile runs in the macOS App Sandbox, as
// Mac App Store builds do. The sandbox only lets it reach the files the
// user hands it, and the store forbids downloading programs, so the
// external engines are off there.
func Sandboxed() bool {
	return sandboxed(runtime.GOOS, os.Getenv)
}

// sandboxed tells from the environment: macOS doesn't always set
// APP_SANDBOX_CONTAINER_ID (a TestFlight copy didn't), but a sandboxed
// app's home is always its container, ~/Library/Containers/<id>/Data.
func sandboxed(goos string, getenv func(string) string) bool {
	if goos != "darwin" {
		return false
	}
	return getenv("APP_SANDBOX_CONTAINER_ID") != "" || strings.Contains(getenv("HOME"), "/Library/Containers/")
}

// IsFile reports whether p is an existing regular file.
func IsFile(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// IsDir reports whether p is an existing folder.
func IsDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}
