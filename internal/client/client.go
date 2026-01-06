package client

import (
	"github.com/codecrafters-io/cli/internal/globals"
	"github.com/codecrafters-io/cli/internal/utils"
)

type CodecraftersClient struct {
	ServerUrl string
}

func NewCodecraftersClient() CodecraftersClient {
	return CodecraftersClient{
		ServerUrl: globals.GetCodecraftersServerURL(),
	}
}

func (c CodecraftersClient) headers() map[string]string {
	return map[string]string{
		"X-Codecrafters-CLI-Version": utils.VersionString(),
	}
}
