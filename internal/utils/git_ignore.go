// Package utils
package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

type GitIgnore struct {
	baseDir         string
	localGitIgnore  *ignore.GitIgnore
	globalGitIgnore *ignore.GitIgnore
	gitInfoExclude  *ignore.GitIgnore
}

func NewGitIgnore(baseDir string) GitIgnore {
	return GitIgnore{
		baseDir:         baseDir,
		localGitIgnore:  compileIgnorer(filepath.Join(baseDir, ".gitignore")),
		globalGitIgnore: compileIgnorer(getGlobalGitIgnorePath()),
		gitInfoExclude:  compileIgnorer(filepath.Join(baseDir, ".git", "info", "exclude")),
	}
}

func (i GitIgnore) SkipFile(path string) (bool, error) {
	for _, ignorer := range []*ignore.GitIgnore{i.localGitIgnore, i.globalGitIgnore, i.gitInfoExclude} {
		if ignorer != nil && ignorer.MatchesPath(path) {
			return true, nil
		}
	}

	return false, nil
}

func compileIgnorer(path string) *ignore.GitIgnore {
	ignorer, err := ignore.CompileIgnoreFile(path)
	if err != nil {
		return nil
	}

	return ignorer
}

func getGlobalGitIgnorePath() string {
	output, err := exec.Command("git", "config", "--get", "core.excludesfile").Output()
	if err != nil {
		return ""
	}

	path := strings.TrimSpace(string(output))
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		path = filepath.Join(homeDir, path[2:])
	}

	return path
}
