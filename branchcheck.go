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
	versionDupCheck = flag.Bool("version-dups", false, "Iterate over all branches and check for duplicate POM versions.  Uses git ls-remote to get remote branches.")
	branchCompat    = flag.Bool("branch-compat", true, "Verify branch name and POM versions are compatible.")

	excludesMap map[string]string
	commit      string
)

func init() {
	debug = strings.ToLower(os.Getenv("BRANCHCHECK_DEBUG")) == "true"
	flag.Parse()
	a := strings.Split(*excludes, ",")
	excludesMap = make(map[string]string, len(a))
	for _, v := range a {
		excludesMap[v] = ""
	}
}

func main() {
	log.Printf("branchcheck build commit ID: %s\n", commit)
	if *version {
		os.Exit(0)
	}

	if *versionDupCheck {
		if versionOccurrenceMap, err := DupCheck(); err != nil {
			log.Fatalf("error running main.DupCheck(): %v\n", err)
		} else {
			someMultiples := false
			for k, v := range versionOccurrenceMap {
				if len(v) > 1 {
					someMultiples = true
					log.Printf("multiple branches %+v with version %s\n", v, k)
				}
			}
			if someMultiples {
				os.Exit(1)
			}
		}
		return
	}

	if *branchCompat {
		if err := BranchCompat(); err != nil {
			log.Printf("%v\n", err)
			os.Exit(1)
		}
		return
	}
}

func BranchCompat() error {
	branch, err := CurrentBranch()
	if err != nil {
		return fmt.Errorf("Cannot determine current branch name.  You may not be in a git repository: %v\n", err)
	}

	if branch == "master" {
		return fmt.Errorf("branchcheck does not analyze branch master.  Returning.\n")
	}

	if branch == "HEAD" {
		return fmt.Errorf("You are not on a branch.\n")
	}

	if debug {
		log.Printf("Analyzing branch %s\n", branch)
	}

	poms, err := FindPoms(".")
	if err != nil {
		return err
	}

	if len(poms) == 0 {
		return fmt.Errorf("Cannot find POMs\n")
	}

	for _, pomFile := range poms {
		if debug {
			log.Printf("Analyzing %s\n", pomFile)
		}

		if _, present := excludesMap[pomFile]; present {
			if debug {
				log.Printf("Skipping excluded pom: %s\n", pomFile)
			}
			continue
		}

		effectiveVersion, err := PomVersion(pomFile)
		if err != nil {
			return err
		}
		if strings.HasPrefix(effectiveVersion, "$") {
			log.Printf("Skipping pom %s because of unresolvable token %s in version element\n", pomFile, effectiveVersion)
			continue
		}

		if branch == "develop" {
			if !IsValidDevelopVersion(effectiveVersion) {
				return fmt.Errorf("Invalid develop branch version %s in %s\n", effectiveVersion, pomFile)
			}
			continue
		}
		if !IsValidFeatureVersion(branch, effectiveVersion) {
			return fmt.Errorf("Feature branch %s has invalid version %s in %s\n", branch, effectiveVersion, pomFile)
		}
	}
	return nil
}

func CurrentBranch() (string, error) {
	if stdout, _, err := Exec("git", "rev-parse", "--abbrev-ref", "HEAD"); err != nil {
		return "", err
	} else {
		return string(stdout[:len(stdout)-1]), nil
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

func FindPoms(dir string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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

func Exec(cmd string, args ...string) ([]byte, []byte, error) {
	if debug {
		log.Printf("%s %v\n", cmd, args)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	command := exec.Command(cmd, args...)
	command.Stdout = stdout
	command.Stderr = stderr

	err := command.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func GitFetch() error {
	if _, _, err := Exec("git", "fetch"); err != nil {
		return err
	}
	return nil
}

func GitStash() error {
	if _, _, err := Exec("git", "stash", "--include-untracked"); err != nil {
		return err
	}
	return nil
}

func GitCheckoutBranch(branchName string) error {
	if _, _, err := Exec("git", "checkout", branchName); err != nil {
		return err
	}
	return nil
}

func GetBranches() ([]string, error) {
	if stdout, _, err := Exec("git", "ls-remote", "--heads"); err != nil {
		return nil, err
	} else {
		r := make([]string, 0)
		readbuffer := bytes.NewBuffer(stdout)
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

/*
DupCheck returns a version-indexed map of slices of branches.  An entry t[s] returns a slice of all branch names with POM version s.
*/
func DupCheck() (map[string][]string, error) {
	if err := GitFetch(); err != nil {
		return nil, fmt.Errorf("Error in git-fetch: %v\n", err)
	}

	branches, err := GetBranches()
	if err != nil {
		return nil, fmt.Errorf("Error getting remote heads: %v\n", err)
	}

	t := make(map[string][]string)
	for _, branch := range branches {
		if err := GitStash(); err != nil {
			return nil, fmt.Errorf("Cannot stash to clean workspace on branch %s: %v\n", branch, err)
		}
		if err := GitCheckoutBranch(branch); err != nil {
			return nil, fmt.Errorf("Cannot checkout branch %s: %v\n", branch, err)
		}
		effectiveVersion, err := PomVersion("pom.xml")
		if err != nil {
			return nil, err
		}
		_, present := t[effectiveVersion]
		if !present {
			t[effectiveVersion] = make([]string, 0)
		}
		t[effectiveVersion] = append(t[effectiveVersion], branch)
	}
	return t, nil
}

func WalkToGitRoot(dir string) (string, error) {
	fileInfo, err := os.Stat(dir + "/.git")
	if err != nil {
		if os.IsNotExist(err) {
			return WalkToGitRoot(dir + "/..")
		} else {
			return "", err
		}
	}
	if fileInfo == nil {
		return WalkToGitRoot(dir + "/..")
	}
	if fileInfo.IsDir() {
		return dir, nil
	} else {
		// weird that a .git inode exists but it is not a directory
		return WalkToGitRoot(dir + "/..")
	}
}

func PomVersion(pomFile string) (string, error) {
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

	if debug {
		log.Printf("effectiveVersion %s in pom %s\n", effectiveVersion, pomFile)
	}

	return effectiveVersion, nil
}
