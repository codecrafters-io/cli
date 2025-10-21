package actions

import (
	"fmt"

	"github.com/codecrafters-io/cli/internal/client"
)

type Action interface {
	Execute() error
}

// Package-level backend functions that will be set by the caller (due to import cycles)
var (
	FetchBuildStatus      func(buildId string) (string, error)
	FetchSubmissionStatus func(submissionId string) (string, error)
)

func ActionFromDefinition(actionDefinition client.ActionDefinition) (Action, error) {
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
