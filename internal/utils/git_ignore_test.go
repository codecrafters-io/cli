package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitIgnore(t *testing.T) {
	t.Run("without gitignore files", func(t *testing.T) {
		gitIgnore := NewGitIgnore(t.TempDir())
		assertFileNotSkipped(t, &gitIgnore, "some/random/file.txt")
	})

	t.Run("with local gitignore", func(t *testing.T) {
		tmpRepoDir := t.TempDir()
		writeFile(t, filepath.Join(tmpRepoDir, ".gitignore"), "ignore/this/file.txt")

		gitIgnore := NewGitIgnore(tmpRepoDir)
		assertFileSkipped(t, &gitIgnore, "ignore/this/file.txt")
		assertFileNotSkipped(t, &gitIgnore, "some/other/file.txt")
	})

	t.Run("with global gitignore", func(t *testing.T) {
		backup := setupGlobalGitIgnore(t, "ignore/this/file.txt")
		defer func() {
			if backup.originalPath == "" {
				unsetGlobalGitIgnoreConfig(t)
			}
			backup.Restore(t)
		}()

		gitIgnore := NewGitIgnore(t.TempDir())
		assertFileSkipped(t, &gitIgnore, "ignore/this/file.txt")
		assertFileNotSkipped(t, &gitIgnore, "some/other/file.txt")
	})

	t.Run("with git info exclude", func(t *testing.T) {
		tmpRepoDir := createEmptyRepository(t)
		backup := setupGitInfoExclude(t, tmpRepoDir, "ignore/this/file.txt")
		defer backup.Restore(t)

		gitIgnore := NewGitIgnore(tmpRepoDir)
		assertFileSkipped(t, &gitIgnore, "ignore/this/file.txt")
		assertFileNotSkipped(t, &gitIgnore, "some/other/file.txt")
	})
}

func assertFileSkipped(t *testing.T, gitIgnore *GitIgnore, path string) {
	skip, err := gitIgnore.SkipFile(path)
	assert.NoError(t, err)
	assert.True(t, skip)
}

func assertFileNotSkipped(t *testing.T, gitIgnore *GitIgnore, path string) {
	skip, err := gitIgnore.SkipFile(path)
	assert.NoError(t, err)
	assert.False(t, skip)
}

type FileBackup struct {
	originalPath string
	backupPath   string
}

func (b *FileBackup) Restore(t *testing.T) {
	if b.originalPath != "" {
		moveFile(t, b.backupPath, b.originalPath)
	}
}

func setupGlobalGitIgnore(t *testing.T, content string) *FileBackup {
	globalGitIgnorePath := getGlobalGitIgnorePath()
	backupPath := filepath.Join(t.TempDir(), ".gitignore_global")

	if globalGitIgnorePath == "" {
		writeFile(t, backupPath, content)
		setGlobalGitIgnoreConfig(t, backupPath)
		return &FileBackup{originalPath: "", backupPath: backupPath}
	}

	moveFile(t, globalGitIgnorePath, backupPath)
	writeFile(t, globalGitIgnorePath, content)
	return &FileBackup{originalPath: globalGitIgnorePath, backupPath: backupPath}
}

func setupGitInfoExclude(t *testing.T, baseDir string, content string) *FileBackup {
	gitInfoExcludePath := filepath.Join(baseDir, ".git", "info", "exclude")
	_, err := os.Stat(gitInfoExcludePath)
	assert.NoError(t, err)

	backupPath := filepath.Join(t.TempDir(), ".git_info_exclude_backup")
	moveFile(t, gitInfoExcludePath, backupPath)
	writeFile(t, gitInfoExcludePath, content)
	return &FileBackup{originalPath: gitInfoExcludePath, backupPath: backupPath}
}

func setGlobalGitIgnoreConfig(t *testing.T, path string) {
	_, err := exec.Command("git", "config", "--global", "core.excludesfile", path).Output()
	assert.NoError(t, err)
}

func unsetGlobalGitIgnoreConfig(t *testing.T) {
	_, err := exec.Command("git", "config", "--global", "--unset", "core.excludesfile").Output()
	assert.NoError(t, err)
}

func moveFile(t *testing.T, srcPath string, dstPath string) {
	err := os.Rename(srcPath, dstPath)
	assert.NoError(t, err)
}

func writeFile(t *testing.T, path string, content string) {
	err := os.WriteFile(path, []byte(content), 0644)
	assert.NoError(t, err)
}
