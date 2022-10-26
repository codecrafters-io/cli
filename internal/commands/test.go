package commands

import (
	"fmt"
	"github.com/codecrafters-io/cli/internal/custom_storage"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
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

	repository.Storer = custom_storage.NewCustomStorage(repository.Storer.(*filesystem.Storage).Filesystem(), cache.NewObjectLRUDefault())

	worktree, err := repository.Worktree()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read worktree: %s", err)
		return 1
	}

	err = worktree.AddWithOptions(&git.AddOptions{All: true})
	if err != nil {
		panic(err)
	}

	status, err := worktree.Status()
	if err != nil {
		panic(err)
	}

	fmt.Println(status)

	return 0
}
