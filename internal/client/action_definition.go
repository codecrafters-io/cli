package client

import (
	"encoding/json"
)

type ActionDefinition struct {
	Type string          `json:"type"`
	Args json.RawMessage `json:"args"`
}
