package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	logstream_redis "github.com/codecrafters-io/logstream/redis"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
)

func HandleSubmission(createSubmissionResponse CreateSubmissionResponse, ctx context.Context, codecraftersClient CodecraftersClient) (err error) {
	logger := zerolog.Ctx(ctx)

	for _, message := range createSubmissionResponse.OnInitMessages {
		fmt.Println("")
		message.Print()
	}

	if createSubmissionResponse.BuildLogstreamURL != "" {
		logger.Debug().Msgf("streaming build logs from %s", createSubmissionResponse.BuildLogstreamURL)

		fmt.Println("")
		err = streamLogs(createSubmissionResponse.BuildLogstreamURL)
		if err != nil {
			return fmt.Errorf("stream build logs: %w", err)
		}

		logger.Debug().Msg("Finished streaming build logs")
		logger.Debug().Msg("fetching build")

		fetchBuildResponse, err := codecraftersClient.FetchBuild(createSubmissionResponse.BuildID)
		if err != nil {
			// TODO: Notify sentry
			red := color.New(color.FgRed).SprintFunc()
			fmt.Fprintln(os.Stderr, red(err.Error()))
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, red("We couldn't fetch the results of your submission. Please try again?"))
			fmt.Fprintln(os.Stderr, red("Let us know at hello@codecrafters.io if this error persists."))
			return err
		}

		logger.Debug().Msgf("finished fetching build: %v", fetchBuildResponse)
		red := color.New(color.FgRed).SprintFunc()

		switch fetchBuildResponse.Status {
		case "failure":
			fmt.Fprintln(os.Stderr, red(""))
			fmt.Fprintln(os.Stderr, red("Looks like your codebase failed to build."))
			fmt.Fprintln(os.Stderr, red("If you think this is a CodeCrafters error, please let us know at hello@codecrafters.io."))
			fmt.Fprintln(os.Stderr, red(""))
			os.Exit(0)
		case "success":
			time.Sleep(1 * time.Second) // The delay in-between build and test logs is usually 5-10 seconds, so let's buy some time
		default:
			red := color.New(color.FgRed).SprintFunc()

			fmt.Fprintln(os.Stderr, red("We couldn't fetch the results of your build. Please try again?"))
			fmt.Fprintln(os.Stderr, red("Let us know at hello@codecrafters.io if this error persists."))
			os.Exit(1)
		}
	}

	fmt.Println("")
	fmt.Println("Running tests. Logs should appear shortly...")
	fmt.Println("")

	err = streamLogs(createSubmissionResponse.LogstreamURL)
	if err != nil {
		return fmt.Errorf("stream logs: %w", err)
	}

	logger.Debug().Msgf("fetching submission %s", createSubmissionResponse.Id)

	fetchSubmissionResponse, err := codecraftersClient.FetchSubmission(createSubmissionResponse.Id)
	if err != nil {
		// TODO: Notify sentry
		red := color.New(color.FgRed).SprintFunc()
		fmt.Fprintln(os.Stderr, red(err.Error()))
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, red("We couldn't fetch the results of your submission. Please try again?"))
		fmt.Fprintln(os.Stderr, red("Let us know at hello@codecrafters.io if this error persists."))
		return err
	}

	logger.Debug().Msgf("finished fetching submission, status: %s", fetchSubmissionResponse.Status)

	switch fetchSubmissionResponse.Status {
	case "failure":
		for _, message := range createSubmissionResponse.OnFailureMessages {
			fmt.Println("")
			message.Print()
		}
	case "success":
		for _, message := range createSubmissionResponse.OnSuccessMessages {
			fmt.Println("")
			message.Print()
		}
	default:
		fmt.Println("")
	}

	if fetchSubmissionResponse.IsError {
		return fmt.Errorf("%s", fetchSubmissionResponse.ErrorMessage)
	}

	return nil
}

func streamLogs(logstreamUrl string) error {
	consumer, err := logstream_redis.NewConsumer(logstreamUrl)
	if err != nil {
		return fmt.Errorf("new log consumer: %w", err)
	}

	_, err = io.Copy(os.Stdout, consumer)
	if err != nil {
		return fmt.Errorf("stream data: %w", err)
	}

	return nil
}
