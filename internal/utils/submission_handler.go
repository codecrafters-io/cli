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

	// Convert action definitions to concrete actions
	actionsToExecute := []actions.Action{}
	for _, actionDef := range createSubmissionResponse.Actions {
		action, err := actions.ActionFromDefinition(actions.ActionDefinition{
			Type: actionDef.Type,
			Args: actionDef.Args,
		})
		if err != nil {
			return fmt.Errorf("failed to create action from definition: %w", err)
		}
		actionsToExecute = append(actionsToExecute, action)
	}

	// Create adapter for CodecraftersClient
	clientAdapter := &codecraftersClientAdapter{client: codecraftersClient}

	// Inject CodecraftersClient into actions that need it
	for _, action := range actionsToExecute {
		actions.InjectCodecraftersClient(action, clientAdapter)
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

// codecraftersClientAdapter adapts utils.CodecraftersClient to actions.CodecraftersClient
type codecraftersClientAdapter struct {
	client CodecraftersClient
}

func (a *codecraftersClientAdapter) FetchBuild(buildId string) (actions.FetchBuildStatusResponse, error) {
	resp, err := a.client.FetchBuild(buildId)
	if err != nil {
		return actions.FetchBuildStatusResponse{}, err
	}
	return actions.FetchBuildStatusResponse{Status: resp.Status}, nil
}

func (a *codecraftersClientAdapter) FetchSubmission(submissionId string) (actions.FetchSubmissionResponse, error) {
	resp, err := a.client.FetchSubmission(submissionId)
	if err != nil {
		return actions.FetchSubmissionResponse{}, err
	}
	return actions.FetchSubmissionResponse{Status: resp.Status}, nil
}
