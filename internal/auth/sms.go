package auth

import (
	"context"
	"errors"
)

type DevelopmentSMSProvider struct {
	code string
}

func NewDevelopmentSMSProvider(code string) *DevelopmentSMSProvider {
	return &DevelopmentSMSProvider{code: code}
}

func (p *DevelopmentSMSProvider) SendCode(_ context.Context, _, code string) error {
	if code != p.code {
		return errors.New("development SMS provider received unexpected code")
	}
	return nil
}

type DisabledSMSProvider struct{}

func (DisabledSMSProvider) SendCode(context.Context, string, string) error {
	return ErrSMSUnavailable
}
