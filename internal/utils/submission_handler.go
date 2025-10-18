package utils

import (
	"context"
	"fmt"

	"github.com/codecrafters-io/cli/internal/actions"
	"github.com/rs/zerolog"
)

func HandleSubmission(createSubmissionResponse CreateSubmissionResponse, ctx context.Context, codecraftersClient CodecraftersClient) (err error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Msgf("Handling submission with %d actions", len(createSubmissionResponse.Actions))

	actions.FetchBuildStatus = func(buildId string) (string, error) {
		resp, err := codecraftersClient.FetchBuild(buildId)
		if err != nil {
			return "", err
		}
		return resp.Status, nil
	}

	actions.FetchSubmissionStatus = func(submissionId string) (string, error) {
		resp, err := codecraftersClient.FetchSubmission(submissionId)
		if err != nil {
			return "", err
		}
		return resp.Status, nil
	}

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
		logger.Debug().Msgf("Executing action %d of %d", i+1, len(actionsToExecute))
		if err := action.Execute(); err != nil {
			return fmt.Errorf("failed to execute action: %w", err)
		}
	}

	return nil
}
