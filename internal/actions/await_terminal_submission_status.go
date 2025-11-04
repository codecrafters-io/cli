package actions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/codecrafters-io/cli/internal/client"
	"github.com/getsentry/sentry-go"
)

type AwaitTerminalSubmissionStatusAction struct {
	SubmissionID     string
	OnSuccessActions []Action
	OnFailureActions []Action
}

type AwaitTerminalSubmissionStatusActionArgs struct {
	SubmissionID     string                    `json:"submission_id"`
	OnSuccessActions []client.ActionDefinition `json:"on_success_actions"`
	OnFailureActions []client.ActionDefinition `json:"on_failure_actions"`
}

func NewAwaitTerminalSubmissionStatusAction(argsJson json.RawMessage) (*AwaitTerminalSubmissionStatusAction, error) {
	var awaitTerminalSubmissionStatusActionArgs AwaitTerminalSubmissionStatusActionArgs
	if err := json.Unmarshal(argsJson, &awaitTerminalSubmissionStatusActionArgs); err != nil {
		return nil, err
	}

	onSuccessActions := []Action{}
	for _, actionDefinition := range awaitTerminalSubmissionStatusActionArgs.OnSuccessActions {
		action, err := ActionFromDefinition(actionDefinition)
		if err != nil {
			return nil, err
		}

		onSuccessActions = append(onSuccessActions, action)
	}

	onFailureActions := []Action{}
	for _, actionDefinition := range awaitTerminalSubmissionStatusActionArgs.OnFailureActions {
		action, err := ActionFromDefinition(actionDefinition)
		if err != nil {
			return nil, err
		}

		onFailureActions = append(onFailureActions, action)
	}

	return &AwaitTerminalSubmissionStatusAction{
		SubmissionID:     awaitTerminalSubmissionStatusActionArgs.SubmissionID,
		OnSuccessActions: onSuccessActions,
		OnFailureActions: onFailureActions,
	}, nil
}

func (a *AwaitTerminalSubmissionStatusAction) Execute() error {
	attempts := 0
	submissionStatus := "evaluating"

	for submissionStatus == "evaluating" && attempts < 10 {
		var err error

		codecraftersClient := client.NewCodecraftersClient()
		resp, err := codecraftersClient.FetchSubmission(a.SubmissionID)
		if err != nil {
			// We have retries, so we can proceed here anyway
			sentry.CaptureException(err)
		} else {
			submissionStatus = resp.Status
		}

		attempts += 1
		time.Sleep(time.Duration(100*attempts) * time.Millisecond)
	}

	switch submissionStatus {
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
		err := fmt.Errorf("unexpected submission status: %s", submissionStatus)
		sentry.CaptureException(err)

		printErr := PrintMessageAction{Color: "red", Text: "We couldn't fetch the results of your submission. Please try again?"}.Execute()
		if printErr != nil {
			return printErr
		}
		printErr = PrintMessageAction{Color: "red", Text: "Let us know at hello@codecrafters.io if this error persists."}.Execute()
		if printErr != nil {
			return printErr
		}
		printErr = PrintMessageAction{Color: "plain", Text: ""}.Execute()
		if printErr != nil {
			return printErr
		}

		return TerminateAction{ExitCode: 1}.Execute()
	}

	return nil
}
