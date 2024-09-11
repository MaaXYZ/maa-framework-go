package maa

import (
	"encoding/json"
)

type J map[string]any

func toJSON(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
