package towerdiscord

import (
	"encoding/json"
)

type HTTPErrorResponse struct {
	Code    int             `json:"code"`
	Errors  json.RawMessage `json:"errors"`
	Message string          `json:"message"`
}

type Auth struct {
	Type  string
	Token string
}
