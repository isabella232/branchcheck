package main

import (
	"testing"
	"fmt"
)

func TestFeatureValid(t *testing.T) {
	branches := map[string]string{
		"1.2-b-SNAPSHOT":                           "a/b",
		"1.1-us1922-SNAPSHOT":                     "feature/US1922",
		"14.6-trnk_12-SNAPSHOT":                    "feature/TRNK-12",
		"1.2.3-123-SNAPSHOT":                       "bug/123",
		"0.0-prj_4385_tok_tik_tx_trailer-SNAPSHOT": "feature/PRJ-4385-tok-tik-tx-trailer",
	}
	for version, branch := range branches {
		fmt.Printf("%s  %s\n", version, branch)
		b := IsValidFeatureVersion(branch, version)
		if !b {
			t.Fatalf("IsValidFeatureBranch(%s,%s) expecting true", branch, version)
		}
	}
}

func TestFeatureNotValid(t *testing.T) {
	branches := map[string]string{
		"1.2-b":                "a/b",             // no -SNAPSHOT
		"1.1-US-1922-SNAPSHOT": "feature/US-19222", // wrong case
		"foo": "somebranch", // branches without a "/" in the name are not valid here
	}
	for version, branch := range branches {
		b := IsValidFeatureVersion(branch, version)
		if b {
			t.Fatalf("IsValidFeatureBranch(%s,%s) expecting true", branch, version)
		}
	}
}

func TestDevelopVersion(t *testing.T) {
	versions := []string{"1.0-SNAPSHOT", "2.14-SNAPSHOT", "2.14.15-SNAPSHOT"}
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
			t.Fatalf("IsValidDevelopVersion(%s) expecting false", version)
		}
	}
}

func TestTruncateSnapshot(t *testing.T) {
	if truncateSnapshot("1.0-SNAPSHOT") != "1.0" {
		t.Fatalf("Want 1.0")
	}
}

func TestNormalizeStory(t *testing.T) {
	if normalizeStoryPart("PRJ-2") != "prj_2" {
		t.Fatalf("Want prj_2")
	}

	if normalizeStoryPart("PRJ2") != "prj2" {
		t.Fatalf("Want prj2")
	}

	if normalizeStoryPart("PRJ_2") != "prj_2" {
		t.Fatalf("Want prj_2")
	}
}
