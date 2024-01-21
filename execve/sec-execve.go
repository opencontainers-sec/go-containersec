package execve

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/opencontainers-sec/go-containersec/execve/system"
	"github.com/opencontainers-sec/go-containersec/path"
	"golang.org/x/sys/unix"
)

func isShebang(f int) (bool, error) {
	// Read the first 2 bytes to check if it's a shebang.
	var buf [2]byte
	if _, err := unix.Read(f, buf[:]); err != nil {
		return false, err
	}
	return buf[0] == '#' && buf[1] == '!', nil
}

func readScript(f int, cmd string, args []string) (string, []string, error) {
	scanner := bufio.NewScanner(os.NewFile(uintptr(f), "script"))
	if scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		arr := strings.Split(line, " ")
		if len(arr) > 0 {
			nargs := append([]string{cmd}, arr[1:]...)
			nargs = append(nargs, args...)
			return arr[0], nargs, nil
		}
	}
	return "", args, fmt.Errorf("failed to read script")
}

func GetSecExecve(cmd string, args []string, env []string) (int, string, []string, []string, error) {
	f, err := unix.Open(cmd, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return 0, "", args, env, fmt.Errorf("failed to open %s: %w", cmd, err)
	}

	ncmd := cmd
	nargs := args[:]
	depth := 0
	for {
		sb, err := isShebang(f)
		if err != nil {
			_ = unix.Close(f)
			return 0, "", args, env, err
		}
		if sb {
			depth++
			/* This allows 4 levels of binfmt rewrites before failing hard.
			   https://github.com/torvalds/linux/blob/9d64bf433c53cab2f48a3fff7a1f2a696bc5229a/fs/exec.c#L1773-L1777
			*/
			if depth > 5 {
				_ = unix.Close(f)
				return 0, "", args, env, unix.ELOOP
			}

			ncmd, nargs, err = readScript(f, ncmd, nargs)
			_ = unix.Close(f)
			if err != nil {
				return 0, "", args, env, err
			}
			f, err = unix.Open(ncmd, unix.O_RDONLY|unix.O_CLOEXEC, 0)
			if err != nil {
				return 0, "", args, env, fmt.Errorf("failed to open %s: %w", ncmd, err)
			}
		} else {
			break
		}
	}

	_, err = unix.Seek(f, 0, 0)
	if err != nil {
		_ = unix.Close(f)
		return 0, "", args, env, err
	}

	return f, ncmd, nargs, env, nil
}

func Run(cmd string, args []string, env []string) error {
	sfd, scmd, sargs, senv, err := GetSecExecve(cmd, args, env)
	if err != nil {
		return err
	}
	jail, err := path.IsPathInJail(scmd)
	if err != nil {
		return err
	}
	if !jail {
		_ = unix.Close(sfd)
		return fmt.Errorf("can't find %s in the current file system", scmd)
	}
	return system.Fexecve(uintptr(sfd), sargs, senv)
}
