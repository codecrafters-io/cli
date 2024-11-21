package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/getsentry/sentry-go"
	cp "github.com/otiai10/copy"
	"github.com/rs/zerolog"
)

func TestCommand(ctx context.Context, shouldTestPrevious bool) (err error) {
	logger := zerolog.Ctx(ctx)

	logger.Debug().Msg("test command starts")
	defer func() {
		logger.Debug().Err(err).Msg("test command ends")
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

	logger.Debug().Msg("copying repository to temp directory")

	tmpDir, err := copyRepositoryDirToTempDir(repoDir)
	if err != nil {
		return fmt.Errorf("make a repo temp copy: %w", err)
	}

	defer os.RemoveAll(tmpDir)

	logger.Debug().Msgf("copied repository to temp directory: %s", tmpDir)

	tempBranchName := "cli-test-" + strconv.FormatInt(time.Now().UnixMilli(), 10)

	logger.Debug().Msgf("creating temp branch: %s", tempBranchName)

	err = checkoutNewBranch(tempBranchName, tmpDir)
	if err != nil {
		return fmt.Errorf("create temp branch: %w", err)
	}

	logger.Debug().Msgf("committing changes to %s", tempBranchName)

	tempCommitSha, err := commitChanges(tmpDir, fmt.Sprintf("CLI tests (%s)", tempBranchName))
	if err != nil {
		return fmt.Errorf("commit changes: %w", err)
	}

	// Place this before the push so that it "feels" fast
	fmt.Println("Initiating test run...")

	err = pushBranchToRemote(tmpDir, codecraftersRemote.Name)
	if err != nil {
		return fmt.Errorf("push changes: %w", err)
	}

	logger.Debug().Msgf("pushed changes to remote branch %s", tempBranchName)

	codecraftersClient := utils.NewCodecraftersClient(codecraftersRemote.CodecraftersServerURL())

	logger.Debug().Msgf("creating submission for %s", tempCommitSha)

	stageSelectionStrategy := "current_and_previous_descending"

	if shouldTestPrevious {
		stageSelectionStrategy = "current_and_previous_ascending"
	}

	createSubmissionResponse, err := codecraftersClient.CreateSubmission(codecraftersRemote.CodecraftersRepositoryId(), tempCommitSha, "test", stageSelectionStrategy)
	if err != nil {
		return fmt.Errorf("create submission: %w", err)
	}

	logger.Debug().Msgf("submission created: %v", createSubmissionResponse.Id)

	return utils.HandleSubmission(createSubmissionResponse, ctx, codecraftersClient)
}

func copyRepositoryDirToTempDir(repoDir string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "codecrafters")

	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	gitIgnore := utils.NewGitIgnore(repoDir)

	err = cp.Copy(repoDir, tmpDir, cp.Options{
		Skip: gitIgnore.SkipFile,
	})
	if err != nil {
		return "", fmt.Errorf("copy files: %w", err)
	}

	return tmpDir, nil
}

func checkoutNewBranch(tempBranchName string, tmpDir string) error {
	outputBytes, err := exec.Command("git", "-C", tmpDir, "checkout", "-b", tempBranchName).CombinedOutput()
	if err != nil {
		return wrapError(err, outputBytes, "run git command")
	}

	return nil
}

func commitChanges(tmpDir string, commitMessage string) (string, error) {
	// This is statically injected failure.
	// Set SENTRY_DEBUG_FAULT=commitChanges to produce error here.
	// Or set SENTRY_DEBUG_FAULT=commitChanges=panic to produce panic.
	if v, ok := os.LookupEnv("SENTRY_DEBUG_FAULT"); ok && strings.HasPrefix(v, "commitChanges") {
		if _, r, _ := strings.Cut(v, "="); r == "panic" {
			panic("test sentry panic")
		}

		return "", errors.New("test sentry error")
	}

	outputBytes, err := exec.Command("git", "-C", tmpDir, "add", ".").CombinedOutput()
	if err != nil {
		return "", wrapError(err, outputBytes, "add all files")
	}

	outputBytes, err = exec.Command("git", "-C", tmpDir, "commit", "--allow-empty", "-a", "-m", commitMessage).CombinedOutput()
	if err != nil {
		return "", wrapError(err, outputBytes, "create commit")
	}

	outputBytes, err = exec.Command("git", "-C", tmpDir, "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return "", wrapError(err, outputBytes, "get commit hash")
	}

	return strings.TrimSpace(string(outputBytes)), nil
}

func pushBranchToRemote(tmpDir string, remoteName string) error {
	outputBytes, err := exec.Command("git", "-C", tmpDir, "push", remoteName, "HEAD").CombinedOutput()
	if err != nil {
		return wrapError(err, outputBytes, "run git command")
	}

	return nil
}

func wrapError(err error, output []byte, msg string) error {
	if _, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("add all files: %s", output)
	}

	return fmt.Errorf("add all files: %w", err)
}
