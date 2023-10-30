package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/levigross/grequests"
)

type CreateSubmissionResponse struct {
	Id string `json:"id"`

	// BuildLogstreamURL is returned when the submission is waiting on a build
	BuildID           string `json:"build_id"`
	BuildLogstreamURL string `json:"build_logstream_url"`

	CommitSHA string `json:"commit_sha"`

	// LogstreamURL contains test logs.
	LogstreamURL string `json:"logstream_url"`

	// Messages to be displayed to the user
	OnSuccessMessage     string `json:"on_success_message"`
	OnFailureMessage     string `json:"on_failure_message"`
	OnInitSuccessMessage string `json:"on_init_success_message"`
	OnInitWarningMessage string `json:"on_init_warning_message"`

	// IsError is true when the submission failed to be created, and ErrorMessage is the human-friendly error message
	IsError      bool   `json:"is_error"`
	ErrorMessage string `json:"error_message"`
}

type FetchSubmissionResponse struct {
	ErrorMessage string `json:"error_message"`
	IsError      bool   `json:"is_error"`
	Status       string `json:"status"`
}

type CodecraftersClient struct {
	ServerUrl string
}

func NewCodecraftersClient(serverUrl string) CodecraftersClient {
	return CodecraftersClient{ServerUrl: serverUrl}
}

func (c CodecraftersClient) headers() map[string]string {
	return map[string]string{
		"X-Codecrafters-CLI-Version": VersionString(),
	}
}

func (c CodecraftersClient) CreateSubmission(repositoryId string, commitSha string) (CreateSubmissionResponse, error) {
	response, err := grequests.Post(c.ServerUrl+"/submissions", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"repository_id":       repositoryId,
			"commit_sha":          commitSha,
			"should_auto_advance": false,
		},
		Headers: c.headers(),
	})

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

	if createSubmissionResponse.IsError {
		return createSubmissionResponse, fmt.Errorf("%s", createSubmissionResponse.ErrorMessage)
	}

	return createSubmissionResponse, nil
}

func (c CodecraftersClient) FetchSubmission(submissionId string) (FetchSubmissionResponse, error) {
	var fetchSubmissionResponse FetchSubmissionResponse

	err := retry.Do(
		func() error {
			var err error
			fetchSubmissionResponse, err = c.doFetchSubmission(submissionId)
			if err != nil {
				return err
			}

			if fetchSubmissionResponse.Status != "failure" && fetchSubmissionResponse.Status != "success" {
				return fmt.Errorf("unexpected submission status: %s", fetchSubmissionResponse.Status)
			}

			return nil
		},
		retry.Attempts(5),
		retry.DelayType(retry.BackOffDelay),
		retry.MaxDelay(2*time.Second),
		retry.Delay(500*time.Millisecond),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		return FetchSubmissionResponse{}, err
	}

	return fetchSubmissionResponse, nil
}

func (c CodecraftersClient) doFetchSubmission(submissionId string) (FetchSubmissionResponse, error) {
	response, err := grequests.Get(fmt.Sprintf("%s/submissions/%s", c.ServerUrl, submissionId), &grequests.RequestOptions{Headers: c.headers()})

	if err != nil {
		return FetchSubmissionResponse{}, fmt.Errorf("failed to fetch submission result from CodeCrafters: %s", err)
	}

	if !response.Ok {
		return FetchSubmissionResponse{}, fmt.Errorf("failed to fetch submission result from CodeCrafters. status code: %d", response.StatusCode)
	}

	fetchSubmissionResponse := FetchSubmissionResponse{}

	err = json.Unmarshal(response.Bytes(), &fetchSubmissionResponse)
	if err != nil {
		return FetchSubmissionResponse{}, fmt.Errorf("failed to fetch submission result from CodeCrafters: %s", err)
	}

	return fetchSubmissionResponse, nil
}
