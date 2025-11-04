package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/cli/internal/client"
	"github.com/codecrafters-io/cli/internal/globals"
	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/getsentry/sentry-go"
)

func UpdateBuildpackCommand() (err error) {
	utils.Logger.Debug().Msg("update-buildpack command starts")
	defer func() {
		utils.Logger.Debug().Err(err).Msg("update-buildpack command ends")
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

	utils.Logger.Debug().Msg("fetching current buildpack from server")

	globals.SetCodecraftersServerURL(codecraftersRemote.CodecraftersServerURL())
	codecraftersClient := client.NewCodecraftersClient()

	repositoryBuildpackResponse, err := codecraftersClient.FetchRepositoryBuildpack(codecraftersRemote.CodecraftersRepositoryId())
	if err != nil {
		utils.Logger.Debug().Err(err).Msg("failed to fetch repository buildpack")
		return fmt.Errorf("failed to fetch repository buildpack: %w", err)
	}

	currentBuildpackSlug := repositoryBuildpackResponse.Buildpack.Slug

	utils.Logger.Debug().Msg("fetching latest buildpack from server")

	buildpacksResponse, err := codecraftersClient.FetchBuildpacks(codecraftersRemote.CodecraftersRepositoryId())
	if err != nil {
		utils.Logger.Debug().Err(err).Msg("failed to fetch buildpacks")
		return fmt.Errorf("failed to fetch buildpacks: %w", err)
	}

	var latestBuildpack client.BuildpackInfo
	for _, buildpack := range buildpacksResponse.Buildpacks {
		if buildpack.IsLatest {
			latestBuildpack = buildpack
			break
		}
	}

	utils.Logger.Debug().Msgf("current buildpack: %s, latest buildpack: %s", currentBuildpackSlug, latestBuildpack.Slug)

	if currentBuildpackSlug == latestBuildpack.Slug {
		fmt.Printf("Buildpack is already up to date (%s)\n", currentBuildpackSlug)
		fmt.Println("Let us know at hello@codecrafters.io if you’d like us to upgrade the supported buildpack version.")
		return nil
	}

	fmt.Printf("Current buildpack: %s\n", currentBuildpackSlug)
	fmt.Printf("Do you want to upgrade to %s? (Press any key to proceed, CTRL-C to cancel)\n", latestBuildpack.Slug)

	reader := bufio.NewReader(os.Stdin)
	_, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	utils.Logger.Debug().Msg("calling update buildpack API")

	updateResponse, err := codecraftersClient.UpdateBuildpack(codecraftersRemote.CodecraftersRepositoryId())
	if err != nil {
		return fmt.Errorf("failed to update buildpack: %w", err)
	}

	if updateResponse.IsError {
		return fmt.Errorf("update failed: %s", updateResponse.ErrorMessage)
	}

	utils.Logger.Debug().Msg("reading and updating codecrafters.yml file")

	codecraftersYmlPath := filepath.Join(repoDir, "codecrafters.yml")

	content, err := os.ReadFile(codecraftersYmlPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("codecrafters.yml file not found in repository root")
		}
		return fmt.Errorf("failed to read codecrafters.yml: %w", err)
	}

	updatedContent := utils.ReplaceYAMLField(string(content), "language_pack", "buildpack")
	updatedContent = utils.ReplaceYAMLFieldValue(updatedContent, "buildpack", updateResponse.Buildpack.Slug)

	err = os.WriteFile(codecraftersYmlPath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated codecrafters.yml: %w", err)
	}

	fmt.Printf("Updated buildpack from %s to %s\n", currentBuildpackSlug, updateResponse.Buildpack.Slug)
	return nil
}
