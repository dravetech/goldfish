package golden

import "testing"

import "os/exec"

func TestSomething(t *testing.T) {
	cases := []CommandTestCase{
		{
			Name:       "everything_fine",
			GoldenPath: "testdata",
			Command:    `echo "hello, how are you?"`,
		},
		{
			Name:        "command_not_in_path",
			GoldenPath:  "testdata",
			Command:     "ls-la",
			ExpectedErr: exec.ErrNotFound{},
		},
		{
			Name:       "command_failed",
			GoldenPath: "testdata",
			Command:    "ls /lalalalalalala",
			ExitCode:   2,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			tc.Run(t)
		})
	}
}
