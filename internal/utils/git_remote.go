package utils

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type GitRemote struct {
	Url  string
	Name string
}

func (r GitRemote) IsCodecrafters() bool {
	return r.CodecraftersServerURL() != ""
}

func (r GitRemote) CodecraftersRepositoryId() string {
	return strings.Split(r.Url, "/")[len(strings.Split(r.Url, "/"))-1]
}

func (r GitRemote) CodecraftersServerURL() string {
	if strings.Contains(r.Url, "git.codecrafters.io") {
		return "https://backend.codecrafters.io"
	}

	if strings.Contains(r.Url, "git-staging.codecrafters.io") {
		return "https://backend-staging.codecrafters.io"
	}

	ngrokDevServerRegex := regexp.MustCompile(`cc-([\w-]+).ngrok.io`)

	// cc-paul-git.ngrok.io -> paul-backend.ccdev.dev
	if ngrokDevServerRegex.MatchString(r.Url) {
		replacedUrl := regexp.MustCompile(`\-git.ngrok.io/.*`).ReplaceAllString(r.Url, "-backend.ccdev.dev") // cc-paul-backend.ccdev.dev
		replacedUrl = regexp.MustCompile("cc-").ReplaceAllString(replacedUrl, "")                            // paul-backend.ccdev.dev
		return replacedUrl
	}

	cloudflareDevServerRegex := regexp.MustCompile("(.*)-git.ccdev.dev")

	// cc-paul-git.ccdev.dev -> paul.ccdev.dev
	if cloudflareDevServerRegex.MatchString(r.Url) {
		replacedUrl := regexp.MustCompile("-git").ReplaceAllString(r.Url, "-backend")
		replacedUrl = regexp.MustCompile("ccdev.dev/.*").ReplaceAllString(replacedUrl, "ccdev.dev")
		return replacedUrl
	}

	return ""
}

type NoCodecraftersRemoteFoundError struct {
	error
	Remotes []GitRemote
}

func (e NoCodecraftersRemoteFoundError) Error() string {
	remoteUrls := []string{}
	for _, remote := range e.Remotes {
		remoteUrls = append(remoteUrls, remote.Url)
	}

	return fmt.Sprintf("No CodeCrafters git remotes found. Available remotes: %s\nPlease run this command from within your CodeCrafters Git repository.", strings.Join(remoteUrls, ", "))
}

type MultipleCodecraftersRemotesFoundError struct {
	Remotes []GitRemote
}

func (e MultipleCodecraftersRemotesFoundError) Error() string {
	remoteUrls := []string{}
	for _, remote := range e.Remotes {
		remoteUrls = append(remoteUrls, remote.Url)
	}

	return "Multiple CodeCrafters git remotes found: " + strings.Join(remoteUrls, ", ")
}

func IdentifyGitRemote(repositoryDir string) (GitRemote, error) {
	remotes, err := listRemotes(repositoryDir)
	if err != nil {
		return GitRemote{}, err
	}

	codecraftersRemotes := []GitRemote{}

	for _, remote := range remotes {
		if remote.IsCodecrafters() {
			codecraftersRemotes = append(codecraftersRemotes, remote)
		}
	}

	if len(codecraftersRemotes) == 0 {
		return GitRemote{}, NoCodecraftersRemoteFoundError{Remotes: remotes}
	}

	if len(codecraftersRemotes) > 1 {
		return GitRemote{}, MultipleCodecraftersRemotesFoundError{Remotes: codecraftersRemotes}
	}

	return codecraftersRemotes[0], nil
}

func listRemotes(repositoryDir string) ([]GitRemote, error) {
	outputBytes, err := exec.Command("git", "-C", repositoryDir, "remote", "-v").Output()
	if err != nil {
		return []GitRemote{}, err
	}

	remoteLineRegex := regexp.MustCompile(`^(\S+)\s+(\S+)\s+`)
	remotes := []GitRemote{}

OUTER:
	for _, line := range strings.Split(string(outputBytes), "\n") {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if remoteLineRegex.MatchString(line) {
			matches := remoteLineRegex.FindStringSubmatch(line)
			remote := GitRemote{Name: matches[1], Url: matches[2]}

			for _, existingRemote := range remotes {
				if existingRemote.Url == remote.Url {
					continue OUTER
				}
			}

			remotes = append(remotes, remote)
		}
	}

	return remotes, nil
}
