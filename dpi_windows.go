//go:build windows

package giu

import (
	"syscall"
)

func fixDpi() {
	// Set DPI awareness to system aware (value of 1)
	shcore := syscall.NewLazyDLL("Shcore.dll")
	shproc := shcore.NewProc("SetProcessDpiAwareness")
	shproc.Call(uintptr(1))
}
