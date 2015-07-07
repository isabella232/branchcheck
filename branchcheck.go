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
	excludes        = flag.String("excludes", "", "comma-separated poms to exclude, by path relative to repository top level (e.g., a/pom.xml,b/pom.xml).  Used with branch-compat.")
	info            = flag.Bool("version", false, "Print git commit from which we were built")
	versionDupCheck = flag.Bool("version-dups", false, "Iterate over all branches and check for duplicate POM versions.  Uses git ls-remote to get remote branches.")
	branchCompat    = flag.Bool("branch-compat", true, "Verify branch name and POM versions are compatible.  If version-dups is set, branch compat will not be run.")
	pomVersion      = flag.Bool("pom-version", false, "Display POM version for ./pom.xml and exit.")
	debug           = flag.Bool("debug", false, "Log debug info to the console.")

	excludesMap map[string]string

	buildInfo string
)

func init() {
	flag.Parse()
	a := strings.Split(*excludes, ",")
	excludesMap = make(map[string]string, len(a))
	for _, v := range a {
		excludesMap[v] = ""
	}
}

func main() {
	log.Printf("branchcheck: %s\n", buildInfo)
	if *info {
		os.Exit(0)
	}

	if *pomVersion {
		if _, err := os.Stat(".git"); err != nil && os.IsNotExist(err) {
			log.Fatalf("This command must be run from the top level of the repository: %v\n", err)
		}

		data, err := ioutil.ReadFile("pom.xml")
		if err != nil {
			log.Fatalf("Error reading ./pom.xml: %v\n", err)
		}

		var pom POM
		reader := bytes.NewBuffer(data)
		if err := xml.NewDecoder(reader).Decode(&pom); err != nil {
			log.Fatalf("Error deserializing ./pom.xml to XML: %v\n", err)
		}

		if pom.Version == "" && pom.Parent.Version == "" {
			log.Fatalf("pom version and parent are both empty in pom.xml\n")
		}

		fmt.Printf("pom version: %s\n", pom.Version)

		return
	}

	if *versionDupCheck {
		CurrentBranch, err := CurrentBranch()
		if err != nil {
			log.Fatalf("Cannot get current branch; %v\n", err)
		}
		defer GitCheckoutBranch(CurrentBranch)

		if versionOccurrenceMap, err := DupCheck(); err != nil {
			GitCheckoutBranch(CurrentBranch)
			log.Fatalf("error running main.DupCheck(): %v\n", err)
		} else {
			someMultiples := false
			for k, v := range versionOccurrenceMap {
				if len(v) > 1 {
					someMultiples = true
					log.Printf("multiple branches %+v with version %s\n", v, k)
				}
			}
			GitCheckoutBranch(CurrentBranch)
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

	if *debug {
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
		if *debug {
			log.Printf("Analyzing %s\n", pomFile)
		}

		if _, present := excludesMap[pomFile]; present {
			if *debug {
				log.Printf("Skipping excluded pom: %s\n", pomFile)
			}
			continue
		}

		effectiveVersion, err := PomVersion(pomFile)
		if err != nil {
			return err
		}
		if strings.HasPrefix(effectiveVersion, "$") {
			if *debug {
				log.Printf("Skipping pom %s because of unresolvable token %s in version element\n", pomFile, effectiveVersion)
			}
			continue
		}

		if branch == "develop" {
			if !IsValidDevelopVersion(effectiveVersion) {
				return fmt.Errorf("Invalid develop branch version %s in %s\n", effectiveVersion, pomFile)
			}
			continue
		}
		if !IsBranchVersionCompatible(branch, effectiveVersion) {
			return fmt.Errorf("Branch %s has invalid version %s in %s\n", branch, effectiveVersion, pomFile)
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

func IsBranchVersionCompatible(branch, version string) bool {
	// For a branch feature/ABC-2, the branch prefix == feature and the story == ABC-2
	branchPrefix, story, ok := branchParts(branch)
	if !ok {
		log.Printf("Branch name %s is malformed.  Valid branch names have the form [feature|hotfix]/<storypart>.\n", branch)
		return false
	}

	// For a version 1.0-abc_2-SNAPSHOT, truncatedVersion is 1.0-abc_2
	truncatedVersion, ok := truncateVersion(version)
	if !ok {
		log.Printf("POM version %s does not end in -SNAPSHOT.  This is not a feature branch.\n", version)
		return false
	}

	switch branchPrefix {
	case "feature":
		// For a story ABC-2, normalizedStory is abc_2
		normalizedStory := normalizeStory(story)

		// The branch validates if 1.0-abc_2 has-suffix abc_2 respecting case
		validates := strings.HasSuffix(truncatedVersion, normalizedStory)
		if !validates {
			log.Printf("feature/ branch %s fails validation.  jgitflow would have lowered the case of the POM <version> %s and replaced - with _.\n", branch, version)
			log.Printf("See https://xoomcorp.atlassian.net/wiki/display/Eng/branchcheck%3A++A+tool+for+feature+branch+developers\n")
		}
		return validates
	case "hotfix":
		// The branch validates if 1.0-abc-2 has-suffix abc-2 independent of case
		validates := strings.HasSuffix(strings.ToLower(truncatedVersion), strings.ToLower(story))
		if !validates {
			log.Printf("hotfix/ branch %s fails validation.  jgitflow would have preserved hotfix name case in POM <version> %s and retained uses of '-'.\n", branch, version)
			log.Printf("See https://xoomcorp.atlassian.net/wiki/display/Eng/branchcheck%3A++A+tool+for+feature+branch+developers\n")
		}
		return validates
	}
	log.Printf("Unknown branch prefix: %s\n", branchPrefix)
	return false
}

func truncateVersion(version string) (string, bool) {
	if !strings.HasSuffix(version, "-SNAPSHOT") {
		return "", false
	}
	return version[0:strings.Index(version, "-SNAPSHOT")], true
}

func normalizeStory(story string) string {
	s := strings.ToLower(story)
	return strings.Replace(s, "-", "_", -1)
}

func branchParts(branch string) (string, string, bool) {
	parts := strings.Split(branch, "/")
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
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
	if *debug {
		log.Printf("found %d poms\n", len(files))
	}
	return files, nil
}

func Exec(cmd string, args ...string) ([]byte, []byte, error) {
	if *debug {
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
	if _, stderr, err := Exec("git", "fetch"); err != nil {
		log.Printf("git-fetch stderr: %s\n", string(stderr))
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
		if *debug {
			log.Printf("Using parent-version in pom %s\n", pomFile)
		}
	} else {
		effectiveVersion = pom.Version
	}

	if *debug {
		log.Printf("effectiveVersion %s in pom %s\n", effectiveVersion, pomFile)
	}

	return effectiveVersion, nil
}
