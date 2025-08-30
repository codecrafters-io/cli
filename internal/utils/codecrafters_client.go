package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/getsentry/sentry-go"
	"github.com/levigross/grequests"
	"github.com/mitchellh/go-wordwrap"
)

type Message struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}

func (m Message) Print() {
	wrapped := wordwrap.WrapString(m.Text, 79)

	lineFormat := "%s\n"

	switch m.Color {
	case "red":
		lineFormat = "\033[31m%s\033[0m\n"
	case "green":
		lineFormat = "\033[32m%s\033[0m\n"
	case "yellow":
		lineFormat = "\033[33m%s\033[0m\n"
	case "blue":
		lineFormat = "\033[34m%s\033[0m\n"
	}

	for _, line := range strings.Split(wrapped, "\n") {
		fmt.Printf(lineFormat, line)
	}
}

type BuildpackInfo struct {
	Slug     string `json:"slug"`
	IsLatest bool   `json:"is_latest"`
}

type CreateSubmissionResponse struct {
	Id string `json:"id"`

	// BuildLogstreamURL is returned when the submission is waiting on a build
	BuildID           string `json:"build_id"`
	BuildLogstreamURL string `json:"build_logstream_url"`

	CommitSHA string `json:"commit_sha"`

	// LogstreamURL contains test logs.
	LogstreamURL string `json:"logstream_url"`

	// Messages to be displayed to the user at various stages of the submission lifecycle
	OnInitMessages    []Message `json:"on_init_messages"`
	OnSuccessMessages []Message `json:"on_success_messages"`
	OnFailureMessages []Message `json:"on_failure_messages"`

	// IsError is true when the submission failed to be created, and ErrorMessage is the human-friendly error message
	IsError      bool   `json:"is_error"`
	ErrorMessage string `json:"error_message"`
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
