package client

import (
	"encoding/json"
	"fmt"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/levigross/grequests"
)

type FetchSubmissionResponse struct {
	ErrorMessage string `json:"error_message"`
	IsError      bool   `json:"is_error"`
	Status       string `json:"status"`
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
	utils.Logger.Debug().Msgf("GET /services/cli/fetch_submission?submission_id=%s", submissionId)

	response, err := grequests.Get(fmt.Sprintf("%s/services/cli/fetch_submission", c.ServerUrl), &grequests.RequestOptions{
		Params: map[string]string{
			"submission_id": submissionId,
		},
		Headers: c.headers(),
	})

	if err != nil {
		return FetchSubmissionResponse{}, fmt.Errorf("failed to fetch submission result from CodeCrafters: %s", err)
	}

	utils.Logger.Debug().Msgf("response: %s", response.String())

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
