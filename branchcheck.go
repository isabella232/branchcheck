package main

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
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
	debug = strings.ToLower(os.Getenv("BRANCHCHECK_DEBUG")) == "true"
}

func main() {
	branch, err := CurrentBranch()
	if err != nil {
		log.Printf("Cannot determine current branch name\n", err)
		return
	}

	if branch == "HEAD" {
		log.Printf("You are not on a branch.  Returning.\n")
		return
	}

	if debug {
		log.Printf("Validating branch %s\n", branch)
	}

	poms, err := FindPoms()
	if err != nil || len(poms) == 0 {
		log.Printf("Cannot find POMs\n", err)
		return
	}

	for _, pomFile := range poms {
		if debug {
			log.Printf("Analyzing %s\n", pomFile)
		}

		data, err := ioutil.ReadFile(pomFile)
		if err != nil {
			log.Printf("Error reading %s: %v\n", pomFile, err)
			continue
		}

		var pom POM
		reader := bytes.NewBuffer(data)
		if err := xml.NewDecoder(reader).Decode(&pom); err != nil {
			log.Printf("error parsing pom.xml %s: %v\n", pomFile, err)
			continue
		}
		// An inherited <version> will have pom.Version==""
		if pom.Version == "" || branch == "master" {
			continue
		}
		if branch == "develop" {
			if !IsValidDevelopVersion(pom.Version) {
				log.Printf("invalid develop branch version %s in %s\n", pom.Version, pomFile)
				os.Exit(-1)
			} else {
				continue
			}
		}
		if IsFeatureBranch(branch) && !IsValidFeatureVersion(branch, pom.Version) {
			log.Printf("feature branch %s has invalid version %s in %s\n", branch, pom.Version, pomFile)
			os.Exit(-1)
		}
	}
}

func CurrentBranch() (string, error) {
	cmd := "git"
	args := []string{"rev-parse", "--abbrev-ref", "HEAD"}
	if debug {
		log.Printf("%s %v\n", cmd, args)
	}
	command := exec.Command(cmd, args...)
	if data, err := command.Output(); err != nil {
		return "", err
	} else {
		return string(data[:len(data)-1]), nil
	}
}

func IsFeatureBranch(branch string) bool {
	b := strings.Contains(branch, "/")
	if debug {
		log.Printf("%s is a feature branch: %v\n", branch, b)
	}
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
	if debug {
		log.Printf("%s is a branch compatible with version %s: %v\n", branch, version, match)
	}
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
	if debug {
		log.Printf("found %d poms\n", len(files))
	}
	return files, nil
}
