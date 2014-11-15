package main

import (
	"bufio"
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
	debug           bool
	excludes        = flag.String("excludes", "", "comma-separated poms to exclude, by path relative to repository top level (e.g., a/pom.xml,b/pom.xml")
	version         = flag.Bool("version", false, "Print git commit from which we were built")
	versionDupCheck = flag.Bool("version-dups", false, "Iterate over all branches and check for duplicate POM versions.  Uses git ls-remote.")

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

	if *versionDupCheck {
		DupCheck()
		return
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
				log.Printf("Using parent-version in pom %s\n", pomFile)
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

func GitFetch() error {
	cmd := "git"
	args := []string{"fetch"}
	if debug {
		log.Printf("%s %v\n", cmd, args)
	}
	command := exec.Command(cmd, args...)
	if _, err := command.CombinedOutput(); err != nil {
		return err
	}
	return nil
}

func GitCheckoutBranch(branchName string) error {
	cmd := "git"
	args := []string{"checkout", branchName}
	if debug {
		log.Printf("%s %v\n", cmd, args)
	}
	command := exec.Command(cmd, args...)
	if _, err := command.CombinedOutput(); err != nil {
		return err
	}
	return nil
}

func GetBranches() ([]string, error) {
	cmd := "git"
	args := []string{"ls-remote", "--heads"}
	if debug {
		log.Printf("%s %v\n", cmd, args)
	}

	command := exec.Command(cmd, args...)
	if data, err := command.Output(); err != nil {
		return nil, err
	} else {
		r := make([]string, 0)
		readbuffer := bytes.NewBuffer(data)
		reader := bufio.NewReader(readbuffer)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			parts := strings.Fields(scanner.Text())
			branchName := strings.Replace(parts[1], "refs/heads/", "", -1)
			r = append(r, branchName)
		}
		return r, nil
	}
}

func DupCheck() error {
	if err := GitFetch(); err != nil {
		fmt.Errorf("Error in git-fetch: %v\n", err)
	}
	if branches, err := GetBranches(); err != nil {
		fmt.Errorf("Error getting remote heads: %v\n", err)
	} else {
		t := make(map[string][]string)
		for _, branch := range branches {
			if err := GitCheckoutBranch(branch); err != nil {
				return fmt.Errorf("Cannot checkout branch %s: %v\n", branch, err)
			}
			if err := walkToGitRoot("."); err != nil {
				return err
			}
			effectiveVersion, err := pomVersion("pom.xml")
			if err != nil {
				return err
			}
			_, present := t[effectiveVersion]
			if !present {
				t[effectiveVersion] = make([]string, 0)
			}
			t[effectiveVersion] = append(t[effectiveVersion], branch)
		}
		for k, v := range t {
			if len(v) > 1 {
				log.Printf("multiple branches %+v with version %s\n", v, k)
			}
		}
	}
	return nil
}

func walkToGitRoot(dir string) error {
	return nil
}

func pomVersion(pomFile string) (string, error) {
	data, err := ioutil.ReadFile(pomFile)
	if err != nil {
		return "", err
	}

	var pom POM
	reader := bytes.NewBuffer(data)
	if err := xml.NewDecoder(reader).Decode(&pom); err != nil {
		return "", err
	}

	if pom.Version == "" && pom.Parent.Version == "" {
		return "", fmt.Errorf("pom version and parent are both empty in pom %s\n", pomFile)
	}

	var effectiveVersion string
	if pom.Version == "" {
		effectiveVersion = pom.Parent.Version
		if debug {
			log.Printf("Using parent-version in pom %s\n", pomFile)
		}
	} else {
		effectiveVersion = pom.Version
	}

	if strings.HasPrefix(effectiveVersion, "$") {
		return "", fmt.Errorf("Cannot analyze pom %s because of unresolvable token %s in version element\n", pomFile, effectiveVersion)
	}
	if debug {
		log.Printf("effectiveVersion %s in pom %s\n", effectiveVersion, pomFile)
	}

	return effectiveVersion, nil
}
