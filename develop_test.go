package main

import "testing"

func TestDevelopVersion(t *testing.T) {
	versions := []string{"1.0-SNAPSHOT", "2.14-SNAPSHOT"}
	for _, version := range versions {
		b := IsValidDevelopVersion(version)
		if !b {
			t.Fatalf("IsValidDevelopVersion(%s) expecting true", version)
		}
	}
}

func TestInvalidDevelopVersion(t *testing.T) {
	versions := []string{"1.0-us_feature-SNAPSHOT", "2.14-bug_77-SNAPSHOT", "1.0"}
	for _, version := range versions {
		b := IsValidDevelopVersion(version)
		if b {
			t.Fatalf("IsValidDevelopVersion(%s) expecting true", version)
		}
	}
}
