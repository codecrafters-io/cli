package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/cli/internal/utils"
	logstream_consumer "github.com/codecrafters-io/logstream/consumer"
	"github.com/fatih/color"
	"github.com/getsentry/sentry-go"
	wordwrap "github.com/mitchellh/go-wordwrap"
	cp "github.com/otiai10/copy"
	"github.com/rs/zerolog"
)

func TestCommand(ctx context.Context) (err error) {
	logger := zerolog.Ctx(ctx)

	logger.Debug().Msg("start command")
	defer func() {
		logger.Debug().Err(err).Msg("end command")
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

	logger.Debug().Msg("get repo root")

	repoDir, err := GetRepositoryDir()
	if err != nil {
		return fmt.Errorf("find repository root folder: %w", err)
	}

	logger.Debug().Msg("find remotes")

	codecraftersRemote, err := utils.IdentifyGitRemote(repoDir)
	if err != nil {
		return err
	}

	logger.Debug().Msg("copy repo")

	tmpDir, err := copyRepositoryDirToTempDir(repoDir)
	if err != nil {
		return fmt.Errorf("make a repo temp copy: %w", err)
	}

	tempBranchName := "cli-test-" + strconv.FormatInt(time.Now().UnixMilli(), 10)

	logger.Debug().Msg("create temp branch")

	err = checkoutNewBranch(tempBranchName, tmpDir)
	if err != nil {
		return fmt.Errorf("create temp branch: %w", err)
	}

	logger.Debug().Msg("commit changes")

	tempCommitSha, err := commitChanges(tmpDir, fmt.Sprintf("CLI tests (%s)", tempBranchName))
	if err != nil {
		return fmt.Errorf("commit changes: %w", err)
	}

	// Place this before the push so that it "feels" fast
	fmt.Println("Running tests on your codebase. Streaming logs...")

	logger.Debug().Msg("push changes")

	err = pushBranchToRemote(tmpDir)
	if err != nil {
		return fmt.Errorf("push changes: %w", err)
	}

	logger.Debug().Msg("create codecrafters client")

	codecraftersClient := utils.NewCodecraftersClient(codecraftersRemote.CodecraftersServerURL())

	logger.Debug().Msg("create submission")

	createSubmissionResponse, err := codecraftersClient.CreateSubmission(codecraftersRemote.CodecraftersRepositoryId(), tempCommitSha)
	if err != nil {
		return fmt.Errorf("create submission: %w", err)
	}

	logger.Debug().Interface("response", createSubmissionResponse).Msg("submission created")

	if createSubmissionResponse.OnInitSuccessMessage != "" {
		fmt.Println("")

		wrapped := wordwrap.WrapString(createSubmissionResponse.OnInitSuccessMessage, 79)
		for _, line := range strings.Split(wrapped, "\n") {
			fmt.Printf("\033[1;92m%s\033[0m\n", line)
		}
	}

	if createSubmissionResponse.OnInitWarningMessage != "" {
		fmt.Println("")

		wrapped := wordwrap.WrapString(createSubmissionResponse.OnInitWarningMessage, 79)
		for _, line := range strings.Split(wrapped, "\n") {
			fmt.Printf("\033[31m%s\033[0m\n", line)
		}
	}

	logger.Debug().Msg("stream logs")

	fmt.Println("")
	err = streamLogs(createSubmissionResponse.LogstreamURL)
	if err != nil {
		return fmt.Errorf("stream logs: %w", err)
	}

	logger.Debug().Msg("fetch submission")

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

	logger.Debug().Interface("response", createSubmissionResponse).Msg("submission fetched")

	fmt.Println("")

	switch fetchSubmissionResponse.Status {
	case "failure":
		fmt.Println(createSubmissionResponse.OnFailureMessage)
	case "success":
		fmt.Println(createSubmissionResponse.OnSuccessMessage)
	}

	if fetchSubmissionResponse.IsError {
		return fmt.Errorf("%s", fetchSubmissionResponse.ErrorMessage)
	}

	return nil
}

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

		return "", wrapError(err, outputBytes, "run git command")
	}

	return strings.TrimSpace(string(outputBytes)), nil
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
