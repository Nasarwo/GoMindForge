package ai

import "errors"

var (
	ErrProviderNotFound = errors.New("ai provider not found")
	ErrAPIKeyMissing    = errors.New("api key is missing")
	ErrAPIRequestFailed = errors.New("api request failed")
)
