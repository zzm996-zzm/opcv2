package auth

import (
	"context"
	"errors"
	"testing"
)

func TestDevelopmentSMSProviderAcceptsConfiguredCode(t *testing.T) {
	provider := NewDevelopmentSMSProvider("246810")

	if err := provider.SendCode(context.Background(), "13800138000", "246810"); err != nil {
		t.Fatalf("SendCode() error = %v", err)
	}
}

func TestDevelopmentSMSProviderRejectsUnexpectedCode(t *testing.T) {
	provider := NewDevelopmentSMSProvider("246810")

	if err := provider.SendCode(context.Background(), "13800138000", "123456"); err == nil {
		t.Fatal("SendCode() error = nil, want unexpected code error")
	}
}

func TestDisabledSMSProviderRejectsSending(t *testing.T) {
	provider := DisabledSMSProvider{}

	if err := provider.SendCode(context.Background(), "13800138000", "246810"); !errors.Is(err, ErrSMSUnavailable) {
		t.Fatalf("SendCode() error = %v, want ErrSMSUnavailable", err)
	}
}
