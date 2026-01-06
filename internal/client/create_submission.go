package client

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/levigross/grequests"
)

type CreateSubmissionResponse struct {
	Id string `json:"id"`

	// Actions is the list of actions to execute for this submission
	Actions []ActionDefinition `json:"actions"`

	CommitSHA string `json:"commit_sha"`

	// IsError is true when the submission failed to be created, and ErrorMessage is the human-friendly error message
	IsError      bool   `json:"is_error"`
	ErrorMessage string `json:"error_message"`
}

func (c CodecraftersClient) CreateSubmission(repositoryId string, commitSha string, command string, stageSelectionStrategy string) (CreateSubmissionResponse, error) {
	response, err := grequests.Post(c.ServerUrl+"/services/cli/create_submission", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"repository_id":            repositoryId,
			"commit_sha":               commitSha,
			"command":                  command,
			"stage_selection_strategy": stageSelectionStrategy,
		},
		Headers: c.headers(),
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to submit code to CodeCrafters: %s\n", err)
		return CreateSubmissionResponse{}, err
	}

	if !response.Ok && response.StatusCode != 403 {
		return CreateSubmissionResponse{}, fmt.Errorf("failed to submit code to CodeCrafters. status code: %d, body: %s", response.StatusCode, response.String())
	}

	createSubmissionResponse := CreateSubmissionResponse{}

	err = json.Unmarshal(response.Bytes(), &createSubmissionResponse)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to submit code to CodeCrafters: %s\n", err)
		return CreateSubmissionResponse{}, err
	}

	if createSubmissionResponse.IsError {
		return createSubmissionResponse, fmt.Errorf("%s", createSubmissionResponse.ErrorMessage)
	}

	return createSubmissionResponse, nil
}
