//go:build windows
// +build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

type coord struct {
	X int16
	Y int16
}

type smallRect struct {
	Left   int16
	Top    int16
	Right  int16
	Bottom int16
}

type consoleScreenBufferInfo struct {
	Size              coord
	CursorPosition    coord
	Attributes        uint16
	Window            smallRect
	MaximumWindowSize coord
}

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

// getTerminalHeight attempts to get terminal height on Windows
func getTerminalHeight() int {
	var csbi consoleScreenBufferInfo
	handle := syscall.Handle(os.Stdout.Fd())

	ret, _, _ := procGetConsoleScreenBufferInfo.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&csbi)))

	if ret == 0 {
		return 24 // Fallback
	}

	return int(csbi.Window.Bottom - csbi.Window.Top + 1)
}

// getTerminalWidth attempts to get terminal width on Windows
func getTerminalWidth() int {
	var csbi consoleScreenBufferInfo
	handle := syscall.Handle(os.Stdout.Fd())

	ret, _, _ := procGetConsoleScreenBufferInfo.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&csbi)))

	if ret == 0 {
		return 80 // Fallback
	}

	return int(csbi.Window.Right - csbi.Window.Left + 1)
}
