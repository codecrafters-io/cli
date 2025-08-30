package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
)

func UpdateBuildpackCommand(ctx context.Context) (err error) {
	logger := zerolog.Ctx(ctx)

	logger.Debug().Msg("update-buildpack command starts")
	defer func() {
		logger.Debug().Err(err).Msg("update-buildpack command ends")
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

		var noRepo utils.NoCodecraftersRemoteFoundError
		if errors.Is(err, &noRepo) {
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

	logger.Debug().Msg("reading and updating codecrafters.yml file")

	codecraftersYmlPath := filepath.Join(repoDir, "codecrafters.yml")

	content, err := os.ReadFile(codecraftersYmlPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("codecrafters.yml file not found in repository root")
		}
		return fmt.Errorf("failed to read codecrafters.yml: %w", err)
	}

	updatedContent := utils.ReplaceYAMLField(string(content), "language_pack", "buildpack")

	err = os.WriteFile(codecraftersYmlPath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated codecrafters.yml: %w", err)
	}

	buildpackValue := utils.ExtractYAMLFieldValue(updatedContent, "buildpack")
	if buildpackValue == "" {
		return fmt.Errorf("buildpack value not found in codecrafters.yml")
	}

	logger.Debug().Msg("fetching latest buildpack from server")

	codecraftersClient := utils.NewCodecraftersClient(codecraftersRemote.CodecraftersServerURL())
	buildpacksResponse, err := codecraftersClient.FetchBuildpacks(codecraftersRemote.CodecraftersRepositoryId())
	if err != nil {
		logger.Debug().Err(err).Msg("failed to fetch buildpacks")
		return fmt.Errorf("failed to fetch buildpacks: %w", err)
	}

	var latestBuildpack utils.BuildpackInfo
	for _, buildpack := range buildpacksResponse.Buildpacks {
		if buildpack.IsLatest {
			latestBuildpack = buildpack
			break
		}
	}

	logger.Debug().Msgf("current buildpack: %s, latest buildpack: %s", buildpackValue, latestBuildpack.Slug)

	if buildpackValue == latestBuildpack.Slug {
		fmt.Printf("Buildpack is already up to date (%s)\n", buildpackValue)
		return nil
	}

	fmt.Printf("Current buildpack: %s\n", buildpackValue)
	fmt.Printf("Do you want to upgrade to %s? (Press any key to proceed, CTRL-C to cancel)\n", latestBuildpack.Slug)

	reader := bufio.NewReader(os.Stdin)
	_, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	logger.Debug().Msg("calling update buildpack API")

	updateResponse, err := codecraftersClient.UpdateBuildpack(codecraftersRemote.CodecraftersRepositoryId())
	if err != nil {
		return fmt.Errorf("failed to update buildpack: %w", err)
	}

	if updateResponse.IsError {
		return fmt.Errorf("update failed: %s", updateResponse.ErrorMessage)
	}

	updatedContent = utils.ReplaceYAMLFieldValue(updatedContent, "buildpack", updateResponse.Buildpack.Slug)

	err = os.WriteFile(codecraftersYmlPath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated codecrafters.yml: %w", err)
	}

	fmt.Printf("Updated buildpack from %s to %s\n", buildpackValue, updateResponse.Buildpack.Slug)
	return nil
}
