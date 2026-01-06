package commands

import (
	"errors"
	"fmt"

	"github.com/codecrafters-io/cli/internal/actions"
	"github.com/codecrafters-io/cli/internal/client"
	"github.com/codecrafters-io/cli/internal/globals"
	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/getsentry/sentry-go"
)

func PingCommand() (err error) {
	utils.Logger.Debug().Msg("ping command starts")

	defer func() {
		utils.Logger.Debug().Err(err).Msg("ping command ends")
	}()

	defer func() {
		if p := recover(); p != nil {
			utils.Logger.Panic().Str("panic", fmt.Sprintf("%v", p)).Stack().Msg("panic")
			sentry.CurrentHub().Recover(p)

			panic(p)
		}

		if err == nil {
			return
		}

		var noRepo utils.NoCodecraftersRemoteFoundError
		if errors.Is(err, &noRepo) {
			// ignore
			return
		}

		sentry.CurrentHub().CaptureException(err)
	}()

	utils.Logger.Debug().Msg("computing repository directory")

	repoDir, err := utils.GetRepositoryDir()
	if err != nil {
		return err
	}

	utils.Logger.Debug().Msgf("found repository directory: %s", repoDir)

	utils.Logger.Debug().Msg("identifying remotes")

	codecraftersRemote, err := utils.IdentifyGitRemote(repoDir)
	if err != nil {
		return err
	}

	utils.Logger.Debug().Msgf("identified remote: %s, %s", codecraftersRemote.Name, codecraftersRemote.Url)

	globals.SetCodecraftersServerURL(codecraftersRemote.CodecraftersServerURL())
	codecraftersClient := client.NewCodecraftersClient()

	utils.Logger.Debug().Msg("sending ping request")

	pingResponse, err := codecraftersClient.Ping(codecraftersRemote.CodecraftersRepositoryId())
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	utils.Logger.Debug().Msgf("received %d actions", len(pingResponse.Actions))

	for _, actionDefinition := range pingResponse.Actions {
		action, err := actions.ActionFromDefinition(actionDefinition)
		if err != nil {
			return fmt.Errorf("parse action: %w", err)
		}

		if err := action.Execute(); err != nil {
			return fmt.Errorf("execute action: %w", err)
		}
	}

	return nil
}
