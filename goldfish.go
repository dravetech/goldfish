/*
Package goldfish is used to help testing the command lines
*/
package goldfish

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// CommandTestCase case is a struct that defines a test case for a command
type CommandTestCase struct {
	Name           string   // Name of the test case, it will be used as filename for the golden files
	GoldenPath     string   // Path to the golden files
	Command        []string // Command to run
	Update         bool     // Update the golden files for this test case
	ExitCode       int      // Expected exit code
	StdoutRegex    bool
	StderrRegex    bool
	StdoutJSON     bool
	StdoutJSONList bool
	StderrJSON     bool
}

func (tc *CommandTestCase) StdoutGoldenPath() string {
	return filepath.Join(tc.GoldenPath, tc.Name+".out")
}

func (tc *CommandTestCase) StderrGoldenPath() string {
	return filepath.Join(tc.GoldenPath, tc.Name+".err")
}

// Run executes the command and validates output
func (tc *CommandTestCase) Run(t *testing.T) {
	cmd := exec.Command(tc.Command[0], tc.Command[1:]...)

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

	compareGolden(t, tc.Update, tc.StdoutGoldenPath(), stdout, "stdout", tc.StdoutRegex, tc.StdoutJSON, tc.StdoutJSONList)
	compareGolden(t, tc.Update, tc.StderrGoldenPath(), stderr, "stderr", tc.StderrRegex, tc.StderrJSON, tc.StdoutJSONList)
}

func compareGoldenString(t *testing.T, update bool, path string, buf bytes.Buffer, out string, useRegex bool) {
	goldenOut := get(t, buf.Bytes(), path, update, false)

	match := false

	if useRegex {
		re := regexp.MustCompile(string(goldenOut))
		match = re.Match(buf.Bytes())
	} else {
		match = cmp.Equal(buf.String(), string(goldenOut))
	}
	if !match {
		t.Errorf("%s doesn't match:\n%s", out, cmp.Diff(buf.String(), string(goldenOut)))
	}
}

func compareJSON(t *testing.T, goldenBytes []byte, gotBytes []byte, useRegex bool, out string) {
	goldenData := new(interface{})
	if err := json.Unmarshal(goldenBytes, goldenData); err != nil {
		t.Log("couldn't unmarshal golden data: " + err.Error())
		t.FailNow()
	}

	gotData := new(interface{})
	if err := json.Unmarshal(gotBytes, gotData); err != nil {
		t.Log("couldn't unmarshal gotten data: " + err.Error())
		t.FailNow()
	}

	opts := cmp.Options{}
	if useRegex {
		opts = append(opts, cmp.Comparer(func(x, y string) bool {
			re1, err := regexp.Compile(x)
			if err != nil {
				re1 = regexp.MustCompile("ishallnotmatchanythingatall432423")
			}
			re2, err := regexp.Compile(y)
			if err != nil {
				re2 = regexp.MustCompile("ishallnotmatchanythingatall432423")
			}
			return re1.Match([]byte(y)) || re2.Match([]byte(x)) || x == y
		}))
	}

	if !cmp.Equal(gotData, goldenData, opts...) {
		t.Errorf("%s doesn't match:\n%s", out, cmp.Diff(gotData, goldenData, opts...))
	}

}

func compareGoldenJSON(t *testing.T, update bool, path string, buf bytes.Buffer, out string, useRegex bool) {
	goldenOut := get(t, buf.Bytes(), path, update, true)
	compareJSON(t, goldenOut, buf.Bytes(), useRegex, out)
}

func compareGoldenJSONList(t *testing.T, update bool, path string, buf bytes.Buffer, out string, useRegex bool) {
	goldenOut := get(t, buf.Bytes(), path, update, false)

	goldenOutLines := strings.Split(string(goldenOut), "\n")
	goldenDataLines := strings.Split(buf.String(), "\n")
	for i := range goldenOutLines {
		if goldenOutLines[i] == "" {
			continue
		}
		compareJSON(t, []byte(goldenOutLines[i]), []byte(goldenDataLines[i]), useRegex, out)
	}
}

func compareGolden(t *testing.T, update bool, path string, buf bytes.Buffer, out string, useRegex bool, useJSON bool, useJSONList bool) {
	if useJSON {
		compareGoldenJSON(t, update, path, buf, out, useRegex)
	} else if useJSONList {
		compareGoldenJSONList(t, update, path, buf, out, useRegex)
	} else {
		compareGoldenString(t, update, path, buf, out, useRegex)
	}
}

func get(t *testing.T, actual []byte, goldenPath string, updateGolden bool, useJSON bool) []byte {
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
