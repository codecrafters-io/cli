package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// The maximum delay between prints in seconds
const maxDelayBetweenPrintsInSeconds = 2

// The minimum number of times we will print the progress bar
const minPrintCount = 5

// The length of the progress bar
const progressBarLength = 20

type PrintProgressBarAction struct {
	ExpectedDelayInSeconds int `json:"expected_delay_in_seconds"`
}

func NewPrintProgressBarAction(argsJson json.RawMessage) (PrintProgressBarAction, error) {
	var printProgressBarAction PrintProgressBarAction
	if err := json.Unmarshal(argsJson, &printProgressBarAction); err != nil {
		return PrintProgressBarAction{}, err
	}

	return printProgressBarAction, nil
}

func (a PrintProgressBarAction) numberOfPrints() int {
	return max(a.ExpectedDelayInSeconds/maxDelayBetweenPrintsInSeconds, minPrintCount)
}

func (a PrintProgressBarAction) delayBetweenPrintsInMilliseconds(printsLeft int) int {
	if printsLeft <= 1 {
		return 60 * 1000 // 60 seconds (The action should timeout by then)
	}

	return min(maxDelayBetweenPrintsInSeconds*1000, 1000*a.ExpectedDelayInSeconds/a.numberOfPrints())
}

func (a PrintProgressBarAction) Execute() error {
	return a.ExecuteWithContext(context.Background())
}

func (a PrintProgressBarAction) ExecuteWithContext(ctx context.Context) error {
	contextIsCancelled := false
	lastPrintedPercentage := 0

	for i := 0; i < a.numberOfPrints(); i++ {
		percentage := (i + 1) * 100 / a.numberOfPrints()
		numberOfBars := percentage * progressBarLength / 100
		numberOfSpaces := progressBarLength - numberOfBars

		bars := strings.Repeat("=", numberOfBars)
		if numberOfSpaces > 0 {
			bars = bars[:len(bars)-1] + ">"
		}

		// Print with a random jitter of up to 5%
		percentageToPrint := percentage

		// If this is in-progress, add a random jitter of up to 5%
		if percentage < 100 {
			percentageToPrint = percentage - 5 + rand.Intn(10)
		}

		// Ensure the printed percentage is not greater than 100
		if percentageToPrint > 100 {
			percentageToPrint = 100
		}

		// Ensure the printed percentage always increases
		if percentageToPrint < lastPrintedPercentage {
			percentageToPrint = lastPrintedPercentage
		}

		// Use ANSI color codes for green (same pattern as print_message.go)
		greenStart := "\033[32m"
		greenEnd := "\033[0m"
		fmt.Printf("[%s%s%s%s] %s%s%s\n", greenStart, bars, greenEnd, strings.Repeat(" ", numberOfSpaces), greenStart, fmt.Sprintf("%d%%", percentageToPrint), greenEnd)
		lastPrintedPercentage = percentageToPrint

		// If the context is cancelled, keep looping until we print all bars and exit (with no delay)
		if contextIsCancelled {
			continue
		}

		sleepDurationInMilliseconds := a.delayBetweenPrintsInMilliseconds(a.numberOfPrints() - (i + 1))

		// Add a random jitter of up to 500ms
		sleepDurationInMilliseconds = sleepDurationInMilliseconds - 500 + rand.Intn(1000)

		// If the context is still active, sleep for the max delay between prints
		select {
		case <-ctx.Done():
			contextIsCancelled = true
			continue
		case <-time.After(time.Duration(sleepDurationInMilliseconds) * time.Millisecond):
			continue
		}
	}

	return nil
}
