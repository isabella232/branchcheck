package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type POM struct {
	XMLName xml.Name `xml:"project"`
	Version string   `xml:"version"`
}

var (
	debug bool
)

func init() {
	debug = os.Getenv("BRANCHCHECK_DEBUG") == "true"
}

func main() {
	branch, err := CurrentBranch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine current branch name\n", err)
		return
	}
	if branch == "HEAD" {
		fmt.Fprintf(os.Stderr, "You are not on a branch.  Returning.\n")
		return
	}
	if debug {
		fmt.Fprintf(os.Stderr, "Validating branch %s\n", branch)
	}

	poms, err := FindPoms()
	if err != nil || len(poms) == 0 {
		fmt.Fprintf(os.Stderr, "Cannot find POMs\n", err)
		return
	}

	for _, pomFile := range poms {
		data, err := ioutil.ReadFile(pomFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", pomFile, err)
			continue
		}

		if debug {
			fmt.Fprintf(os.Stderr, "Analyzing %s\n", pomFile)
		}

		var pom POM
		reader := bytes.NewBuffer(data)
		if err := xml.NewDecoder(reader).Decode(&pom); err != nil {
			fmt.Fprintf(os.Stderr, "error parsing pom.xml: %v\n", err)
			continue
		}
		if branch == "develop" && !IsValidDevelopVersion(pom.Version) {
			fmt.Fprintf(os.Stderr, "invalid develop branch version %s in %s\n", pom.Version, pomFile)
			os.Exit(-1)
		}
		if IsFeatureBranch(branch) && !IsValidFeatureVersion(branch, pom.Version) {
			fmt.Fprintf(os.Stderr, "feature branch %s has invalid version %s in %s\n", branch, pom.Version, pomFile)
			os.Exit(-1)
		}
	}
}

func CurrentBranch() (string, error) {
	cmd := "git"
	args := []string{"rev-parse", "--abbrev-ref", "HEAD"}
	if debug {
		fmt.Fprintf(os.Stderr, "%s %v\n", cmd, args)
	}
	command := exec.Command(cmd, args...)
	if data, err := command.Output(); err != nil {
		return "", err
	} else {
		return string(data[:len(data)-1]), nil
	}
}

func IsFeatureBranch(branch string) bool {
	b := strings.Index(branch, "/") != -1
	if debug {
		fmt.Fprintf(os.Stderr, "%s is a feature branch: %v\n", branch, b)
	}
	return b
}

func IsValidFeatureVersion(branch, version string) bool {
	if version == "" {
		return true
	}
	parts := strings.Split(branch, "/")
	if len(parts) != 2 {
		return true
	}
	story := strings.ToLower(parts[1])
	version = strings.Replace(version, "_", "", -1)
	regex := "[1-9]+(\\.[0-9]+)+-" + story + "-SNAPSHOT"
	match, _ := regexp.MatchString(regex, version)
	if debug {
		fmt.Fprintf(os.Stderr, "%s is a branch compatible with version %s\n", branch, version)
	}
	return match
}

func IsValidDevelopVersion(version string) bool {
	if version == "" {
		return true
	}
	match, _ := regexp.MatchString("[1-9]+(\\.[0-9]+)+-SNAPSHOT", version)
	return match
}

func FindPoms() ([]string, error) {
	files := make([]string, 0)
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "pom.xml" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
