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
	Files VCSFilesFunc

	// Metadata returns arbitrary metadata about the underlying VCS for the
	// given path.
	Metadata VCSMetadataFunc
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
		Files:  vcsTrimCmd(vcsFilesCmd("hg", "locate", "-f", "--include", ".")),
	},
	&VCS{
		Name:   "svn",
		Detect: []string{".svn/"},
		Files:  vcsFilesCmd("svn", "ls"),
	},
}

// VCSFilesFunc is the callback invoked to return the files in the VCS.
//
// The return value should be paths relative to the given path.
type VCSFilesFunc func(string) ([]string, error)

// VCSMetadataFunc is the callback invoked to get arbitrary information about
// the current VCS.
//
// The return value should be a map of key-value pairs.
type VCSMetadataFunc func(string) (map[string]string, error)

// vcsDetect detects the VCS that is used for path.
func vcsDetect(path string) (*VCS, error) {
	for _, v := range VCSList {
		dir := path
		for {
			for _, f := range v.Detect {
				check := filepath.Join(dir, f)
				if _, err := os.Stat(check); err == nil {
					return v, nil
				}
			}

			lastDir := dir
			dir = filepath.Dir(dir)
			if dir == lastDir {
				break
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
func vcsFilesCmd(args ...string) VCSFilesFunc {
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

// vcsTrimCmd trims the prefix from the paths returned by another VCSFilesFunc.
// This should be used to wrap another function if the return value is known
// to have full paths rather than relative paths
func vcsTrimCmd(f VCSFilesFunc) VCSFilesFunc {
	return func(path string) ([]string, error) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf(
				"Error expanding VCS path: %s", err)
		}

		// Now that we have the root path, get the inner files
		fs, err := f(path)
		if err != nil {
			return nil, err
		}

		// Trim the root path from the files
		result := make([]string, 0, len(fs))
		for _, f := range fs {
			if !strings.HasPrefix(f, absPath) {
				continue
			}

			f, err = filepath.Rel(absPath, f)
			if err != nil {
				return nil, fmt.Errorf(
					"Error determining path: %s", err)
			}

			result = append(result, f)
		}

		return result, nil
	}
}
