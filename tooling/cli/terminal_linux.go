package main

import (
	"os"
	"syscall"
	"unsafe"
)

func terminalInput(f *os.File) bool {
	var state syscall.Termios
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&state)))
	return err == 0
}
