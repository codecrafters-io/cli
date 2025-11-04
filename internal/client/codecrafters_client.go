package client

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/codecrafters-io/cli/internal/globals"
	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/getsentry/sentry-go"
	"github.com/levigross/grequests"
)

type BuildpackInfo struct {
	Slug     string `json:"slug"`
	IsLatest bool   `json:"is_latest"`
}

type CreateSubmissionResponse struct {
	Id string `json:"id"`

	// Actions is the list of actions to execute for this submission
	Actions []ActionDefinition `json:"actions"`

	CommitSHA string `json:"commit_sha"`

	// IsError is true when the submission failed to be created, and ErrorMessage is the human-friendly error message
	IsError      bool   `json:"is_error"`
	ErrorMessage string `json:"error_message"`
}

type FetchBuildStatusResponse struct {
	Status string `json:"status"`

	ErrorMessage string `json:"error_message"`
	IsError      bool   `json:"is_error"`
}

type FetchSubmissionResponse struct {
	ErrorMessage string `json:"error_message"`
	IsError      bool   `json:"is_error"`
	Status       string `json:"status"`
}

type FetchAutofixRequestResponse struct {
	ErrorMessage string `json:"error_message"`
	IsError      bool   `json:"is_error"`
	Status       string `json:"status"`
}

type FetchBuildpacksResponse struct {
	Buildpacks   []BuildpackInfo `json:"buildpacks"`
	ErrorMessage string          `json:"error_message"`
	IsError      bool            `json:"is_error"`
}

type UpdateBuildpackResponse struct {
	Buildpack    BuildpackInfo `json:"buildpack"`
	ErrorMessage string        `json:"error_message"`
	IsError      bool          `json:"is_error"`
}

type FetchRepositoryBuildpackResponse struct {
	Buildpack    BuildpackInfo `json:"buildpack"`
	ErrorMessage string        `json:"error_message"`
	IsError      bool          `json:"is_error"`
}

type Stage struct {
	Slug                 string `json:"slug"`
	Name                 string `json:"name"`
	IsCurrent            bool   `json:"is_current"`
	InstructionsMarkdown string `json:"instructions_markdown"`
}

func (s Stage) GetDocsMarkdown() string {
	return fmt.Sprintf("# %s (#%s)\n\n%s", s.Name, s.Slug, s.InstructionsMarkdown)
}

type FetchStageListResponse struct {
	Stages       []Stage `json:"stages"`
	ErrorMessage string  `json:"error_message"`
	IsError      bool    `json:"is_error"`
}

type CodecraftersClient struct {
	ServerUrl string
}

func NewCodecraftersClient() CodecraftersClient {
	return CodecraftersClient{
		ServerUrl: globals.GetCodecraftersServerURL(),
	}
}

func (c CodecraftersClient) headers() map[string]string {
	return map[string]string{
		"X-Codecrafters-CLI-Version": utils.VersionString(),
	}
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
	response, err := grequests.Get(fmt.Sprintf("%s/services/cli/fetch_submission", c.ServerUrl), &grequests.RequestOptions{
		Params: map[string]string{
			"submission_id": submissionId,
		},
		Headers: c.headers(),
	})

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

func (c CodecraftersClient) FetchAutofixRequest(submissionId string) (FetchAutofixRequestResponse, error) {
	response, err := grequests.Get(fmt.Sprintf("%s/services/cli/fetch_autofix_request_status", c.ServerUrl), &grequests.RequestOptions{
		Params: map[string]string{
			"submission_id": submissionId,
		},
		Headers: c.headers(),
	})

	if err != nil {
		return FetchAutofixRequestResponse{}, fmt.Errorf("failed to fetch autofix request status from CodeCrafters: %s", err)
	}

	if !response.Ok {
		return FetchAutofixRequestResponse{}, fmt.Errorf("failed to fetch autofix request status from CodeCrafters. status code: %d", response.StatusCode)
	}

	fetchAutofixRequestResponse := FetchAutofixRequestResponse{}

	err = json.Unmarshal(response.Bytes(), &fetchAutofixRequestResponse)
	if err != nil {
		return FetchAutofixRequestResponse{}, fmt.Errorf("failed to parse fetch autofix request response: %s", err)
	}

	if fetchAutofixRequestResponse.IsError {
		return FetchAutofixRequestResponse{}, fmt.Errorf("%s", fetchAutofixRequestResponse.ErrorMessage)
	}

	return fetchAutofixRequestResponse, nil
}

func (c CodecraftersClient) FetchBuild(buildId string) (FetchBuildStatusResponse, error) {
	var fetchBuildResponse FetchBuildStatusResponse

	err := retry.Do(
		func() error {
			var err error
			fetchBuildResponse, err = c.doFetchBuild(buildId)
			if err != nil {
				return err
			}

			if fetchBuildResponse.Status != "failure" && fetchBuildResponse.Status != "success" {
				return fmt.Errorf("unexpected build status: %s", fetchBuildResponse.Status)
			}

			return nil
		},
		retry.Attempts(11),
		retry.DelayType(retry.BackOffDelay),
		retry.MaxDelay(2*time.Second),
		retry.Delay(100*time.Millisecond),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		if fetchBuildResponse.Status != "failure" && fetchBuildResponse.Status != "success" {
			sentry.CaptureException(err)
		}

		return FetchBuildStatusResponse{}, err
	}

	return fetchBuildResponse, nil
}

func (c CodecraftersClient) doFetchBuild(buildId string) (FetchBuildStatusResponse, error) {
	response, err := grequests.Get(fmt.Sprintf("%s/services/cli/fetch_test_runner_build", c.ServerUrl), &grequests.RequestOptions{
		Params: map[string]string{
			"test_runner_build_id": buildId,
		},
		Headers: c.headers(),
	})

	if err != nil {
		return FetchBuildStatusResponse{}, fmt.Errorf("failed to fetch build result from CodeCrafters: %s", err)
	}

	if !response.Ok {
		return FetchBuildStatusResponse{}, fmt.Errorf("failed to fetch build result from CodeCrafters. status code: %d", response.StatusCode)
	}

	fetchBuildResponse := FetchBuildStatusResponse{}

	err = json.Unmarshal(response.Bytes(), &fetchBuildResponse)
	if err != nil {
		return FetchBuildStatusResponse{}, fmt.Errorf("failed to fetch build result from CodeCrafters: %s", err)
	}

	return fetchBuildResponse, nil
}

func (c CodecraftersClient) FetchBuildpacks(repositoryId string) (FetchBuildpacksResponse, error) {
	response, err := grequests.Get(fmt.Sprintf("%s/services/cli/fetch_buildpacks", c.ServerUrl), &grequests.RequestOptions{
		Params: map[string]string{
			"repository_id": repositoryId,
		},
		Headers: c.headers(),
	})

	if err != nil {
		return FetchBuildpacksResponse{}, fmt.Errorf("failed to fetch buildpacks from CodeCrafters: %s", err)
	}

	if !response.Ok {
		return FetchBuildpacksResponse{}, fmt.Errorf("failed to fetch buildpacks from CodeCrafters. status code: %d", response.StatusCode)
	}

	fetchBuildpacksResponse := FetchBuildpacksResponse{}

	err = json.Unmarshal(response.Bytes(), &fetchBuildpacksResponse)
	if err != nil {
		return FetchBuildpacksResponse{}, fmt.Errorf("failed to fetch buildpacks from CodeCrafters: %s", err)
	}

	if fetchBuildpacksResponse.IsError {
		return fetchBuildpacksResponse, fmt.Errorf("%s", fetchBuildpacksResponse.ErrorMessage)
	}

	return fetchBuildpacksResponse, nil
}

func (c CodecraftersClient) UpdateBuildpack(repositoryId string) (UpdateBuildpackResponse, error) {
	response, err := grequests.Post(fmt.Sprintf("%s/services/cli/update_buildpack", c.ServerUrl), &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"repository_id": repositoryId,
		},
		Headers: c.headers(),
	})

	if err != nil {
		return UpdateBuildpackResponse{}, fmt.Errorf("failed to update buildpack: %s", err)
	}

	if !response.Ok {
		return UpdateBuildpackResponse{}, fmt.Errorf("failed to update buildpack. status code: %d", response.StatusCode)
	}

	updateBuildpackResponse := UpdateBuildpackResponse{}

	err = json.Unmarshal(response.Bytes(), &updateBuildpackResponse)
	if err != nil {
		return UpdateBuildpackResponse{}, fmt.Errorf("failed to parse update buildpack response: %s", err)
	}

	if updateBuildpackResponse.IsError {
		return updateBuildpackResponse, fmt.Errorf("%s", updateBuildpackResponse.ErrorMessage)
	}

	return updateBuildpackResponse, nil
}

func (c CodecraftersClient) FetchRepositoryBuildpack(repositoryId string) (FetchRepositoryBuildpackResponse, error) {
	response, err := grequests.Get(fmt.Sprintf("%s/services/cli/fetch_repository_buildpack", c.ServerUrl), &grequests.RequestOptions{
		Params: map[string]string{
			"repository_id": repositoryId,
		},
		Headers: c.headers(),
	})

	if err != nil {
		return FetchRepositoryBuildpackResponse{}, fmt.Errorf("failed to fetch repository buildpack from CodeCrafters: %s", err)
	}

	if !response.Ok {
		return FetchRepositoryBuildpackResponse{}, fmt.Errorf("failed to fetch repository buildpack from CodeCrafters. status code: %d", response.StatusCode)
	}

	fetchRepositoryBuildpackResponse := FetchRepositoryBuildpackResponse{}

	err = json.Unmarshal(response.Bytes(), &fetchRepositoryBuildpackResponse)
	if err != nil {
		return FetchRepositoryBuildpackResponse{}, fmt.Errorf("failed to parse fetch repository buildpack response: %s", err)
	}

	if fetchRepositoryBuildpackResponse.IsError {
		return fetchRepositoryBuildpackResponse, fmt.Errorf("%s", fetchRepositoryBuildpackResponse.ErrorMessage)
	}

	return fetchRepositoryBuildpackResponse, nil
}

func (c CodecraftersClient) FetchStageList(repositoryId string) (FetchStageListResponse, error) {
	response, err := grequests.Get(fmt.Sprintf("%s/services/cli/fetch_stage_list", c.ServerUrl), &grequests.RequestOptions{
		Params: map[string]string{
			"repository_id": repositoryId,
		},
		Headers: c.headers(),
	})

	if err != nil {
		return FetchStageListResponse{}, fmt.Errorf("failed to fetch stage list from CodeCrafters: %s", err)
	}

	if !response.Ok {
		return FetchStageListResponse{}, fmt.Errorf("failed to fetch stage list from CodeCrafters. status code: %d", response.StatusCode)
	}

	fetchStageListResponse := FetchStageListResponse{}

	err = json.Unmarshal(response.Bytes(), &fetchStageListResponse)
	if err != nil {
		return FetchStageListResponse{}, fmt.Errorf("failed to parse fetch stage list response: %s", err)
	}

	if fetchStageListResponse.IsError {
		return fetchStageListResponse, fmt.Errorf("%s", fetchStageListResponse.ErrorMessage)
	}

	return fetchStageListResponse, nil
}

type FetchDynamicActionsResponse struct {
	ActionDefinitions []ActionDefinition `json:"actions"`
}

func (c CodecraftersClient) FetchDynamicActions(eventName string, eventParams map[string]interface{}) (FetchDynamicActionsResponse, error) {
	queryParams := map[string]string{
		"event_name": eventName,
	}

	for key, value := range eventParams {
		queryParams[fmt.Sprintf("event_params[%s]", key)] = fmt.Sprintf("%v", value)
	}

	response, err := grequests.Get(fmt.Sprintf("%s/services/cli/fetch_dynamic_actions", c.ServerUrl), &grequests.RequestOptions{
		Params:  queryParams,
		Headers: c.headers(),
	})

	if err != nil {
		return FetchDynamicActionsResponse{}, fmt.Errorf("failed to fetch dynamic actions from CodeCrafters: %s", err)
	}

	if !response.Ok {
		return FetchDynamicActionsResponse{}, fmt.Errorf("failed to fetch dynamic actions from CodeCrafters. status code: %d, body: %s", response.StatusCode, response.String())
	}

	fetchDynamicActionsResponse := FetchDynamicActionsResponse{}

	err = json.Unmarshal(response.Bytes(), &fetchDynamicActionsResponse)
	if err != nil {
		return FetchDynamicActionsResponse{}, fmt.Errorf("failed to parse fetch dynamic actions response: %s", err)
	}

	return fetchDynamicActionsResponse, nil
}
