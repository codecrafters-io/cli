package client

import (
	"encoding/json"
)

type ActionDefinition struct {
	Type string          `json:"type"`
	Args json.RawMessage `json:"args"`
}

type BuildpackInfo struct {
	Slug     string `json:"slug"`
	IsLatest bool   `json:"is_latest"`
}
