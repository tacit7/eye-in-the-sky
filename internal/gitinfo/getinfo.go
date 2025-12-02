package gitinfo

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// get runs a git command and trims the output.
func get(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// RepoName returns the basename of the current git repository folder.
func RepoName() (string, error) {
	toplevel, err := get("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return filepath.Base(toplevel), nil
}

// RepoPath returns the full path to the current git repository folder.
func RepoPath() (string, error) {
	return get("rev-parse", "--show-toplevel")
}

// RemoteURL returns the URL of the 'origin' remote.
func RemoteURL() (string, error) {
	return get("config", "--get", "remote.origin.url")
}

// Branch returns the current checked-out branch.
func Branch() (string, error) {
	return get("rev-parse", "--abbrev-ref", "HEAD")
}

// CommitHash returns the full commit SHA.
func CommitHash() (string, error) {
	return get("rev-parse", "HEAD")
}

// ShortCommit returns the short 7-character commit SHA.
func ShortCommit() (string, error) {
	return get("rev-parse", "--short", "HEAD")
}

// Summary bundles all metadata into a map.
func Summary() (map[string]string, error) {
	repo, err := RepoName()
	if err != nil {
		return nil, err
	}
	remote, _ := RemoteURL()
	branch, _ := Branch()
	commit, _ := ShortCommit()
	path, _ := RepoPath()

	return map[string]string{
		"repo":   repo,
		"path":   path,
		"remote": remote,
		"branch": branch,
		"commit": commit,
	}, nil
}
