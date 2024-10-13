package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/opencontainers-sec/go-containersec/execve"
	"golang.org/x/sys/unix"
)

func main() {
	if wd, err := unix.Getwd(); errors.Is(err, unix.ENOENT) {
		fmt.Println("current working directory is outside of container mount namespace root -- possible container breakout detected")
	} else if err != nil {
		fmt.Printf("failed to verify if current working directory is safe: %v\n", err)
	} else if !filepath.IsAbs(wd) {
		// We shouldn't ever hit this, but check just in case.
		fmt.Printf("current working directory is not absolute -- possible container breakout detected: cwd is %q\n", wd)
	} else {
		if len(os.Args) < 2 {
			fmt.Printf("Usage: dmz entrypoint args...\n")
		} else {
			if err := execve.Run(os.Args[1], os.Args[1:], os.Environ()); err != nil {
				fmt.Println(err)
			}
		}
	}
	os.Exit(255)
}
