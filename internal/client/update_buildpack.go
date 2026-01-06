package client

import (
	"encoding/json"
	"fmt"

	"github.com/levigross/grequests"
)

type UpdateBuildpackResponse struct {
	Buildpack    BuildpackInfo `json:"buildpack"`
	ErrorMessage string        `json:"error_message"`
	IsError      bool          `json:"is_error"`
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
