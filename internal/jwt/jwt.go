package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type Payload struct {
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
	Iss string `json:"iss"`
	Sub string `json:"sub"`
}

func ExtractExp(token string) (int64, error) {
	payload, err := DecodePayload(token)
	if err != nil {
		return 0, err
	}

	if payload.Exp == 0 {
		return 0, fmt.Errorf("no exp claim in JWT")
	}

	return payload.Exp, nil
}

func DecodePayload(token string) (*Payload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	payloadB64 := parts[1]
	if l := len(payloadB64) % 4; l != 0 {
		payloadB64 += strings.Repeat("=", 4-l)
	}

	decoded, err := base64.URLEncoding.DecodeString(payloadB64)
	if err != nil {
		decoded, err = base64.StdEncoding.DecodeString(payloadB64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
		}
	}

	var payload Payload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse JWT payload: %w", err)
	}

	return &payload, nil
}
