//go:build linux
// +build linux

package system

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

func execveat(fd uintptr, pathname string, args []string, env []string, flags int) error {
	pathnamep, err := syscall.BytePtrFromString(pathname)
	if err != nil {
		return err
	}

	argvp, err := syscall.SlicePtrFromStrings(args)
	if err != nil {
		return err
	}

	envp, err := syscall.SlicePtrFromStrings(env)
	if err != nil {
		return err
	}

	_, _, errno := syscall.Syscall6(
		unix.SYS_EXECVEAT,
		fd,
		uintptr(unsafe.Pointer(pathnamep)),
		uintptr(unsafe.Pointer(&argvp[0])),
		uintptr(unsafe.Pointer(&envp[0])),
		uintptr(flags),
		0,
	)
	return errno
}

func Fexecve(fd uintptr, args []string, env []string) error {
	var err error
	for {
		err = execveat(fd, "", args, env, unix.AT_EMPTY_PATH)
		if err != unix.EINTR { // nolint:errorlint // unix errors are bare
			break
		}
	}

	return os.NewSyscallError("execveat", err)
}
