package commands

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"os"
)

func TestCommand() int {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch current working directory: %s", err)
		return 1
	}

	repository, err := git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read git repository: %s", err)
		return 1
	}

	worktree, err := repository.Worktree()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read worktree: %s", err)
		return 1
	}

	status, err := worktree.Status()
	if err != nil {
		panic(err)
	}

	fmt.Println(status)

	return 0
}
