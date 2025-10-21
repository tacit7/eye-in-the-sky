package app

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProjectInfo holds detected project/repository information
type ProjectInfo struct {
	RepoName    string // Repository name
	Owner       string // Repository owner
	Branch      string // Current git branch
	Commit      string // Current commit hash (short)
	RemoteURL   string // Remote origin URL
	GitRootPath string // Root path of git repository
}

// DetectProject detects the current project information from git
// Returns nil if not in a git repository
func DetectProject() *ProjectInfo {
	// Find git root
	gitRoot := findGitRoot(".")
	if gitRoot == "" {
		log.Println("[PROJECT] Not in a git repository")
		return nil
	}

	log.Printf("[PROJECT] Found git repository at: %s\n", gitRoot)

	info := &ProjectInfo{
		GitRootPath: gitRoot,
	}

	// Get remote URL
	remoteURL := getGitRemoteURL(gitRoot, "origin")
	if remoteURL != "" {
		info.RemoteURL = remoteURL
		// Extract repo name from URL
		info.RepoName = extractRepoName(remoteURL)
		// Extract owner from URL
		info.Owner = extractOwner(remoteURL)
	}

	// Get current branch
	info.Branch = getGitBranch(gitRoot)

	// Get current commit hash
	info.Commit = getGitCommit(gitRoot)

	log.Printf("[PROJECT] Detected: %s/%s (branch: %s, commit: %s)\n",
		info.Owner, info.RepoName, info.Branch, info.Commit)

	return info
}

// findGitRoot searches for a .git directory at or above the current path
func findGitRoot(startPath string) string {
	currentPath, err := os.Getwd()
	if err != nil {
		currentPath = startPath
	}

	for {
		gitPath := filepath.Join(currentPath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return currentPath
		}

		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			// Reached filesystem root
			return ""
		}
		currentPath = parent
	}
}

// getGitRemoteURL retrieves the URL for a git remote
func getGitRemoteURL(repoPath string, remoteName string) string {
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", remoteName)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("[PROJECT] Failed to get remote URL: %v\n", err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

// extractRepoName extracts repository name from git URL
// Handles both HTTPS and SSH URLs
func extractRepoName(url string) string {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH URLs: git@github.com:owner/repo
	if strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	// Handle HTTPS URLs: https://github.com/owner/repo
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return ""
}

// extractOwner extracts repository owner from git URL
func extractOwner(url string) string {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH URLs: git@github.com:owner/repo
	if strings.HasPrefix(url, "git@") {
		// git@github.com:owner/repo -> owner/repo
		parts := strings.Split(url, "/")
		if len(parts) >= 2 {
			// parts[-2] is the owner
			return parts[len(parts)-2]
		}
	}

	// Handle HTTPS URLs: https://github.com/owner/repo
	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		// parts[-2] is the owner
		return parts[len(parts)-2]
	}

	return ""
}

// getGitBranch retrieves the current git branch name
func getGitBranch(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("[PROJECT] Failed to get branch: %v\n", err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

// getGitCommit retrieves the current git commit hash (short form)
func getGitCommit(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("[PROJECT] Failed to get commit: %v\n", err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

// String returns a formatted string representation of ProjectInfo
func (p *ProjectInfo) String() string {
	if p == nil {
		return "No project detected"
	}

	parts := []string{}
	if p.Owner != "" && p.RepoName != "" {
		parts = append(parts, fmt.Sprintf("%s/%s", p.Owner, p.RepoName))
	}
	if p.Branch != "" {
		parts = append(parts, fmt.Sprintf("branch: %s", p.Branch))
	}
	if p.Commit != "" {
		parts = append(parts, fmt.Sprintf("commit: %s", p.Commit))
	}

	return strings.Join(parts, " | ")
}
