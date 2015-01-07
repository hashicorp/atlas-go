package archive

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestVCSPreflight(t *testing.T) {
	if !testHasGit {
		t.Skip("git not found")
	}

	testDir := testFixture("archive-git")
	oldName := filepath.Join(testDir, "DOTgit")
	newName := filepath.Join(testDir, ".git")
	if err := os.Rename(oldName, newName); err != nil {
		t.Fatal(err)
	}
	defer os.Rename(newName, oldName)

	if err := vcsPreflight(testDir); err != nil {
		t.Fatal(err)
	}
}

func TestGitBranch(t *testing.T) {
	if !testHasGit {
		t.Skip("git not found")
	}

	testDir := testFixture("archive-git")
	oldName := filepath.Join(testDir, "DOTgit")
	newName := filepath.Join(testDir, ".git")
	if err := os.Rename(oldName, newName); err != nil {
		t.Fatal(err)
	}
	defer os.Rename(newName, oldName)

	branch, err := gitBranch(testDir)
	if err != nil {
		t.Fatal(err)
	}

	expected := "master"
	if branch != expected {
		t.Fatalf("expected %q to be %q", branch, expected)
	}
}

func TestGitCommit(t *testing.T) {
	if !testHasGit {
		t.Skip("git not found")
	}

	testDir := testFixture("archive-git")
	oldName := filepath.Join(testDir, "DOTgit")
	newName := filepath.Join(testDir, ".git")
	if err := os.Rename(oldName, newName); err != nil {
		t.Fatal(err)
	}
	defer os.Rename(newName, oldName)

	commit, err := gitCommit(testDir)
	if err != nil {
		t.Fatal(err)
	}

	expected := "7525d17cbbb56f3253a20903ffddc07c6c935c76"
	if commit != expected {
		t.Fatalf("expected %q to be %q", commit, expected)
	}
}

func TestGitRemotes(t *testing.T) {
	if !testHasGit {
		t.Skip("git not found")
	}

	testDir := testFixture("archive-git")
	oldName := filepath.Join(testDir, "DOTgit")
	newName := filepath.Join(testDir, ".git")
	if err := os.Rename(oldName, newName); err != nil {
		t.Fatal(err)
	}
	defer os.Rename(newName, oldName)

	remotes, err := gitRemotes(testDir)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"remote.origin":   "https://github.com/hashicorp/origin.git",
		"remote.upstream": "https://github.com/hashicorp/upstream.git",
	}

	if !reflect.DeepEqual(remotes, expected) {
		t.Fatalf("expected %+v to be %+v", remotes, expected)
	}
}

func TestVCSMetadata_git(t *testing.T) {
	if !testHasGit {
		t.Skip("git not found")
	}

	testDir := testFixture("archive-git")
	oldName := filepath.Join(testDir, "DOTgit")
	newName := filepath.Join(testDir, ".git")
	if err := os.Rename(oldName, newName); err != nil {
		t.Fatal(err)
	}
	defer os.Rename(newName, oldName)

	metadata, err := vcsMetadata(testDir)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"branch":          "master",
		"commit":          "7525d17cbbb56f3253a20903ffddc07c6c935c76",
		"remote.origin":   "https://github.com/hashicorp/origin.git",
		"remote.upstream": "https://github.com/hashicorp/upstream.git",
	}

	if !reflect.DeepEqual(metadata, expected) {
		t.Fatalf("expected %+v to be %+v", metadata, expected)
	}
}
