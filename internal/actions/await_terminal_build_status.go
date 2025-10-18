package actions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

type AwaitTerminalBuildStatusAction struct {
	BuildID          string
	OnSuccessActions []Action
	OnFailureActions []Action
}

type AwaitTerminalBuildStatusActionArgs struct {
	BuildID          string             `json:"build_id"`
	OnSuccessActions []ActionDefinition `json:"on_success_actions"`
	OnFailureActions []ActionDefinition `json:"on_failure_actions"`
}

func NewAwaitTerminalBuildStatusAction(argsJson json.RawMessage) (*AwaitTerminalBuildStatusAction, error) {
	var awaitTerminalBuildStatusActionArgs AwaitTerminalBuildStatusActionArgs
	if err := json.Unmarshal(argsJson, &awaitTerminalBuildStatusActionArgs); err != nil {
		return nil, err
	}

	onSuccessActions := []Action{}
	for _, actionDefinition := range awaitTerminalBuildStatusActionArgs.OnSuccessActions {
		action, err := ActionFromDefinition(actionDefinition)
		if err != nil {
			return nil, err
		}

		onSuccessActions = append(onSuccessActions, action)
	}

	onFailureActions := []Action{}
	for _, actionDefinition := range awaitTerminalBuildStatusActionArgs.OnFailureActions {
		action, err := ActionFromDefinition(actionDefinition)
		if err != nil {
			return nil, err
		}

		onFailureActions = append(onFailureActions, action)
	}

	return &AwaitTerminalBuildStatusAction{
		BuildID:          awaitTerminalBuildStatusActionArgs.BuildID,
		OnSuccessActions: onSuccessActions,
		OnFailureActions: onFailureActions,
	}, nil
}

func (a *AwaitTerminalBuildStatusAction) Execute() error {
	attempts := 0
	buildStatus := "not_started"

	// We start waiting for 100 ms, gradually increasing to 2 seconds. Total wait time can be upto 21 seconds ((20 + 21 / 2) * 100ms)
	for buildStatus != "success" && buildStatus != "failure" && buildStatus != "error" && attempts < 20 {
		var err error

		buildStatus, err = FetchBuildStatus(a.BuildID)
		if err != nil {
			// We can still proceed here anyway
			sentry.CaptureException(err)
		}

		attempts += 1
		time.Sleep(time.Duration(100*attempts) * time.Millisecond)
	}

	switch buildStatus {
	case "success":
		for _, action := range a.OnSuccessActions {
			if err := action.Execute(); err != nil {
				return err
			}
		}
	case "failure":
		for _, action := range a.OnFailureActions {
			if err := action.Execute(); err != nil {
				return err
			}
		}
	default:
		err := fmt.Errorf("unexpected build status: %s", buildStatus)
		sentry.CaptureException(err)

		printErr := PrintMessageAction{Color: "red", Text: "We couldn't fetch the results of your build. Please try again?"}.Execute()
		if printErr != nil {
			return printErr
		}
		printErr = PrintMessageAction{Color: "red", Text: "Let us know at hello@codecrafters.io if this error persists."}.Execute()
		if printErr != nil {
			return printErr
		}

		// If the build failed, we don't need to stream test logs
		return TerminateAction{ExitCode: 1}.Execute()
	}

	return nil
}
