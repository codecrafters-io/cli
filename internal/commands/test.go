package commands

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"os"
)

func TestCommand() int {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch current working directory: %s", err)
		return 1
	}

	repository, err := git.PlainOpen(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read git repository: %s", err)
		return 1
	}

	branches, err := repository.Branches()
	if err != nil {
		panic(err)
	}

	branches.ForEach(func(branch *plumbing.Reference) error {
		fmt.Println(branch.Name())
		return nil
	})

	return 0
}
