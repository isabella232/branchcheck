package main

import (
	"strings"
	"testing"
)

func TestExecStdout(t *testing.T) {
	stdout, _, err := Exec("echo", "foo")
	if err != nil {
		t.Fatalf("Not expecting an error but got one: %v\n", err)
	}
	if !strings.Contains(string(stdout), "foo") {
		t.Fatalf("Wanted foo in stdout, but did not find it\n")
	}
}

func TestExecStderr(t *testing.T) {
	_, stderr, err := Exec("ls", "nosuchfilezzz")
	if err == nil {
		t.Fatalf("Expecting an error but got none\n")
	}
	if !strings.Contains(string(stderr), "No such file") {
		t.Fatalf("Wanted No such file in stderr, but did not find it\n")
	}
}
