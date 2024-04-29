package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func GetRepositoryDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get workdir: %w", err)
	}

	outputBytes, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			if regexp.MustCompile("not a git repository").Match(outputBytes) {
				fmt.Fprintf(os.Stderr, "The current directory is not within a Git repository.\n")
				fmt.Fprintf(os.Stderr, "Please run this command from within your CodeCrafters Git repository.\n")

				return "", errors.New("used not in a repository")
			}
		}

		return "", fmt.Errorf("failed to run 'git rev-parse' to get repository dir. err: %v.\n%s", err, string(outputBytes))
	}

	return strings.TrimSpace(string(outputBytes)), nil
}
