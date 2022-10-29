package utils

import (
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"testing"
)

func TestIdentifyGitRemoteWithSingleProductionRemote(t *testing.T) {
	repositoryDir := createEmptyRepository(t)
	createRemote(t, repositoryDir, "origin", "https://git.codecrafters.io/dummy")

	remote, err := IdentifyGitRemote(repositoryDir)
	assert.Nil(t, err)

	assert.Equal(t, "origin", remote.Name)
	assert.Equal(t, "https://git.codecrafters.io/dummy", remote.Url)
	assert.Equal(t, "https://app.codecrafters.io", remote.CodecraftersServerURL())
}

func TestIdentifyGitRemoteWithSingleStagingRemote(t *testing.T) {
	repositoryDir := createEmptyRepository(t)
	createRemote(t, repositoryDir, "origin", "https://git.staging.codecrafters.io/dummy")

	remote, err := IdentifyGitRemote(repositoryDir)
	assert.Nil(t, err)

	assert.Equal(t, "origin", remote.Name)
	assert.Equal(t, "https://git.staging.codecrafters.io/dummy", remote.Url)
	assert.Equal(t, "https://app.staging.codecrafters.io", remote.CodecraftersServerURL())
}

func TestIdentifyGitRemoteWithSingleDevelopmentRemote(t *testing.T) {
	repositoryDir := createEmptyRepository(t)
	createRemote(t, repositoryDir, "origin", "https://codecrafters-paul-git.ngrok.io/dummy")

	remote, err := IdentifyGitRemote(repositoryDir)
	assert.Nil(t, err)

	assert.Equal(t, "origin", remote.Name)
	assert.Equal(t, "https://codecrafters-paul-git.ngrok.io/dummy", remote.Url)
	assert.Equal(t, "https://codecrafters-paul.ngrok.io", remote.CodecraftersServerURL())
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
	assert.EqualError(t, err, "Multiple CodeCrafters git remotes found: origin1, origin2")
}

func TestIdentifyGitRemoteWithNoCodecraftersRemotes(t *testing.T) {
	repositoryDir := createEmptyRepository(t)
	createRemote(t, repositoryDir, "gitlab", "https://gitlab.com/codecrafters-io/dummy1")
	createRemote(t, repositoryDir, "github", "https://github.com/codecrafters-io/dummy2")

	_, err := IdentifyGitRemote(repositoryDir)
	assert.EqualError(t, err, "No CodeCrafters git remotes found. Available remotes: github, gitlab")
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
