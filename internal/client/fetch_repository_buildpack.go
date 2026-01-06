package client

import (
	"encoding/json"
	"fmt"

	"github.com/levigross/grequests"
)

type FetchRepositoryBuildpackResponse struct {
	Buildpack    BuildpackInfo `json:"buildpack"`
	ErrorMessage string        `json:"error_message"`
	IsError      bool          `json:"is_error"`
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
