package commands

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/go-git/go-git/v5/storage/transactional"
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

	repository.Storer = transactional.NewStorage(repository.Storer, memory.NewStorage())

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

	err = worktree.AddWithOptions(&git.AddOptions{All: true})
	if err != nil {
		panic(err)
	}

	status, err = worktree.Status()
	if err != nil {
		panic(err)
	}

	fmt.Println(status)

	commitHash, err := worktree.Commit("test", &git.CommitOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println(commitHash.String())

	err = worktree.Checkout(&git.CheckoutOptions{Keep: true, Branch: "test", Create: true})
	if err != nil {
		panic(err)
	}

	err = repository.Push(&git.PushOptions{
		RemoteName: "origin",
	})
	if err != nil {
		panic(err)
	}

	return 0
}
