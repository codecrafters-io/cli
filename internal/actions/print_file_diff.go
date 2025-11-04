package actions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type PrintFileDiffAction struct {
	DiffStr  string `json:"diff_str"`
	FilePath string `json:"file_path"`
}

func NewPrintFileDiffAction(argsJson json.RawMessage) (PrintFileDiffAction, error) {
	var printFileDiffAction PrintFileDiffAction
	if err := json.Unmarshal(argsJson, &printFileDiffAction); err != nil {
		return PrintFileDiffAction{}, err
	}

	return printFileDiffAction, nil
}

// TODO: Handle printing chunks!
func (a PrintFileDiffAction) Execute() error {
	lipgloss.SetColorProfile(termenv.ANSI256)

	diffBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(0, 1).
		Align(lipgloss.Left).
		Width(70) // Terminal width (80) - "remote: "

	diffLines := strings.Split(a.DiffStr, "\n")
	formattedDiffLines := make([]string, len(diffLines))

	for i, line := range diffLines {
		if strings.HasPrefix(line, "+") {
			formattedDiffLines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render(line)
		} else if strings.HasPrefix(line, "-") {
			formattedDiffLines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(line)
		} else {
			formattedDiffLines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Render(line)
		}
	}

	diffContent := lipgloss.JoinVertical(lipgloss.Left, formattedDiffLines...)
	fmt.Println(diffBoxStyle.Render(diffContent))

	return nil
}

