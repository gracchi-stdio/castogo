package pageview

import "encoding/json"

func parseBlockContent[T any](content json.RawMessage) (*T, error) {
	var result T
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
