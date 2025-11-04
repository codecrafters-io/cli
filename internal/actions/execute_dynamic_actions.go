package actions

import (
	"encoding/json"

	"github.com/codecrafters-io/cli/internal/client"
)

type ExecuteDynamicActionsAction struct {
	EventName   string                 `json:"event_name"`
	EventParams map[string]interface{} `json:"event_params"`
}

func NewExecuteDynamicActionsAction(argsJson json.RawMessage) (ExecuteDynamicActionsAction, error) {
	var executeDynamicActionsAction ExecuteDynamicActionsAction
	if err := json.Unmarshal(argsJson, &executeDynamicActionsAction); err != nil {
		return ExecuteDynamicActionsAction{}, err
	}

	return executeDynamicActionsAction, nil
}

func (a ExecuteDynamicActionsAction) Execute() error {
	codecraftersClient := client.NewCodecraftersClient()
	response, err := codecraftersClient.FetchDynamicActions(a.EventName, a.EventParams)
	if err != nil {
		return err
	}

	actions := []Action{}
	for _, actionDefinition := range response.ActionDefinitions {
		action, err := ActionFromDefinition(actionDefinition)
		if err != nil {
			return err
		}

		actions = append(actions, action)
	}

	for _, action := range actions {
		if err := action.Execute(); err != nil {
			return err
		}
	}

	return nil
}
