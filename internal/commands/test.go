package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	logstream_consumer "github.com/codecrafters-io/logstream/consumer"
	"github.com/levigross/grequests"
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

type CreateSubmissionResponse struct {
	ErrorMessage string `json:"error_message"`
	IsError      bool   `json:"is_error"`
	LogstreamUrl string `json:"logstream_url"`
}

func TestCommand() int {
	repoDir, err := getRepositoryDir()
	if err != nil {
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

	err = pushBranchToRemote(tmpDir)
	if err != nil {
		return 1
	}

	// TODO: Pick repository name from remotes
	repositoryName := strings.Split(repoDir, "/")[len(strings.Split(repoDir, "/"))-1]
	fmt.Println(repositoryName)

	createSubmissionResponse, err := createSubmission("https://app.staging.codecrafters.io", repositoryName, tempCommitSha)
	if err != nil {
		return 1
	}

	if createSubmissionResponse.IsError {
		fmt.Fprintf(os.Stderr, "failed to create submission: %s", createSubmissionResponse.ErrorMessage)
		return 1
	} else {
		fmt.Println("submitted!")
		fmt.Println(createSubmissionResponse.LogstreamUrl)
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

func createSubmission(serverUrl string, repositoryId string, commitSha string) (CreateSubmissionResponse, error) {
	// TODO: Include version in headers?
	response, err := grequests.Post(serverUrl+"/submissions", &grequests.RequestOptions{JSON: map[string]interface{}{
		"repository_id":       repositoryId,
		"commit_sha":          commitSha,
		"should_auto_advance": false,
	}})

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to submit code to CodeCrafters: %s", err)
		return CreateSubmissionResponse{}, err
	}

	if !response.Ok && response.StatusCode != 403 {
		fmt.Fprintf(os.Stderr, "failed to submit code to CodeCrafters. status code: %d. body: %s", response.StatusCode, response.String())
		return CreateSubmissionResponse{}, errors.New("dummy")
	}

	createSubmissionResponse := CreateSubmissionResponse{}

	err = json.Unmarshal(response.Bytes(), &createSubmissionResponse)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to submit code to CodeCrafters: %s", err)
		return CreateSubmissionResponse{}, err
	}

	return createSubmissionResponse, nil
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
