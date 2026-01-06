package client

import (
	"encoding/json"
	"fmt"

	"github.com/levigross/grequests"
)

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
