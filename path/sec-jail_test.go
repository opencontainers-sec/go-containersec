package path

import (
	"fmt"
	"os"
	"path"
	"testing"

	"golang.org/x/sys/unix"
)

func TestSecLookPath(t *testing.T) {
	type testData struct {
		name string
		path string
		bl   bool
		err  bool
	}
	testCases := []testData{
		{
			name: "basic",
			path: "/bin/echo",
			bl:   true,
			err:  false,
		},
		{
			name: "error",
			path: "/bin/ech",
			bl:   false,
			err:  true,
		},
	}
	for _, tc := range testCases {
		bl, err := IsPathInJail(tc.path)
		if err != nil && !tc.err {
			t.Fatalf("failed to get the path %s: %v", tc.path, err)
		}
		if bl != tc.bl {
			t.Fatalf("the path %s should be %v, but got %v", tc.path, tc.bl, bl)
		}
	}
}

func TestSecLookPathInJail(t *testing.T) {
	blChroot := false
	dir, err := os.MkdirTemp("/tmp", "jail-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if blChroot {
			os.RemoveAll("/")
		} else {
			os.RemoveAll(dir)
		}
	}()
	err = os.Mkdir(path.Join(dir, "proc"), 0755)
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if blChroot {
			os.RemoveAll("/proc")
		} else {
			os.RemoveAll(path.Join(dir, "proc"))
		}
	}()
	err = unix.Mount("proc", path.Join(dir, "proc"), "proc", 0, "")
	if err != nil {
		t.Fatalf("failed to mount proc dir: %v", err)
	}
	defer func() {
		if blChroot {
			unix.Unmount("/proc", 0)
		} else {
			unix.Unmount(path.Join(dir, "proc"), 0)
		}
	}()

	fd, err := unix.Open("../execve/testdata/basic/echo.sh", unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		t.Fatalf("failed to open the file: %v", err)
	}
	if fd <= 0 {
		t.Fatalf("failed to open the file")
	}
	defer unix.Close(fd)
	err = unix.Chroot(dir)
	if err != nil {
		t.Fatalf("failed to chroot: %v", err)
	}
	blChroot = true
	fdDir, err := os.Open("/proc/self/fd")
	if err != nil {
		t.Fatalf("failed to open the file")
	}
	defer fdDir.Close()
	fdList, err := fdDir.Readdirnames(-1)
	if err != nil {
		t.Fatalf("failed to open the file")
	}
	for _, fdStr := range fdList {
		t.Log(fdStr)
	}
	bl, err := IsPathInJail(fmt.Sprintf("/proc/self/fd/%d", fd))
	if err != nil {
		t.Fatalf("failed to get the path: %v", err)
	}
	if bl {
		t.Fatalf("the path should not be in the jail")
	}
}
