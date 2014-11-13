package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

type POM struct {
	XMLName xml.Name `xml:"project"`
	Parent  struct {
		XMLName xml.Name `xml:"parent"`
		Version string   `xml:"version"`
	} `xml:"parent"`
	Version string `xml:"version"`
}

var (
	debug    bool
	excludes = flag.String("excludes", "", "comma-separated poms to exclude, by path relative to repository top level (e.g., a/pom.xml,b/pom.xml")
	version  = flag.Bool("version", false, "Print git commit from which we were built")

	skipMap map[string]string
	commit  string
)

func init() {
	debug = strings.ToLower(os.Getenv("BRANCHCHECK_DEBUG")) == "true"
	flag.Parse()
	a := strings.Split(*excludes, ",")
	skipMap = make(map[string]string, len(a))
	for _, v := range a {
		skipMap[v] = ""
	}
}

func main() {
	log.Printf("branchcheck build commit ID: %s\n", commit)
	if *version {
		os.Exit(0)
	}

	branch, err := CurrentBranch()
	if err != nil {
		log.Printf("Cannot determine current branch name.  You may not be in a git repository: %v\n", err)
		return
	}

	if branch == "master" {
		log.Printf("branchcheck does not analyze branch master.  Returning.\n")
		return
	}

	if branch == "HEAD" {
		log.Printf("You are not on a branch.\n")
		return
	}

	if debug {
		log.Printf("Analyzing branch %s\n", branch)
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

		if _, present := skipMap[pomFile]; present {
			if debug {
				log.Printf("Skipping excluded pom: %s\n", pomFile)
			}
			continue
		}

		data, err := ioutil.ReadFile(pomFile)
		if err != nil {
			log.Printf("Error reading %s: %v\n", pomFile, err)
			continue
		}

		var pom POM
		reader := bytes.NewBuffer(data)
		if err := xml.NewDecoder(reader).Decode(&pom); err != nil {
			log.Printf("Error parsing pom.xml %s: %v\n", pomFile, err)
			continue
		}

		if pom.Version == "" && pom.Parent.Version == "" {
			panic(fmt.Sprintf("pom version and parent are both empty in pom %s\n", pomFile))
		}

		var effectiveVersion string
		if pom.Version == "" {
			effectiveVersion = pom.Parent.Version
			if debug {
				log.Printf("Using parent-version %s in pom %s\n", effectiveVersion, pomFile)
			}
		} else {
			effectiveVersion = pom.Version
		}

		if strings.HasPrefix(effectiveVersion, "$") {
			if debug {
				log.Printf("Skipping pom %s because of unresolvable token %s in version element\n", pomFile, effectiveVersion)
			}
			continue
		}
		if debug {
			log.Printf("effectiveVersion %s in pom %s\n", effectiveVersion, pomFile)
		}

		if branch == "develop" {
			if !IsValidDevelopVersion(effectiveVersion) {
				log.Printf("Invalid develop branch version %s in %s\n", effectiveVersion, pomFile)
				os.Exit(-1)
			}
			continue
		}
		if !IsValidFeatureVersion(branch, effectiveVersion) {
			log.Printf("Feature branch %s has invalid version %s in %s\n", branch, effectiveVersion, pomFile)
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

func IsValidFeatureVersion(branch, version string) bool {
	// local convention that a feature branch has the form a/b.
	parts := strings.Split(branch, "/")
	if len(parts) != 2 {
		return false
	}

	// normalize the story part of the branch by lowercasing and filtering out any non-digit and non-letter characters
	var normalizedStory string
	for _, v := range strings.ToLower(parts[1]) {
		if unicode.IsDigit(rune(v)) || unicode.IsLetter(rune(v)) {
			normalizedStory = normalizedStory + string(v)
		}
	}

	// normalize the POM version by lowercasing and filtering out any non-digit and non-letter characters
	var normalizedVersion string
	for _, v := range strings.ToLower(version) {
		if unicode.IsDigit(rune(v)) || unicode.IsLetter(rune(v)) {
			normalizedVersion = normalizedVersion + string(v)
		}
	}

	return strings.HasSuffix(normalizedVersion, normalizedStory+"snapshot")
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
