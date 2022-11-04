package commands

import (
	"fmt"
	"github.com/codecrafters-io/cli/internal/utils"
	logstream_consumer "github.com/codecrafters-io/logstream/consumer"
	"github.com/fatih/color"
	wordwrap "github.com/mitchellh/go-wordwrap"
	cp "github.com/otiai10/copy"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func TestCommand() int {
	repoDir, err := getRepositoryDir()
	if err != nil {
		return 1
	}

	codecraftersRemote, err := utils.IdentifyGitRemote(repoDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	tmpDir, err := copyRepositoryDirToTempDir(repoDir)
	if err != nil {
		return 1
	}

	tempBranchName := "cli-test-" + strconv.FormatInt(time.Now().UnixMilli(), 10)

	err = checkoutNewBranch(tempBranchName, tmpDir)
	if err != nil {
		return 1
	}

	tempCommitSha, err := commitChanges(tmpDir, fmt.Sprintf("CLI tests (%s)", tempBranchName))
	if err != nil {
		return 1
	}

	// Place this before the push so that it "feels" fast
	fmt.Println("Running tests on your codebase. Streaming logs...")

	err = pushBranchToRemote(tmpDir)
	if err != nil {
		return 1
	}

	codecraftersClient := utils.NewCodecraftersClient(codecraftersRemote.CodecraftersServerURL())

	createSubmissionResponse, err := codecraftersClient.CreateSubmission(codecraftersRemote.CodecraftersRepositoryId(), tempCommitSha)
	if err != nil {
		return 1
	}

	if createSubmissionResponse.IsError {
		fmt.Fprintf(os.Stderr, "failed to create submission: %s", createSubmissionResponse.ErrorMessage)
		return 1
	}

	if createSubmissionResponse.OnInitSuccessMessage != "" {
		fmt.Println("")

		wrapped := wordwrap.WrapString(createSubmissionResponse.OnInitSuccessMessage, 79)
		for _, line := range strings.Split(wrapped, "\n") {
			fmt.Println(fmt.Sprintf("\033[1;92m%s\033[0m", line))
		}
	}

	if createSubmissionResponse.OnInitWarningMessage != "" {
		fmt.Println("")

		wrapped := wordwrap.WrapString(createSubmissionResponse.OnInitWarningMessage, 79)
		for _, line := range strings.Split(wrapped, "\n") {
			fmt.Println(fmt.Sprintf("\033[31m%s\033[0m", line))
		}
	}

	fmt.Println("")
	err = streamLogs(createSubmissionResponse.LogstreamUrl)
	if err != nil {
		return 1
	}

	fetchSubmissionResponse, err := codecraftersClient.FetchSubmission(createSubmissionResponse.Id)
	if err != nil {
		// TODO: Notify sentry
		red := color.New(color.FgRed).SprintFunc()
		fmt.Fprintln(os.Stderr, red(err.Error()))
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, red("We couldn't fetch the results of your submission. Please try again?"))
		fmt.Fprintln(os.Stderr, red("Let us know at hello@codecrafters.io if this error persists."))
		return 1
	}

	if fetchSubmissionResponse.Status == "failure" {
		fmt.Println("")
		fmt.Println(createSubmissionResponse.OnFailureMessage)
		return 1
	}

	if fetchSubmissionResponse.Status == "success" {
		fmt.Println("")
		fmt.Println(createSubmissionResponse.OnSuccessMessage)
		return 0
	}

	return 0
}

func getRepositoryDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch current working directory: %s", err)
		return "", err
	}

	outputBytes, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			if regexp.MustCompile("not a git repository").Match(outputBytes) {
				fmt.Fprintf(os.Stderr, "The current directory is not within a Git repository.\n")
				fmt.Fprintf(os.Stderr, "Please run this command from within your CodeCrafters Git repository.\n")
			} else {
				fmt.Fprintln(os.Stderr, string(outputBytes))
			}

			return "", err
		} else {
			panic(err)
		}
	}

	return strings.TrimSpace(string(outputBytes)), nil
}

func copyRepositoryDirToTempDir(repoDir string) (string, error) {
	tmpDir, err := ioutil.TempDir("", "codecrafters")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temporary directory: %s", err)
		return "", err
	}

	err = cp.Copy(repoDir, tmpDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to copy to temporary directory: %s", err)
		return "", err
	}

	return tmpDir, nil
}

func checkoutNewBranch(tempBranchName string, tmpDir string) error {
	err := exec.Command("git", "-C", tmpDir, "checkout", "-b", tempBranchName).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp branch: %s", err)
		return err
	}

	return nil
}

func commitChanges(tmpDir string, commitMessage string) (string, error) {
	outputBytes, err := exec.Command("git", "-C", tmpDir, "commit", "--allow-empty", "-a", "-m", commitMessage).CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(os.Stderr, "failed to create temp commit: %s", outputBytes)
			return "", err
		} else {
			fmt.Fprintf(os.Stderr, "failed to create temp commit: %s", err)
			return "", err
		}
	}

	outputBytes, err = exec.Command("git", "-C", tmpDir, "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(os.Stderr, "failed to fetch temp commit sha: %s", outputBytes)
			return "", err
		} else {
			fmt.Fprintf(os.Stderr, "failed to fetch temp commit sha: %s", err)
			return "", err
		}
	}

	return strings.TrimSpace(string(outputBytes)), nil
}

func pushBranchToRemote(tmpDir string) error {
	// TODO: Find CodeCrafters remote and use that
	outputBytes, err := exec.Command("git", "-C", tmpDir, "push", "origin", "HEAD").CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(os.Stderr, "failed to push temp branch: %s", outputBytes)
			return err
		} else {
			fmt.Fprintf(os.Stderr, "failed to push temp branch: %s", err)
			return err
		}
	}

	return nil
}

func streamLogs(logstreamUrl string) error {
	consumer, err := logstream_consumer.NewConsumer(logstreamUrl, func(message string) {})
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		return err
	}

	_, err = io.Copy(os.Stdout, consumer)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		return err
	}

	return nil
}
