package execve

import (
	"testing"

	"golang.org/x/sys/unix"
)

func TestGetSecExecve(t *testing.T) {
	type testData struct {
		name  string
		cmd   string
		args  []string
		env   []string
		scmd  string
		sargs []string
		err   bool
	}
	testCases := []testData{
		{
			name:  "basic",
			cmd:   "testdata/basic/echo.sh",
			args:  []string{"hello", "world"},
			env:   []string{},
			scmd:  "/bin/echo",
			sargs: []string{"testdata/basic/echo.sh", "hello", "world"},
			err:   false,
		},
		{
			name:  "complex",
			cmd:   "testdata/complex/first.sh",
			args:  []string{"hello", "world"},
			env:   []string{},
			scmd:  "/bin/echo",
			sargs: []string{"testdata/basic/echo.sh", "testdata/complex/first.sh", "hello", "world"},
			err:   false,
		},
		{
			name:  "error",
			cmd:   "testdata/error/first.sh",
			args:  []string{"hello", "world"},
			env:   []string{},
			scmd:  "",
			sargs: nil,
			err:   true,
		},
		{
			name:  "binfmt",
			cmd:   "/bin/echo",
			args:  []string{"hello", "world"},
			env:   []string{},
			scmd:  "/bin/echo",
			sargs: []string{"hello", "world"},
			err:   false,
		},
	}
	for _, tc := range testCases {
		sfd, scmd, sargs, senv, err := GetSecExecve(tc.cmd, tc.args, tc.env)
		if tc.err && err == nil {
			t.Fatalf("Failed to run %s: we should got an error.", tc.name)
		}
		if !tc.err && err != nil {
			t.Fatalf("Failed to run %s: %s", tc.name, err)
		}
		if !tc.err {
			if sfd <= 0 {
				t.Fatalf("Failed to get fd")
			}
			defer unix.Close(sfd)
			if scmd != tc.scmd {
				t.Fatalf("Failed to get cmd %s: expect %s, got %s", tc.name, tc.scmd, scmd)
			}
			if len(sargs) != len(tc.sargs) {
				t.Fatalf("Failed to get args %s: %s", tc.name, sargs)
			}
			for i := 0; i < len(sargs); i++ {
				if sargs[i] != tc.sargs[i] {
					t.Fatalf("Failed to get args %s: expect %s, got %s", tc.name, tc.sargs[i], sargs[i])
				}
			}
			if len(senv) != len(tc.env) {
				t.Fatalf("Failed to get env %s: %s", tc.name, senv)
			}
		} else {
			t.Logf("%s: %s", tc.name, err)
		}
	}
}
