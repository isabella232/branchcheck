package main

import (
	"testing"
)

func TestFeatureValid(t *testing.T) {
	branches := map[string]string{
		"1.2-b-SNAPSHOT":                           "feature/b",
		"1.1-us1922-SNAPSHOT":                      "feature/US1922",
		"14.6-trnk_12-SNAPSHOT":                    "feature/TRNK-12",
		"1.2.3-123-SNAPSHOT":                       "hotfix/123",
		"0.0-prj_4385_tok_tik_tx_trailer-SNAPSHOT": "feature/PRJ-4385-tok-tik-tx-trailer",
		"1.1.32.1-PRJ-000-hotfix-test-SNAPSHOT":    "hotfix/1.1.32.1-PRJ-000-hotfix-test",
		"1.1.32.1-PRJ-000-XXX-hotfix-test-SNAPSHOT":    "hotfix/1.1.32.1-PRJ-000-xxx-hotfix-test",
	}
	for version, branch := range branches {
		b := IsBranchVersionCompatible(branch, version)
		if !b {
			t.Fatalf("IsValidFeatureBranch(%s,%s) expecting true", branch, version)
		}
	}
}

func TestFeatureNotValid(t *testing.T) {
	branches := map[string]string{
		"1.2-b":                "a/b",              // no -SNAPSHOT
		"1.1-US-1922-SNAPSHOT": "feature/US-19222", // wrong case
		"1.1-US_1922-SNAPSHOT": "feature/US-19222", // wrong case and "-" not mapped to "_"
		"2.0-SNAPSHOT":          "bug/US-19222", // prefix bug unknown
		"1.0-SNAPSHOT": "somebranch",   // branches without a "/" in the name are not valid here
	}
	for version, branch := range branches {
		b := IsBranchVersionCompatible(branch, version)
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

func TestTruncateVersion(t *testing.T) {
	r, ok := truncateVersion("1.0-SNAPSHOT")
	if !ok {
		t.Fatalf("Branch has the expected -SNAPSHOT suffix")
	}
	if r != "1.0" {
		t.Fatalf("Want 1.0 but got %s\n", r)
	}
	if _, ok := truncateVersion("1.0"); ok {
		t.Fatalf("Want not ok.  No -SNAPSHOT suffix.")
	}
}

func TestNormalizeStory(t *testing.T) {
	if normalizeStory("PRJ-2") != "prj_2" {
		t.Fatalf("Want prj_2")
	}

	if normalizeStory("PRJ2") != "prj2" {
		t.Fatalf("Want prj2")
	}

	if normalizeStory("PRJ_2") != "prj_2" {
		t.Fatalf("Want prj_2")
	}
}

func TestBranchPartsOK(t *testing.T) {
	x, y, ok := branchParts("feature/PRJ-22")
	if !ok {
		t.Fatalf("Branch is well-formed.\n")
	}

	if x != "feature" {
		t.Fatalf("Want feature but got %s\n", x)
	}

	if y != "PRJ-22" {
		t.Fatalf("Want PRJ-22 but got %s\n", y)
	}
}

func TestBranchPartsMalformed(t *testing.T) {
	if _, _, ok := branchParts("feature/PRJ-22/2"); ok {
		t.Fatalf("Branch with more than 1 / is not well-formed.\n")
	}

	if _, _, ok := branchParts("feature"); ok {
		t.Fatalf("Branch without a single / is not well-formed.\n")
	}
}
