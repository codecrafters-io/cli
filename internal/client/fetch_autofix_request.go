package client

import (
	"encoding/json"
	"fmt"

	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/levigross/grequests"
)

type FetchAutofixRequestResponse struct {
	ErrorMessage string `json:"error_message"`
	IsError      bool   `json:"is_error"`
	Status       string `json:"status"`
}

func (c CodecraftersClient) FetchAutofixRequest(submissionId string) (FetchAutofixRequestResponse, error) {
	utils.Logger.Debug().Msgf("GET /services/cli/fetch_autofix_request?submission_id=%s", submissionId)

	response, err := grequests.Get(fmt.Sprintf("%s/services/cli/fetch_autofix_request", c.ServerUrl), &grequests.RequestOptions{
		Params: map[string]string{
			"submission_id": submissionId,
		},
		Headers: c.headers(),
	})

	if err != nil {
		return FetchAutofixRequestResponse{}, fmt.Errorf("failed to fetch autofix request status from CodeCrafters: %s", err)
	}

	utils.Logger.Debug().Msgf("response: %s", response.String())

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
