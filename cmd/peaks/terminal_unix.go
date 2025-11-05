//go:build darwin || linux || freebsd || openbsd || netbsd
// +build darwin linux freebsd openbsd netbsd

package main

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// getTerminalHeight attempts to get terminal height using ioctl
func getTerminalHeight() int {
	ws := &unix.Winsize{}

	// Try stdout first (works better in daemon mode)
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(unix.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	// If stdout fails, try stderr
	if errno != 0 {
		_, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stderr),
			uintptr(unix.TIOCGWINSZ),
			uintptr(unsafe.Pointer(ws)))
	}

	// If both fail, try stdin as last resort
	if errno != 0 {
		_, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stdin),
			uintptr(unix.TIOCGWINSZ),
			uintptr(unsafe.Pointer(ws)))
	}

	if errno != 0 {
		return 24 // Fallback
	}

	return int(ws.Row)
}

// getTerminalWidth attempts to get terminal width using ioctl
func getTerminalWidth() int {
	ws := &unix.Winsize{}

	// Try stdout first (works better in daemon mode)
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(unix.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	// If stdout fails, try stderr
	if errno != 0 {
		_, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stderr),
			uintptr(unix.TIOCGWINSZ),
			uintptr(unsafe.Pointer(ws)))
	}

	// If both fail, try stdin as last resort
	if errno != 0 {
		_, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stdin),
			uintptr(unix.TIOCGWINSZ),
			uintptr(unsafe.Pointer(ws)))
	}

	if errno != 0 {
		return 80 // Fallback
	}

	return int(ws.Col)
}
