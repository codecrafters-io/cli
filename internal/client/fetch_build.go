package client

import (
	"encoding/json"
	"fmt"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/getsentry/sentry-go"
	"github.com/levigross/grequests"
)

type FetchBuildStatusResponse struct {
	Status string `json:"status"`

	ErrorMessage string `json:"error_message"`
	IsError      bool   `json:"is_error"`
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
