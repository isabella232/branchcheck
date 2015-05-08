package main

import "testing"

func TestFeatureValid(t *testing.T) {
	branches := map[string]string{
		"1.2-b-SNAPSHOT":                           "a/b",
		"1.1-US_1922-SNAPSHOT":                     "feature/US1922",
		"14.6-TRNK_12-SNAPSHOT":                    "feature/TRNK-12",
		"1.2.3-123-SNAPSHOT":                       "bug/123",
		"0.0-PRJ-4385-tok-tik-tx-trailer-SNAPSHOT": "feature/PRJ-4385-tok-tik-tx-trailer",
	}
	for version, branch := range branches {
		b := IsValidFeatureVersion(branch, version)
		if !b {
			t.Fatalf("IsValidFeatureBranch(%s,%s) expecting true", branch, version)
		}
	}
}

func TestFeatureNotValid(t *testing.T) {
	branches := map[string]string{
		"1.2-b":                "a/b",             // no -SNAPSHOT
		"1.1-us_1922-SNAPSHOT": "feature/US19222", // wrong case
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
