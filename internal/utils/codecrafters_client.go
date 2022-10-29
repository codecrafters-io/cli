package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/levigross/grequests"
	"os"
)

type CreateSubmissionResponse struct {
	ErrorMessage         string `json:"error_message"`
	IsError              bool   `json:"is_error"`
	LogstreamUrl         string `json:"logstream_url"`
	OnTestsFailedMessage string `json:"on_tests_failed_message"`
	OnTestsPassedMessage string `json:"on_tests_passed_message"`
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

func (c CodecraftersClient) CreateSubmission(repositoryId string, commitSha string) (CreateSubmissionResponse, error) {
	// TODO: Include version in headers?
	response, err := grequests.Post(c.ServerUrl+"/submissions", &grequests.RequestOptions{JSON: map[string]interface{}{
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
