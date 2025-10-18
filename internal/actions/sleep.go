package actions

import (
	"encoding/json"
	"time"
)

type SleepAction struct {
	DurationInMilliseconds int `json:"duration_in_milliseconds"`
}

func NewSleepAction(argsJson json.RawMessage) (SleepAction, error) {
	var sleepAction SleepAction
	if err := json.Unmarshal(argsJson, &sleepAction); err != nil {
		return SleepAction{}, err
	}

	return sleepAction, nil
}

func (a SleepAction) Execute() error {
	time.Sleep(time.Duration(a.DurationInMilliseconds) * time.Millisecond)

	return nil
}
