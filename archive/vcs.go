package archive

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// VCS is a struct that explains how to get the file list for a given
// VCS.
type VCS struct {
	Name string

	// Detect is a list of files/folders that if they exist, signal that
	// this VCS is the VCS in use.
	Detect []string

	// Files returns the files that are under version control for the
	// given path.
	Files func(path string) ([]string, error)
}

// VCSList is the list of VCS we recognize.
var VCSList = []*VCS{
	&VCS{
		Name:   "git",
		Detect: []string{".git/"},
		Files:  vcsFilesCmd("git", "ls-files"),
	},
	&VCS{
		Name:   "hg",
		Detect: []string{".hg/"},
		Files:  vcsFilesCmd("hg", "locate --include ."),
	},
}

// vcsDetect detects the VCS that is used for path.
func vcsDetect(path string) (*VCS, error) {
	for _, v := range VCSList {
		for _, f := range v.Detect {
			check := filepath.Join(path, f)
			if _, err := os.Stat(check); err == nil {
				return v, nil
			}
		}
	}

	return nil, nil
}

// vcsFiles returns the files for the VCS directory path.
func vcsFiles(path string) ([]string, error) {
	vcs, err := vcsDetect(path)
	if err != nil {
		return nil, fmt.Errorf("Error detecting VCS: %s", err)
	}
	if vcs == nil {
		return nil, fmt.Errorf("No VCS found for path: %s", path)
	}

	return vcs.Files(path)
}

// vcsFilesCmd creates a Files-compatible function that reads the files
// by executing the command in the repository path and returning each
// line in stdout.
func vcsFilesCmd(args ...string) func(string) ([]string, error) {
	return func(path string) ([]string, error) {
		var stderr, stdout bytes.Buffer

		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = path
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf(
				"Error executing %s: %s",
				strings.Join(args, " "),
				err)
		}

		// Read each line of output as a path
		result := make([]string, 0, 100)
		scanner := bufio.NewScanner(&stdout)
		for scanner.Scan() {
			result = append(result, scanner.Text())
		}

		return result, nil
	}
}
