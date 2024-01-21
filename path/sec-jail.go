package path

import (
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/sys/unix"
)

// IsPathInJail is to ensure the path is in the jail
func IsPathInJail(path string) (bool, error) {
	var stat, statLink unix.Stat_t

	// The path maybe not just only a magic link path, so open the path to get the fd
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()
	err = unix.Fstat(int(file.Fd()), &stat)
	if err != nil {
		return false, err
	}

	// We can get the real path from `/proc/self/fd/<fd>`
	link, err := os.Readlink(fmt.Sprintf("/proc/self/fd/%d", file.Fd()))
	if err != nil {
		return false, err
	}

	err = unix.Lstat(link, &statLink)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	if stat.Dev == statLink.Dev && stat.Ino == statLink.Ino {
		return true, nil
	}
	return false, fmt.Errorf("can't find %s", path)
}

// LookPath is to find the exactly path of the executable file
func SecLookPath(path string) (string, error) {
	name, err := exec.LookPath(path)
	if err != nil {
		return "", err
	}
	bl, err := IsPathInJail(name)
	if err != nil {
		return "", err
	}
	if !bl {
		return "", fmt.Errorf("can't find %s in the current file system", path)
	}
	return name, nil
}
