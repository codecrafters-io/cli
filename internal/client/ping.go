package client

import (
	"encoding/json"
	"fmt"

	"github.com/levigross/grequests"
)

type PingResponse struct {
	Actions []ActionDefinition `json:"actions"`
}

func (c CodecraftersClient) Ping(repositoryId string) (PingResponse, error) {
	response, err := grequests.Post(c.ServerUrl+"/services/cli/ping", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"repository_id": repositoryId,
		},
		Headers: c.headers(),
	})

	if err != nil {
		return PingResponse{}, fmt.Errorf("failed to ping CodeCrafters: %s", err)
	}

	if !response.Ok {
		return PingResponse{}, fmt.Errorf("failed to ping CodeCrafters. status code: %d, body: %s", response.StatusCode, response.String())
	}

	pingResponse := PingResponse{}

	err = json.Unmarshal(response.Bytes(), &pingResponse)
	if err != nil {
		return PingResponse{}, fmt.Errorf("failed to parse ping response: %s", err)
	}

	return pingResponse, nil
}
