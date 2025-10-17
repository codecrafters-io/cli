package actions

import (
	"encoding/json"
	"fmt"
)

// ActionDefinition represents the JSON structure of an action from core
type ActionDefinition struct {
	Type string          `json:"type"`
	Args json.RawMessage `json:"args"`
}

// Action is the interface that all actions must implement
type Action interface {
	Execute() error
}

// CodecraftersClient is an interface for the client (to avoid import cycles)
type CodecraftersClient interface {
	FetchBuild(buildId string) (FetchBuildStatusResponse, error)
	FetchSubmission(submissionId string) (FetchSubmissionResponse, error)
}

// FetchBuildStatusResponse matches the response from utils
type FetchBuildStatusResponse struct {
	Status string `json:"status"`
}

// FetchSubmissionResponse matches the response from utils
type FetchSubmissionResponse struct {
	Status string `json:"status"`
}

// ActionWithClient is an interface for actions that need a CodecraftersClient
type ActionWithClient interface {
	Action
	SetCodecraftersClient(CodecraftersClient)
}

// ActionFromDefinition converts an ActionDefinition into a concrete Action
func ActionFromDefinition(actionDefinition ActionDefinition) (Action, error) {
	switch actionDefinition.Type {
	case "await_terminal_build_status":
		return NewAwaitTerminalBuildStatusAction(actionDefinition.Args)
	case "await_terminal_submission_status":
		return NewAwaitTerminalSubmissionStatusAction(actionDefinition.Args)
	case "print_message":
		return NewPrintMessageAction(actionDefinition.Args)
	case "sleep":
		return NewSleepAction(actionDefinition.Args)
	case "stream_logs":
		return NewStreamLogsAction(actionDefinition.Args)
	case "terminate":
		return NewTerminateAction(actionDefinition.Args)
	default:
		return nil, fmt.Errorf("unexpected action type: %s", actionDefinition.Type)
	}
}

// InjectCodecraftersClient recursively injects the CodecraftersClient into all actions that need it
func InjectCodecraftersClient(action Action, client CodecraftersClient) {
	if actionWithClient, ok := action.(ActionWithClient); ok {
		actionWithClient.SetCodecraftersClient(client)
	}

	// Handle nested actions
	switch a := action.(type) {
	case *AwaitTerminalBuildStatusAction:
		for _, nestedAction := range a.OnSuccessActions {
			InjectCodecraftersClient(nestedAction, client)
		}
		for _, nestedAction := range a.OnFailureActions {
			InjectCodecraftersClient(nestedAction, client)
		}
	case *AwaitTerminalSubmissionStatusAction:
		for _, nestedAction := range a.OnSuccessActions {
			InjectCodecraftersClient(nestedAction, client)
		}
		for _, nestedAction := range a.OnFailureActions {
			InjectCodecraftersClient(nestedAction, client)
		}
	}
}
