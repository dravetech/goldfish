package goldfish

import "testing"

func TestPass(t *testing.T) {
	cases := []CommandTestCase{
		{
			Name:       "everything_fine",
			GoldenPath: "testdata",
			Command:    []string{"echo", "hello, how are you?"},
		},
		{
			Name:        "json_output_with_regex",
			GoldenPath:  "testdata",
			StdoutRegex: true,
			StdoutJSON:  true,
			Command:     []string{"echo", `{"asd":"qweqwe","qwe": [1, 2, 3]}`},
		},
		{
			Name:        "multiline_regex",
			GoldenPath:  "testdata",
			StdoutRegex: true,
			Command:     []string{"cat", "testdata/multiline_regex.mock"},
		},
		{
			Name:       "command_failed",
			GoldenPath: "testdata",
			Command:    []string{"ls", "/lalalalalalala"},
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
