package client

import (
	"encoding/json"
	"fmt"

	"github.com/levigross/grequests"
)

type FetchBuildpacksResponse struct {
	Buildpacks   []BuildpackInfo `json:"buildpacks"`
	ErrorMessage string          `json:"error_message"`
	IsError      bool            `json:"is_error"`
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
