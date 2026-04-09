package cli

import (
	"os"
	"syscall"
	"unsafe"
)

type Terminal struct {
	orig  *syscall.Termios
	fd    int
	isRaw bool
}

func NewTerminal() (*Terminal, error) {
	fd := int(os.Stdin.Fd())
	var orig syscall.Termios

	_, _, err := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&orig)),
		0, 0, 0,
	)
	if err != 0 {
		return nil, err
	}

	return &Terminal{
		fd:   fd,
		orig: &orig,
	}, nil
}

func (t *Terminal) EnableRaw() error {
	raw := *t.orig

	raw.Lflag &^= syscall.ICANON | syscall.ECHO

	_, _, err := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(t.fd),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(&raw)),
		0, 0, 0,
	)
	if err != 0 {
		return err
	}

	t.isRaw = true

	return nil
}

func (t *Terminal) Restore() error {
	_, _, err := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(t.fd),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(t.orig)),
		0, 0, 0,
	)
	if err != 0 {
		return err
	}

	t.isRaw = false
	return nil
}
