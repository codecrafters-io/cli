package commands

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
)

func SubmitCommand(ctx context.Context) (err error) {
	logger := zerolog.Ctx(ctx)

	logger.Debug().Msg("submit command starts")
	defer func() {
		logger.Debug().Err(err).Msg("submit command ends")
	}()

	defer func() {
		if p := recover(); p != nil {
			logger.Panic().Str("panic", fmt.Sprintf("%v", p)).Stack().Msg("panic")
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

	logger.Debug().Msg("computing repository directory")

	repoDir, err := utils.GetRepositoryDir()
	if err != nil {
		return err
	}

	logger.Debug().Msgf("found repository directory: %s", repoDir)

	logger.Debug().Msg("identifying remotes")

	codecraftersRemote, err := utils.IdentifyGitRemote(repoDir)
	if err != nil {
		return err
	}

	logger.Debug().Msgf("identified remote: %s, %s", codecraftersRemote.Name, codecraftersRemote.Url)

	currentBranchName, err := getCurrentBranch(repoDir)
	if err != nil {
		return fmt.Errorf("get current branch: %w", err)
	}

	defaultBranchName := "master" // TODO: Change when we allow customizing the defaultBranch

	if currentBranchName != defaultBranchName {
		return fmt.Errorf("You need to be on the `%s` branch to run this command.", defaultBranchName)
	}

	logger.Debug().Msgf("committing changes to %s", defaultBranchName)

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

	logger.Debug().Msgf("pushed changes to remote branch %s", defaultBranchName)

	codecraftersClient := utils.NewCodecraftersClient(codecraftersRemote.CodecraftersServerURL())

	logger.Debug().Msgf("creating submission for %s", commitSha)

	createSubmissionResponse, err := codecraftersClient.CreateSubmission(codecraftersRemote.CodecraftersRepositoryId(), commitSha, "submit", "current_and_previous_descending")
	if err != nil {
		return fmt.Errorf("create submission: %w", err)
	}

	logger.Debug().Msgf("submission created: %v", createSubmissionResponse.Id)

	return utils.HandleSubmission(createSubmissionResponse, ctx, codecraftersClient)
}

func getCurrentBranch(repoDir string) (string, error) {
	outputBytes, err := exec.Command("git", "-C", repoDir, "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		return "", wrapError(err, outputBytes, "get current branch")
	}

	return strings.TrimSpace(string(outputBytes)), nil
}
