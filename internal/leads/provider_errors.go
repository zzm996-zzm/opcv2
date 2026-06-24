package leads

import "errors"

const (
	ErrorProviderTimeout     = "provider_timeout"
	ErrorProviderRateLimited = "provider_rate_limited"
)

var (
	ErrProviderTimeout     = errors.New("lead provider timeout")
	ErrProviderRateLimited = errors.New("lead provider rate limited")
)
