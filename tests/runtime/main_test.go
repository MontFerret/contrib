package main

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestRuntimeProcess(t *testing.T) {
	if os.Getenv("CONTRIB_TEST_RUNTIME_PROCESS") == "1" {
		os.Args = []string{"runtime"}
		flag.CommandLine = flag.NewFlagSet("runtime", flag.ExitOnError)
		main()
		os.Exit(0)
	}

	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name  string
		query string
		want  string
		fails bool
	}{
		{name: "success", query: "RETURN {ok: true}", want: "{\"ok\":true}\n"},
		{name: "execution failure", query: `RETURN DB::SQLITE::CLOSE("invalid")`, fails: true},
		{name: "compilation failure", query: "RETURN (", fails: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			cmd := exec.CommandContext(t.Context(), executable, "-test.run=^TestRuntimeProcess$")
			cmd.Env = append(os.Environ(), "CONTRIB_TEST_RUNTIME_PROCESS=1")
			cmd.Stdin = strings.NewReader(test.query)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if test.fails {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 || stderr.Len() == 0 {
					t.Fatalf("error = %v, stderr = %q", err, stderr.String())
				}
			} else if err != nil {
				t.Fatalf("error = %v, stderr = %q", err, stderr.String())
			}

			if stdout.String() != test.want {
				t.Fatalf("stdout = %q, want %q", stdout.String(), test.want)
			}
		})
	}
}
