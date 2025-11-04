package commands

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/codecrafters-io/cli/internal/client"
	"github.com/codecrafters-io/cli/internal/globals"
	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/getsentry/sentry-go"
)

func SubmitCommand() (err error) {
	utils.Logger.Debug().Msg("submit command starts")

	defer func() {
		utils.Logger.Debug().Err(err).Msg("submit command ends")
	}()

	defer func() {
		if p := recover(); p != nil {
			utils.Logger.Panic().Str("panic", fmt.Sprintf("%v", p)).Stack().Msg("panic")
			sentry.CurrentHub().Recover(p)

			panic(p)
		}

		if err == nil {
			return
		}

		var noRepo utils.NoCodecraftersRemoteFoundError
		if errors.Is(err, &noRepo) {
			// ignore
			return
		}

		sentry.CurrentHub().CaptureException(err)
	}()

	utils.Logger.Debug().Msg("computing repository directory")

	repoDir, err := utils.GetRepositoryDir()
	if err != nil {
		return err
	}

	utils.Logger.Debug().Msgf("found repository directory: %s", repoDir)

	utils.Logger.Debug().Msg("identifying remotes")

	codecraftersRemote, err := utils.IdentifyGitRemote(repoDir)
	if err != nil {
		return err
	}

	utils.Logger.Debug().Msgf("identified remote: %s, %s", codecraftersRemote.Name, codecraftersRemote.Url)

	currentBranchName, err := getCurrentBranch(repoDir)
	if err != nil {
		return fmt.Errorf("get current branch: %w", err)
	}

	defaultBranchName := "master" // TODO: Change when we allow customizing the defaultBranch

	if currentBranchName != defaultBranchName {
		return fmt.Errorf("You need to be on the `%s` branch to run this command.", defaultBranchName)
	}

	utils.Logger.Debug().Msgf("committing changes to %s", defaultBranchName)

	commitSha, err := commitChanges(repoDir, "codecrafters submit [skip ci]")
	if err != nil {
		return fmt.Errorf("commit changes: %w", err)
	}

	// Place this before the push so that it "feels" fast
	fmt.Printf("Submitting changes (commit: %s)...\n", commitSha[:7])

	err = pushBranchToRemote(repoDir, codecraftersRemote.Name)
	if err != nil {
		return fmt.Errorf("push changes: %w", err)
	}

	utils.Logger.Debug().Msgf("pushed changes to remote branch %s", defaultBranchName)

	globals.SetCodecraftersServerURL(codecraftersRemote.CodecraftersServerURL())
	codecraftersClient := client.NewCodecraftersClient()

	utils.Logger.Debug().Msgf("creating submission for %s", commitSha)

	createSubmissionResponse, err := codecraftersClient.CreateSubmission(codecraftersRemote.CodecraftersRepositoryId(), commitSha, "submit", "current_and_previous_descending")
	if err != nil {
		return fmt.Errorf("create submission: %w", err)
	}

	utils.Logger.Debug().Msgf("submission created: %v", createSubmissionResponse.Id)

	return handleSubmission(createSubmissionResponse, codecraftersClient)
}

func getCurrentBranch(repoDir string) (string, error) {
	outputBytes, err := exec.Command("git", "-C", repoDir, "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		return "", wrapError(err, outputBytes, "get current branch")
	}

	return strings.TrimSpace(string(outputBytes)), nil
}
