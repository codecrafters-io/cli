package commands

import (
	"fmt"
	"reflect"

	"github.com/codecrafters-io/cli/internal/actions"
	"github.com/codecrafters-io/cli/internal/client"
	"github.com/codecrafters-io/cli/internal/utils"
)

func handleSubmission(createSubmissionResponse client.CreateSubmissionResponse, codecraftersClient client.CodecraftersClient) (err error) {
	utils.Logger.Debug().Msgf("Handling submission with %d actions", len(createSubmissionResponse.Actions))

	// Convert action definitions to concrete actions
	actionsToExecute := []actions.Action{}
	for _, actionDef := range createSubmissionResponse.Actions {
		action, err := actions.ActionFromDefinition(actionDef)
		if err != nil {
			return fmt.Errorf("failed to create action from definition: %w", err)
		}
		actionsToExecute = append(actionsToExecute, action)
	}

	// Execute all actions in sequence
	for i, action := range actionsToExecute {
		utils.Logger.Debug().Msgf("Executing %s (%d/%d)", reflect.TypeOf(action).String(), i+1, len(actionsToExecute))

		if err := action.Execute(); err != nil {
			return fmt.Errorf("failed to execute action: %w", err)
		}

		utils.Logger.Debug().Msgf("%s completed", reflect.TypeOf(action).String())
	}

	return nil
}
