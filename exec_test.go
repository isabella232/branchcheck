package main

import (
	"strings"
	"testing"
)

func TestExecStdout(t *testing.T) {
	stdout, _, err := Exec("ls", "-al")
	if err != nil {
		t.Fatalf("Not expecting an error but got one: %v\n", err)
	}
	if !strings.Contains(string(stdout), "LICENSE") {
		t.Fatalf("Wanted LICENSE in stdout, but did not find it\n")
	}
}

func TestExecStderr(t *testing.T) {
	_, stderr, err := Exec("ls", "nosuchfile")
	if err == nil {
		t.Fatalf("Expecting an error but got none\n")
	}
	if !strings.Contains(string(stderr), "No such file") {
		t.Fatalf("Wanted No such file in stderr, but did not find it\n")
	}
}
