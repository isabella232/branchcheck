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

func main() {
	branchName, err := CurrentBranchName()
	if err != nil {
		fmt.Printf("Cannot determine current branch name\n", err)
		return
	}

	poms, err := FindPoms()
	if err != nil {
		fmt.Printf("Cannot find POMs\n", err)
		return
	}

	for _, pom := range poms {
		data, err := ioutil.ReadFile(pom)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", pom, err)
			continue
		}
		var pom POM
		reader := bytes.NewBuffer(data)
		if err := xml.NewDecoder(reader).Decode(&pom); err != nil {
			fmt.Printf("error parsing pom.xml: %v\n", err)
			continue
		}
		if branchName == "develop" && !IsValidDevelopVersion(pom.Version) {
			fmt.Printf("invalid develop branch version %s in %s\n", pom.Version, pom)
			os.Exit(-1)
		}
	}
}

func CurrentBranchName() (string, error) {
	command := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	if data, err := command.Output(); err != nil {
		return "", err
	} else {
		return string(data[:len(data)-1]), nil
	}
}

func IsFeatureBranch(branch string) bool {
	b := strings.HasPrefix(branch, "feature/") || strings.HasPrefix(branch, "bug/")
	return b
}

func IsValidFeatureVersion(branch, version string) bool {
	parts := strings.Split(branch, "/")
	if len(parts) != 2 {
		return true
	}
	story := strings.ToLower(parts[1])
	version = strings.Replace(version, "_", "", -1)
	regex := "[1-9]+(\\.[0-9]+)+-" + story + "-SNAPSHOT"
	match, _ := regexp.MatchString(regex, version)
	return match
}

func IsValidDevelopVersion(version string) bool {
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
