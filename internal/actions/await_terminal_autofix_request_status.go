package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/codecrafters-io/cli/internal/client"
	"github.com/getsentry/sentry-go"
)

type AwaitTerminalAutofixRequestStatusAction struct {
	InProgressActions []Action
	OnFailureActions  []Action
	OnSuccessActions  []Action
	SubmissionID      string
}

type AwaitTerminalAutofixRequestStatusActionArgs struct {
	InProgressActions []client.ActionDefinition `json:"in_progress_actions"`
	OnFailureActions  []client.ActionDefinition `json:"on_failure_actions"`
	OnSuccessActions  []client.ActionDefinition `json:"on_success_actions"`
	SubmissionID      string                    `json:"submission_id"`
}

func NewAwaitTerminalAutofixRequestStatusAction(argsJson json.RawMessage) (*AwaitTerminalAutofixRequestStatusAction, error) {
	var awaitTerminalAutofixRequestStatusActionArgs AwaitTerminalAutofixRequestStatusActionArgs
	if err := json.Unmarshal(argsJson, &awaitTerminalAutofixRequestStatusActionArgs); err != nil {
		return nil, err
	}

	onSuccessActions := []Action{}
	for _, actionDefinition := range awaitTerminalAutofixRequestStatusActionArgs.OnSuccessActions {
		action, err := ActionFromDefinition(actionDefinition)
		if err != nil {
			return nil, err
		}

		onSuccessActions = append(onSuccessActions, action)
	}

	onFailureActions := []Action{}
	for _, actionDefinition := range awaitTerminalAutofixRequestStatusActionArgs.OnFailureActions {
		action, err := ActionFromDefinition(actionDefinition)
		if err != nil {
			return nil, err
		}

		onFailureActions = append(onFailureActions, action)
	}

	inProgressActions := []Action{}
	for _, actionDefinition := range awaitTerminalAutofixRequestStatusActionArgs.InProgressActions {
		action, err := ActionFromDefinition(actionDefinition)
		if err != nil {
			return nil, err
		}

		inProgressActions = append(inProgressActions, action)
	}

	return &AwaitTerminalAutofixRequestStatusAction{
		InProgressActions: inProgressActions,
		OnFailureActions:  onFailureActions,
		OnSuccessActions:  onSuccessActions,
		SubmissionID:      awaitTerminalAutofixRequestStatusActionArgs.SubmissionID,
	}, nil
}

func (a *AwaitTerminalAutofixRequestStatusAction) Execute() error {
	attempts := 0
	autofixRequestStatus := "in_progress"

	inProgressActionsDoneCh := make(chan bool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := a.executeInProgressActions(ctx); err != nil {
			sentry.CaptureException(err)
		}

		inProgressActionsDoneCh <- true
	}()

	// We wait for upto 60 seconds (+ the time it takes to fetch status each time)
	for autofixRequestStatus == "in_progress" && attempts < 60 {
		var err error

		codecraftersClient := client.NewCodecraftersClient()
		autofixRequestStatusResponse, err := codecraftersClient.FetchAutofixRequest(a.SubmissionID)
		if err != nil {
			// We can still proceed here anyway
			sentry.CaptureException(err)
		} else {
			autofixRequestStatus = autofixRequestStatusResponse.Status
		}

		attempts += 1
		time.Sleep(time.Second)
	}

	// Ensure interruptible actions (like printing progress bars) finish early
	cancel()
	<-inProgressActionsDoneCh

	switch autofixRequestStatus {
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
		err := fmt.Errorf("unexpected autofix request status: %s", autofixRequestStatus)
		sentry.CaptureException(err)

		PrintMessageAction{Color: "red", Text: "We failed to analyze your test failure in time. Please try again?"}.Execute()
		PrintMessageAction{Color: "red", Text: "Let us know at hello@codecrafters.io if this error persists."}.Execute()

		// This is an internal error, let's terminate
		TerminateAction{ExitCode: 1}.Execute()
	}

	return nil
}

func (a *AwaitTerminalAutofixRequestStatusAction) executeInProgressActions(ctx context.Context) error {
	for _, action := range a.InProgressActions {
		if interruptibleAction, ok := action.(InterruptibleAction); ok {
			if err := interruptibleAction.ExecuteWithContext(ctx); err != nil {
				return err
			}
		} else {
			if err := action.Execute(); err != nil {
				return err
			}
		}
	}

	return nil
}
