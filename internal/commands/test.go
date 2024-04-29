package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/cli/internal/utils"
	logstream_consumer "github.com/codecrafters-io/logstream/consumer"
	"github.com/fatih/color"
	"github.com/getsentry/sentry-go"
	cp "github.com/otiai10/copy"
	"github.com/rs/zerolog"
)

func TestCommand(ctx context.Context) (err error) {
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
		return fmt.Errorf("find repository root folder: %w", err)
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

	err = pushBranchToRemote(tmpDir)
	if err != nil {
		return fmt.Errorf("push changes: %w", err)
	}

	logger.Debug().Msgf("pushed changes to remote branch %s", tempBranchName)

	codecraftersClient := utils.NewCodecraftersClient(codecraftersRemote.CodecraftersServerURL())

	logger.Debug().Msgf("creating submission for %s", tempCommitSha)

	createSubmissionResponse, err := codecraftersClient.CreateSubmission(codecraftersRemote.CodecraftersRepositoryId(), tempCommitSha)
	if err != nil {
		return fmt.Errorf("create submission: %w", err)
	}

	logger.Debug().Msgf("submission created: %v", createSubmissionResponse.Id)

	for _, message := range createSubmissionResponse.OnInitMessages {
		fmt.Println("")
		message.Print()
	}

	if createSubmissionResponse.BuildLogstreamURL != "" {
		logger.Debug().Msgf("streaming build logs from %s", createSubmissionResponse.BuildLogstreamURL)

		fmt.Println("")
		err = streamLogs(createSubmissionResponse.BuildLogstreamURL)
		if err != nil {
			return fmt.Errorf("stream build logs: %w", err)
		}

		logger.Debug().Msg("Finished streaming build logs")
		logger.Debug().Msg("fetching build")

		fetchBuildResponse, err := codecraftersClient.FetchBuild(createSubmissionResponse.BuildID)
		if err != nil {
			// TODO: Notify sentry
			red := color.New(color.FgRed).SprintFunc()
			fmt.Fprintln(os.Stderr, red(err.Error()))
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, red("We couldn't fetch the results of your submission. Please try again?"))
			fmt.Fprintln(os.Stderr, red("Let us know at hello@codecrafters.io if this error persists."))
			return err
		}

		logger.Debug().Msgf("finished fetching build: %v", fetchBuildResponse)
		red := color.New(color.FgRed).SprintFunc()

		switch fetchBuildResponse.Status {
		case "failure":
			fmt.Fprintln(os.Stderr, red(""))
			fmt.Fprintln(os.Stderr, red("Looks like your codebase failed to build."))
			fmt.Fprintln(os.Stderr, red("If you think this is a CodeCrafters error, please let us know at hello@codecrafters.io."))
			fmt.Fprintln(os.Stderr, red(""))
			os.Exit(0)
		case "success":
			time.Sleep(1 * time.Second) // The delay in-between build and test logs is usually 5-10 seconds, so let's buy some time
		default:
			red := color.New(color.FgRed).SprintFunc()

			fmt.Fprintln(os.Stderr, red("We couldn't fetch the results of your build. Please try again?"))
			fmt.Fprintln(os.Stderr, red("Let us know at hello@codecrafters.io if this error persists."))
			os.Exit(1)
		}
	}

	fmt.Println("")
	fmt.Println("Running tests. Logs should appear shortly...")
	fmt.Println("")

	err = streamLogs(createSubmissionResponse.LogstreamURL)
	if err != nil {
		return fmt.Errorf("stream logs: %w", err)
	}

	logger.Debug().Msgf("fetching submission %s", createSubmissionResponse.Id)

	fetchSubmissionResponse, err := codecraftersClient.FetchSubmission(createSubmissionResponse.Id)
	if err != nil {
		// TODO: Notify sentry
		red := color.New(color.FgRed).SprintFunc()
		fmt.Fprintln(os.Stderr, red(err.Error()))
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, red("We couldn't fetch the results of your submission. Please try again?"))
		fmt.Fprintln(os.Stderr, red("Let us know at hello@codecrafters.io if this error persists."))
		return err
	}

	logger.Debug().Msgf("finished fetching submission, status: %s", fetchSubmissionResponse.Status)

	switch fetchSubmissionResponse.Status {
	case "failure":
		for _, message := range createSubmissionResponse.OnFailureMessages {
			fmt.Println("")
			message.Print()
		}
	case "success":
		for _, message := range createSubmissionResponse.OnSuccessMessages {
			fmt.Println("")
			message.Print()
		}
	default:
		fmt.Println("")
	}

	if fetchSubmissionResponse.IsError {
		return fmt.Errorf("%s", fetchSubmissionResponse.ErrorMessage)
	}

	return nil
}

func copyRepositoryDirToTempDir(repoDir string) (string, error) {
	tmpDir, err := ioutil.TempDir("", "codecrafters")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	err = cp.Copy(repoDir, tmpDir)
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

func pushBranchToRemote(tmpDir string) error {
	// TODO: Find CodeCrafters remote and use that
	outputBytes, err := exec.Command("git", "-C", tmpDir, "push", "origin", "HEAD").CombinedOutput()
	if err != nil {
		return wrapError(err, outputBytes, "run git command")
	}

	return nil
}

func streamLogs(logstreamUrl string) error {
	consumer, err := logstream_consumer.NewConsumer(logstreamUrl, func(message string) {})
	if err != nil {
		return fmt.Errorf("new log consumer: %w", err)
	}

	_, err = io.Copy(os.Stdout, consumer)
	if err != nil {
		return fmt.Errorf("stream data: %w", err)
	}

	return nil
}

func wrapError(err error, output []byte, msg string) error {
	if _, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("add all files: %s", output)
	}

	return fmt.Errorf("add all files: %w", err)
}
