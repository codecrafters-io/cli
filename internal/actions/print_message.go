package actions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mitchellh/go-wordwrap"
)

type PrintMessageAction struct {
	Color string `json:"color"`
	Text  string `json:"text"`
}

func NewPrintMessageAction(argsJson json.RawMessage) (PrintMessageAction, error) {
	var printMessageAction PrintMessageAction
	if err := json.Unmarshal(argsJson, &printMessageAction); err != nil {
		return PrintMessageAction{}, err
	}

	return printMessageAction, nil
}

func (a PrintMessageAction) Execute() error {
	wrapped := wordwrap.WrapString(a.Text, 79)

	lineFormat := "%s\n"

	switch a.Color {
	case "red":
		lineFormat = "\033[31m%s\033[0m\n"
	case "green":
		lineFormat = "\033[32m%s\033[0m\n"
	case "yellow":
		lineFormat = "\033[33m%s\033[0m\n"
	case "blue":
		lineFormat = "\033[34m%s\033[0m\n"
	case "plain":
		lineFormat = "%s\n"
	default:
		return fmt.Errorf("invalid color: %s", a.Color)
	}

	for _, line := range strings.Split(wrapped, "\n") {
		fmt.Printf(lineFormat, line)
	}

	return nil
}
