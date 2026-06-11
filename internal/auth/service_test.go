package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeCodeStore struct {
	issuedPhone string
	issuedCode  string
	verifyErr   error
}

func (s *fakeCodeStore) Issue(_ context.Context, phone, code string, _, _ time.Duration) error {
	s.issuedPhone = phone
	s.issuedCode = code
	return nil
}

func (s *fakeCodeStore) Verify(_ context.Context, _, _ string) error {
	return s.verifyErr
}

type fakeSMSProvider struct {
	phone string
	code  string
	err   error
}

func (p *fakeSMSProvider) SendCode(_ context.Context, phone, code string) error {
	p.phone = phone
	p.code = code
	return p.err
}

type fakeUserRepository struct {
	user         User
	created      bool
	findErr      error
	recordedMeta LoginMeta
}

func (r *fakeUserRepository) FindOrCreateByPhone(_ context.Context, nickname, phone string, _ time.Time) (User, bool, error) {
	if r.findErr != nil {
		return User{}, false, r.findErr
	}
	if r.user.ID == 0 {
		r.user = User{ID: 42, Nickname: nickname, Phone: phone, Status: "active"}
	}
	return r.user, r.created, nil
}

func (r *fakeUserRepository) RecordLogin(_ context.Context, _ int64, meta LoginMeta) error {
	r.recordedMeta = meta
	return nil
}

func (r *fakeUserRepository) FindByID(_ context.Context, userID int64) (User, error) {
	if r.user.ID == userID {
		return r.user, nil
	}
	return User{}, ErrUserNotFound
}

type fakeTokenManager struct{}

func (fakeTokenManager) IssueAccess(user User) (string, time.Time, error) {
	return "access-token", time.Now().Add(15 * time.Minute), nil
}

func (fakeTokenManager) ParseAccess(string) (int64, error) {
	return 42, nil
}

type fakeSessionStore struct {
	userID int64
	token  string
}

func (s *fakeSessionStore) Create(_ context.Context, userID int64, _ time.Duration) (string, error) {
	s.userID = userID
	s.token = "refresh-token"
	return s.token, nil
}

func (s *fakeSessionStore) Consume(_ context.Context, token string) (int64, error) {
	if token != s.token {
		return 0, ErrInvalidRefreshToken
	}
	return s.userID, nil
}

func (s *fakeSessionStore) Delete(_ context.Context, _ string) error {
	return nil
}

func TestSendCodeValidatesPhoneAndSendsGeneratedCode(t *testing.T) {
	codeStore := &fakeCodeStore{}
	sms := &fakeSMSProvider{}
	service := NewService(Dependencies{
		Codes: codeStore,
		SMS:   sms,
		GenerateCode: func() (string, error) {
			return "135790", nil
		},
	})

	if err := service.SendCode(context.Background(), "13800138000"); err != nil {
		t.Fatalf("SendCode() error = %v", err)
	}
	if codeStore.issuedPhone != "13800138000" || codeStore.issuedCode != "135790" {
		t.Fatalf("issued code = %q/%q", codeStore.issuedPhone, codeStore.issuedCode)
	}
	if sms.phone != "13800138000" || sms.code != "135790" {
		t.Fatalf("SMS = %q/%q", sms.phone, sms.code)
	}
}

func TestSendCodeRejectsInvalidPhone(t *testing.T) {
	service := NewService(Dependencies{})

	if err := service.SendCode(context.Background(), "123"); !errors.Is(err, ErrInvalidPhone) {
		t.Fatalf("SendCode() error = %v, want ErrInvalidPhone", err)
	}
}

func TestLoginCreatesSessionAfterCodeVerification(t *testing.T) {
	users := &fakeUserRepository{created: true}
	sessions := &fakeSessionStore{}
	service := NewService(Dependencies{
		Codes:        &fakeCodeStore{},
		Users:        users,
		Tokens:       fakeTokenManager{},
		Sessions:     sessions,
		RefreshTTL:   7 * 24 * time.Hour,
		GenerateCode: func() (string, error) { return "000000", nil },
	})

	result, err := service.Login(context.Background(), LoginInput{
		Nickname:          "张晨",
		Phone:             "13800138000",
		Code:              "135790",
		AgreementAccepted: true,
		IP:                "127.0.0.1",
		UserAgent:         "test-agent",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.User.ID != 42 || result.AccessToken != "access-token" || result.RefreshToken != "refresh-token" {
		t.Fatalf("Login() result = %+v", result)
	}
	if users.recordedMeta.IP != "127.0.0.1" || users.recordedMeta.UserAgent != "test-agent" {
		t.Fatalf("login meta = %+v", users.recordedMeta)
	}
}

func TestLoginRequiresAgreement(t *testing.T) {
	service := NewService(Dependencies{})

	_, err := service.Login(context.Background(), LoginInput{
		Nickname: "张晨",
		Phone:    "13800138000",
		Code:     "135790",
	})
	if !errors.Is(err, ErrAgreementRequired) {
		t.Fatalf("Login() error = %v, want ErrAgreementRequired", err)
	}
}

func TestLoginReturnsInvalidCode(t *testing.T) {
	service := NewService(Dependencies{
		Codes: &fakeCodeStore{verifyErr: ErrInvalidCode},
	})

	_, err := service.Login(context.Background(), LoginInput{
		Nickname:          "张晨",
		Phone:             "13800138000",
		Code:              "000000",
		AgreementAccepted: true,
	})
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("Login() error = %v, want ErrInvalidCode", err)
	}
}

func TestRefreshRejectsDisabledUser(t *testing.T) {
	users := &fakeUserRepository{
		user: User{ID: 42, Phone: "13800138000", Status: "disabled"},
	}
	sessions := &fakeSessionStore{userID: 42, token: "refresh-token"}
	service := NewService(Dependencies{
		Users:    users,
		Tokens:   fakeTokenManager{},
		Sessions: sessions,
	})

	_, err := service.Refresh(context.Background(), "refresh-token")
	if !errors.Is(err, ErrUserDisabled) {
		t.Fatalf("Refresh() error = %v, want ErrUserDisabled", err)
	}
}
