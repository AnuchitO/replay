package main

import (
	"syscall"
	"unsafe"
)

type termios syscall.Termios

func makeRaw(fd uintptr) (*termios, error) {
	var old termios
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TIOCGETA, uintptr(unsafe.Pointer(&old)), 0, 0, 0); err != 0 {
		return nil, err
	}

	raw := old
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TIOCSETA, uintptr(unsafe.Pointer(&raw)), 0, 0, 0); err != 0 {
		return nil, err
	}

	return &old, nil
}

func restore(fd uintptr, state *termios) {
	syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TIOCSETA, uintptr(unsafe.Pointer(state)), 0, 0, 0)
}
