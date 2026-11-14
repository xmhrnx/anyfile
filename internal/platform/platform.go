// Package platform hands files to the operating system: opening them
// in the user's own viewers (AnyFile never renders content itself) and
// revealing them in the file manager. It also tells how AnyFile was
// installed, since the stores' packages limit what it may do.
package platform

import (
	"os"
	"runtime"
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
	return runtime.GOOS == "darwin" && os.Getenv("APP_SANDBOX_CONTAINER_ID") != ""
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
