package client

import (
	"encoding/json"
	"fmt"

	"github.com/levigross/grequests"
)

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
