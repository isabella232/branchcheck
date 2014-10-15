package main

import "testing"

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
			t.Fatalf("IsValidDevelopVersion(%s) expecting true", version)
		}
	}
}

func TestFeatureVersion(t *testing.T) {
	versions := map[string]string{
		"1.0-us_1355-SNAPSHOT":    "feature/US1355",
		"1.1.1-us_17459-SNAPSHOT": "feature/US17459",
		"2.14-bug_77-SNAPSHOT":    "feature/Bug77",
	}
	for version, branch := range versions {
		b := IsValidFeatureVersion(branch, version)
		if !b {
			t.Fatalf("IsValidFeatureVersion(%s, %s) expecting true", branch, version)
		}
	}
}
