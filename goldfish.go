/*
Package goldfish is used to help testing the command lines
*/
package golden

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	update = flag.Bool("test.update", false, "update golden file")
)

// CommandTestCase case is a struct that defines a test case for a command
type CommandTestCase struct {
	Name        string // Name of the test case, it will be used as filename for the golden files
	GoldenPath  string // Path to the golden files
	Command     string // Command to run
	Update      bool   // Update the golden files for this test case
	ExitCode    int    // Expected exit code
	ExpectedErr error
}

func parseCommand(command string) []string {
	return strings.Split(command, " ")
}

func (tc *CommandTestCase) StdoutGoldenPath() string {
	return filepath.Join(tc.GoldenPath, tc.Name+".out")
}

func (tc *CommandTestCase) StderrGoldenPath() string {
	return filepath.Join(tc.GoldenPath, tc.Name+".err")
}

// Run executes the command and validates output
func (tc *CommandTestCase) Run(t *testing.T) {
	c := parseCommand(tc.Command)
	cmd := exec.Command(c[0], c[1:]...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	exitCode := 0

	err := cmd.Run()
	switch v := err.(type) {
	case *exec.ExitError:
		exitCode = v.ExitCode()
	case nil:
	default:
		t.Log("command execution triggered unknown error: " + v.Error())
		t.FailNow()
	}

	if exitCode != tc.ExitCode {
		t.Errorf("exit code doesn't match; got (%d), expected (%d)\n", exitCode, tc.ExitCode)
	}

	goldenOut := get(t, stdout.Bytes(), tc.StdoutGoldenPath(), tc.Update || *update)
	if !cmp.Equal(stdout.String(), string(goldenOut)) {
		t.Error("stdout doesn't match:\n" + cmp.Diff(stdout.String(), string(goldenOut)))
	}

	goldenErr := get(t, stderr.Bytes(), tc.StderrGoldenPath(), tc.Update || *update)
	if !cmp.Equal(stderr.String(), string(goldenErr)) {
		t.Error("stderr doesn't match:\n" + cmp.Diff(stderr.String(), string(goldenErr)))
	}
}

func get(t *testing.T, actual []byte, goldenPath string, updateGolden bool) []byte {
	if updateGolden {
		if err := ioutil.WriteFile(goldenPath, actual, 0644); err != nil {
			t.Fatal(err)
		}
	}
	expected, err := ioutil.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	return expected
}
