package actions

import (
	"fmt"

	"github.com/codecrafters-io/cli/internal/client"
)

type Action interface {
	Execute() error
}

func ActionFromDefinition(actionDefinition client.ActionDefinition) (Action, error) {
	switch actionDefinition.Type {
	case "await_terminal_build_status":
		return NewAwaitTerminalBuildStatusAction(actionDefinition.Args)
	case "await_terminal_submission_status":
		return NewAwaitTerminalSubmissionStatusAction(actionDefinition.Args)
	case "await_terminal_autofix_request_status":
		return NewAwaitTerminalAutofixRequestStatusAction(actionDefinition.Args)
	case "execute_dynamic_actions":
		return NewExecuteDynamicActionsAction(actionDefinition.Args)
	case "print_file_diff":
		return NewPrintFileDiffAction(actionDefinition.Args)
	case "print_message":
		return NewPrintMessageAction(actionDefinition.Args)
	case "print_terminal_commands_box":
		return NewPrintTerminalCommandsBoxAction(actionDefinition.Args)
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
