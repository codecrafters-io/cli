package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	retry "github.com/avast/retry-go"
	"github.com/levigross/grequests"
	"os"
	"time"
)

type CreateSubmissionResponse struct {
	Id               string `json:"id"`
	ErrorMessage     string `json:"error_message"`
	IsError          bool   `json:"is_error"`
	LogstreamUrl     string `json:"logstream_url"`
	OnFailureMessage string `json:"on_failure_message"`
	OnSuccessMessage string `json:"on_success_message"`
}

type FetchSubmissionResponse struct {
	ErrorMessage string `json:"error_message"`
	IsError      bool   `json:"is_error"`
	Status       string `json:"status"`
}

type CodecraftersClient struct {
	ServerUrl  string
	CLIVersion string
}

func NewCodecraftersClient(serverUrl string, cliVersion string) CodecraftersClient {
	return CodecraftersClient{ServerUrl: serverUrl, CLIVersion: cliVersion}
}

func (c CodecraftersClient) headers() map[string]string {
	return map[string]string{
		"X-Codecrafters-CLI-Version": c.CLIVersion,
	}
}

func (c CodecraftersClient) CreateSubmission(repositoryId string, commitSha string) (CreateSubmissionResponse, error) {
	// TODO: Include version in headers?
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
	// TODO: Include version in headers?
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
