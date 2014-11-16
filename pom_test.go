package main

import "testing"

func TestPomVersion(t *testing.T) {
	version, err := PomVersion("test-data/no-parent/pom.xml")
	if err != nil {
		t.Fatalf("No expecting an error, but got one: %v\n", err)
	}
	if version != "1.0-FAKE_1234-SNAPSHOT" {
		t.Fatalf("Wanted 1.0-FAKE_1234-SNAPSHOT, but got one: %v\n", version)
	}
}

func TestPomWithParentVersion(t *testing.T) {
	version, err := PomVersion("test-data/with-parent/pom.xml")
	if err != nil {
		t.Fatalf("No expecting an error, but got one: %v\n", err)
	}
	if version != "2.0-FAKE_1234-SNAPSHOT" {
		t.Fatalf("Wanted 2.0-FAKE_1234-SNAPSHOT, but got one: %v\n", version)
	}
}

func TestPomWithOverriddenParentVersion(t *testing.T) {
	version, err := PomVersion("test-data/with-parent-but-override-version/pom.xml")
	if err != nil {
		t.Fatalf("No expecting an error, but got one: %v\n", err)
	}
	if version != "3.0-FAKE_1234-SNAPSHOT" {
		t.Fatalf("Wanted 3.0-FAKE_1234-SNAPSHOT, but got one: %v\n", version)
	}
}

func TestFileNotFound(t *testing.T) {
	_, err := PomVersion("test-data/nosuch-pom.xml")
	if err == nil {
		t.Fatalf("Expecting an error, but did not get one\n")
	}
}
