package main

import "testing"

func TestIsFeatureBranch(t *testing.T) {
	branches := []string{"a/b", "feature/US1922", "bug/123", "a/b/c"}
	for _, branch := range branches {
		b := IsFeatureBranch(branch)
		if !b {
			t.Fatalf("IsFeatureBranch(%s) expecting true", branch)
		}
	}
}

func TestIsNotFeatureBranch(t *testing.T) {
	branches := []string{"a", "feature"}
	for _, branch := range branches {
		b := IsFeatureBranch(branch)
		if b {
			t.Fatalf("IsFeatureBranch(%s) expecting false", branch)
		}
	}
}
