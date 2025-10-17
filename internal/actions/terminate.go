package actions

import (
	"encoding/json"
	"os"
)

type TerminateAction struct {
	ExitCode int `json:"exit_code"`
}

func NewTerminateAction(argsJson json.RawMessage) (TerminateAction, error) {
	var terminateAction TerminateAction
	if err := json.Unmarshal(argsJson, &terminateAction); err != nil {
		return TerminateAction{}, err
	}

	return terminateAction, nil
}

func (a TerminateAction) Execute() error {
	os.Exit(a.ExitCode)

	return nil
}
