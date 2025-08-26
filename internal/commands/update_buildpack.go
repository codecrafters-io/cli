package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/rs/zerolog"
)

func UpdateBuildpackCommand(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)

	logger.Debug().Msg("update-buildpack command starts")

	logger.Debug().Msg("computing repository directory")

	repoDir, err := utils.GetRepositoryDir()
	if err != nil {
		return err
	}

	logger.Debug().Msgf("found repository directory: %s", repoDir)

	logger.Debug().Msg("identifying remotes")

	_, err = utils.IdentifyGitRemote(repoDir)
	if err != nil {
		return err
	}

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

	return nil
}
