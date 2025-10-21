package commands

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/charmbracelet/glamour"
	"github.com/codecrafters-io/cli/internal/client"
	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
)

func TaskCommand(ctx context.Context, stageSlug string, raw bool) (err error) {
	logger := zerolog.Ctx(ctx)

	logger.Debug().Msg("task command starts")
	defer func() {
		logger.Debug().Err(err).Msg("task command ends")
	}()

	defer func() {
		if p := recover(); p != nil {
			logger.Panic().Str("panic", fmt.Sprintf("%v", p)).Stack().Msg("panic")
			sentry.CurrentHub().Recover(p)

			panic(p)
		}

		if err == nil {
			return
		}

		var noRepo *utils.NoCodecraftersRemoteFoundError
		if errors.As(err, &noRepo) {
			// ignore
			return
		}

		sentry.CurrentHub().CaptureException(err)
	}()

	logger.Debug().Msg("computing repository directory")

	repoDir, err := utils.GetRepositoryDir()
	if err != nil {
		return err
	}

	logger.Debug().Msgf("found repository directory: %s", repoDir)

	logger.Debug().Msg("identifying remotes")

	codecraftersRemote, err := utils.IdentifyGitRemote(repoDir)
	if err != nil {
		return err
	}

	logger.Debug().Msgf("identified remote: %s, %s", codecraftersRemote.Name, codecraftersRemote.Url)

	codecraftersClient := client.NewCodecraftersClient(codecraftersRemote.CodecraftersServerURL())

	logger.Debug().Msg("fetching stage list")

	stageListResponse, err := codecraftersClient.FetchStageList(codecraftersRemote.CodecraftersRepositoryId())
	if err != nil {
		return fmt.Errorf("fetch stage list: %w", err)
	}

	logger.Debug().Msgf("fetched %d stages", len(stageListResponse.Stages))

	var currentStageIndex int = -1
	for i := range stageListResponse.Stages {
		if stageListResponse.Stages[i].IsCurrent {
			currentStageIndex = i
			break
		}
	}

	if currentStageIndex == -1 {
		panic("no current stage found")
	}

	var targetStage *client.Stage

	if stageSlug == "" {
		targetStage = &stageListResponse.Stages[currentStageIndex]
	} else if offset, err := strconv.Atoi(stageSlug); err == nil {
		targetIndex := currentStageIndex + offset

		if targetIndex < 0 || targetIndex >= len(stageListResponse.Stages) {
			return buildStageError(fmt.Sprintf("Stage offset %d is out of range", offset), currentStageIndex, stageListResponse.Stages)
		}

		targetStage = &stageListResponse.Stages[targetIndex]
	} else {
		for i := range stageListResponse.Stages {
			if stageListResponse.Stages[i].Slug == stageSlug {
				targetStage = &stageListResponse.Stages[i]
				break
			}
		}

		if targetStage == nil {
			return buildStageError(fmt.Sprintf("Stage '%s' not found", stageSlug), currentStageIndex, stageListResponse.Stages)
		}
	}

	if raw {
		fmt.Println(targetStage.GetDocsMarkdown())
	} else {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to create renderer: %v", err))
		}

		rendered, err := renderer.Render(targetStage.GetDocsMarkdown())
		if err != nil {
			return fmt.Errorf("failed to render markdown: %w", err)
		}

		fmt.Print(rendered)
	}

	return nil
}

func buildStageError(message string, currentStageIndex int, stages []client.Stage) error {
	errorMsg := fmt.Sprintf("%s.\n\n", message)
	errorMsg += "Available stages:\n\n"

	for i, stage := range stages {
		marker := "  "
		if i == currentStageIndex {
			marker = "→ "
		}
		relativeOffset := i - currentStageIndex
		offsetStr := ""
		if relativeOffset > 0 {
			offsetStr = fmt.Sprintf(" (+%d)", relativeOffset)
		} else if relativeOffset < 0 {
			offsetStr = fmt.Sprintf(" (%d)", relativeOffset)
		}
		errorMsg += fmt.Sprintf("%s%d. [%s] %s%s\n", marker, i, stage.Slug, stage.Name, offsetStr)
	}

	return fmt.Errorf("%s", errorMsg)
}
