package actions

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	logstream_redis "github.com/codecrafters-io/logstream/redis"
)

type StreamLogsAction struct {
	LogstreamURL string `json:"logstream_url"`
}

func NewStreamLogsAction(argsJson json.RawMessage) (StreamLogsAction, error) {
	var streamLogsAction StreamLogsAction
	if err := json.Unmarshal(argsJson, &streamLogsAction); err != nil {
		return StreamLogsAction{}, err
	}

	return streamLogsAction, nil
}

func (a StreamLogsAction) Execute() error {
	consumer, err := logstream_redis.NewConsumer(a.LogstreamURL)
	if err != nil {
		return fmt.Errorf("failed to create logstream consumer: %w", err)
	}

	_, err = io.Copy(os.Stdout, consumer)
	if err != nil {
		return fmt.Errorf("failed to read from stream: %w", err)
	}

	return nil
}
