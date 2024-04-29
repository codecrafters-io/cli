package utils

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentifyGitRemoteWithSingleProductionRemote(t *testing.T) {
	repositoryDir := createEmptyRepository(t)
	createRemote(t, repositoryDir, "origin", "https://git.codecrafters.io/dummy")

	remote, err := IdentifyGitRemote(repositoryDir)
	assert.Nil(t, err)

	assert.Equal(t, "origin", remote.Name)
	assert.Equal(t, "https://git.codecrafters.io/dummy", remote.Url)
	assert.Equal(t, "https://backend.codecrafters.io", remote.CodecraftersServerURL())
}

func TestIdentifyGitRemoteWithSingleStagingRemote(t *testing.T) {
	repositoryDir := createEmptyRepository(t)
	createRemote(t, repositoryDir, "origin", "https://git-staging.codecrafters.io/dummy")

	remote, err := IdentifyGitRemote(repositoryDir)
	assert.Nil(t, err)

	assert.Equal(t, "origin", remote.Name)
	assert.Equal(t, "https://git-staging.codecrafters.io/dummy", remote.Url)
	assert.Equal(t, "https://backend-staging.codecrafters.io", remote.CodecraftersServerURL())
}

func TestIdentifyGitRemoteWithSingleDevelopmentRemote(t *testing.T) {
	repositoryDir := createEmptyRepository(t)
	createRemote(t, repositoryDir, "origin", "https://cc-paul-git.ngrok.io/dummy")

	remote, err := IdentifyGitRemote(repositoryDir)
	assert.Nil(t, err)

	assert.Equal(t, "origin", remote.Name)
	assert.Equal(t, "https://cc-paul-git.ngrok.io/dummy", remote.Url)
	assert.Equal(t, "https://paul-backend.ccdev.dev", remote.CodecraftersServerURL())
}

func TestIdentifyGitRemoteWithMultipleRemotes(t *testing.T) {
	repositoryDir := createEmptyRepository(t)
	createRemote(t, repositoryDir, "origin", "https://git.codecrafters.io/dummy1")
	createRemote(t, repositoryDir, "github", "https://github.com/codecrafters-io/dummy2")

	remote, err := IdentifyGitRemote(repositoryDir)
	assert.Nil(t, err)

	assert.Equal(t, "origin", remote.Name)
	assert.Equal(t, "https://git.codecrafters.io/dummy1", remote.Url)
}

func TestIdentifyGitRemoteWithMultipleCodecraftersRemotes(t *testing.T) {
	repositoryDir := createEmptyRepository(t)
	createRemote(t, repositoryDir, "origin1", "https://git.codecrafters.io/dummy1")
	createRemote(t, repositoryDir, "origin2", "https://git.codecrafters.io/dummy2")

	_, err := IdentifyGitRemote(repositoryDir)
	assert.EqualError(t, err, "Multiple CodeCrafters git remotes found: https://git.codecrafters.io/dummy1, https://git.codecrafters.io/dummy2")
}

func TestIdentifyGitRemoteWithNoCodecraftersRemotes(t *testing.T) {
	repositoryDir := createEmptyRepository(t)
	createRemote(t, repositoryDir, "gitlab", "https://gitlab.com/codecrafters-io/dummy1")
	createRemote(t, repositoryDir, "github", "https://github.com/codecrafters-io/dummy2")

	_, err := IdentifyGitRemote(repositoryDir)
	assert.EqualError(t, err, "No CodeCrafters git remotes found. Available remotes: https://github.com/codecrafters-io/dummy2, https://gitlab.com/codecrafters-io/dummy1\nPlease run this command from within your CodeCrafters Git repository.")
}

func createEmptyRepository(t *testing.T) string {
	repositoryDir, err := os.MkdirTemp("", "cc-test")
	if err != nil {
		t.Fatal(err)
	}

	err = exec.Command("git", "-C", repositoryDir, "init").Run()
	if err != nil {
		t.Fatal(err)
	}

	return repositoryDir
}

func createRemote(t *testing.T, repositoryDir string, remoteName string, remoteUrl string) {
	err := exec.Command("git", "-C", repositoryDir, "remote", "add", remoteName, remoteUrl).Run()
	if err != nil {
		t.Fatal(err)
	}
}
