package actions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type PrintTerminalCommandsBoxAction struct {
	Commands []string `json:"commands"`
}

func NewPrintTerminalCommandsBoxAction(argsJson json.RawMessage) (PrintTerminalCommandsBoxAction, error) {
	var printTerminalCommandsBoxAction PrintTerminalCommandsBoxAction
	if err := json.Unmarshal(argsJson, &printTerminalCommandsBoxAction); err != nil {
		return PrintTerminalCommandsBoxAction{}, err
	}

	return printTerminalCommandsBoxAction, nil
}

func (a PrintTerminalCommandsBoxAction) Execute() error {
	lipgloss.SetColorProfile(termenv.ANSI256)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(0, 1).
		Align(lipgloss.Left).
		MaxWidth(70) // Terminal width (80) - "remote: "

	text := ""
	for _, command := range a.Commands {
		grayDollarSign := lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Render("$")
		text += fmt.Sprintf("%s %s\n", grayDollarSign, command)
	}

	text = strings.TrimSuffix(text, "\n")
	box := boxStyle.Render(text)

	fmt.Println(box)

	return nil
}

