package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPhone        = errors.New("invalid phone")
	ErrInvalidCode         = errors.New("invalid verification code")
	ErrCodeRateLimited     = errors.New("verification code rate limited")
	ErrAgreementRequired   = errors.New("agreement acceptance required")
	ErrNicknameRequired    = errors.New("nickname required")
	ErrInvalidAccount      = errors.New("invalid account")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrAccountExists       = errors.New("account already exists")
	ErrSMSUnavailable      = errors.New("SMS service is unavailable")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserDisabled        = errors.New("user disabled")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrAuthNotConfigured   = errors.New("auth service is not configured")
)

var mainlandPhonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)
var accountPattern = regexp.MustCompile(`^[A-Za-z0-9_]{4,32}$`)

type User struct {
	ID        int64     `json:"id"`
	Nickname  string    `json:"nickname"`
	Phone     string    `json:"phone"`
	Account   string    `json:"account,omitempty"`
	Wechat    string    `json:"wechat,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type UserCredentials struct {
	User         User
	PasswordHash string
}

type LoginMeta struct {
	IP        string
	UserAgent string
}

type LoginInput struct {
	Nickname          string
	Phone             string
	Code              string
	Account           string
	Password          string
	AgreementAccepted bool
	IP                string
	UserAgent         string
}

type RegisterInput struct {
	Nickname          string
	Account           string
	Password          string
	AgreementAccepted bool
	IP                string
	UserAgent         string
}

type LoginResult struct {
	User               User      `json:"user"`
	AccessToken        string    `json:"access_token"`
	AccessTokenExpires time.Time `json:"access_token_expires_at"`
	RefreshToken       string    `json:"refresh_token"`
	IsNewUser          bool      `json:"is_new_user"`
}

type CodeStore interface {
	Issue(ctx context.Context, phone, code string, ttl, cooldown time.Duration) error
	Verify(ctx context.Context, phone, code string) error
}

type SMSProvider interface {
	SendCode(ctx context.Context, phone, code string) error
}

type UserRepository interface {
	FindOrCreateByPhone(ctx context.Context, nickname, phone string, agreementAcceptedAt time.Time) (User, bool, error)
	RegisterAccount(ctx context.Context, nickname, account, passwordHash string, agreementAcceptedAt time.Time) (User, error)
	FindCredentialsByAccount(ctx context.Context, account string) (UserCredentials, error)
	RecordLogin(ctx context.Context, userID int64, meta LoginMeta) error
	FindByID(ctx context.Context, userID int64) (User, error)
}

type TokenManager interface {
	IssueAccess(user User) (token string, expiresAt time.Time, err error)
	ParseAccess(token string) (userID int64, err error)
}

type SessionStore interface {
	Create(ctx context.Context, userID int64, ttl time.Duration) (string, error)
	Consume(ctx context.Context, token string) (userID int64, err error)
	Delete(ctx context.Context, token string) error
}

type Dependencies struct {
	Codes        CodeStore
	SMS          SMSProvider
	Users        UserRepository
	Tokens       TokenManager
	Sessions     SessionStore
	GenerateCode func() (string, error)
	CodeTTL      time.Duration
	CodeCooldown time.Duration
	RefreshTTL   time.Duration
}

type Service struct {
	deps Dependencies
}

func NewService(deps Dependencies) *Service {
	if deps.GenerateCode == nil {
		deps.GenerateCode = generateSixDigitCode
	}
	if deps.CodeTTL <= 0 {
		deps.CodeTTL = 5 * time.Minute
	}
	if deps.CodeCooldown <= 0 {
		deps.CodeCooldown = 60 * time.Second
	}
	if deps.RefreshTTL <= 0 {
		deps.RefreshTTL = 7 * 24 * time.Hour
	}
	return &Service{deps: deps}
}

func (s *Service) SendCode(ctx context.Context, rawPhone string) error {
	phone := strings.TrimSpace(rawPhone)
	if !mainlandPhonePattern.MatchString(phone) {
		return ErrInvalidPhone
	}
	if s.deps.Codes == nil || s.deps.SMS == nil {
		return ErrAuthNotConfigured
	}
	code, err := s.deps.GenerateCode()
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}
	if err := s.deps.Codes.Issue(ctx, phone, code, s.deps.CodeTTL, s.deps.CodeCooldown); err != nil {
		return err
	}
	if err := s.deps.SMS.SendCode(ctx, phone, code); err != nil {
		return fmt.Errorf("send SMS code: %w", err)
	}
	return nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	input.Nickname = strings.TrimSpace(input.Nickname)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Code = strings.TrimSpace(input.Code)
	input.Account = normalizeAccount(input.Account)
	input.Password = strings.TrimSpace(input.Password)

	if input.Account != "" || input.Password != "" {
		return s.loginWithPassword(ctx, input)
	}

	if !input.AgreementAccepted {
		return LoginResult{}, ErrAgreementRequired
	}
	if input.Nickname == "" {
		return LoginResult{}, ErrNicknameRequired
	}
	if !mainlandPhonePattern.MatchString(input.Phone) {
		return LoginResult{}, ErrInvalidPhone
	}
	if s.deps.Codes == nil {
		return LoginResult{}, ErrAuthNotConfigured
	}
	if err := s.deps.Codes.Verify(ctx, input.Phone, input.Code); err != nil {
		return LoginResult{}, err
	}
	if s.deps.Users == nil || s.deps.Tokens == nil || s.deps.Sessions == nil {
		return LoginResult{}, ErrAuthNotConfigured
	}

	user, created, err := s.deps.Users.FindOrCreateByPhone(ctx, input.Nickname, input.Phone, time.Now())
	if err != nil {
		return LoginResult{}, fmt.Errorf("find or create user: %w", err)
	}
	if user.Status != "active" {
		return LoginResult{}, ErrUserDisabled
	}
	if err := s.deps.Users.RecordLogin(ctx, user.ID, LoginMeta{IP: input.IP, UserAgent: input.UserAgent}); err != nil {
		return LoginResult{}, fmt.Errorf("record login: %w", err)
	}
	accessToken, expiresAt, err := s.deps.Tokens.IssueAccess(user)
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue access token: %w", err)
	}
	refreshToken, err := s.deps.Sessions.Create(ctx, user.ID, s.deps.RefreshTTL)
	if err != nil {
		return LoginResult{}, fmt.Errorf("create refresh session: %w", err)
	}
	return LoginResult{
		User:               user,
		AccessToken:        accessToken,
		AccessTokenExpires: expiresAt,
		RefreshToken:       refreshToken,
		IsNewUser:          created,
	}, nil
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (LoginResult, error) {
	input.Nickname = strings.TrimSpace(input.Nickname)
	input.Account = normalizeAccount(input.Account)
	input.Password = strings.TrimSpace(input.Password)

	if !input.AgreementAccepted {
		return LoginResult{}, ErrAgreementRequired
	}
	if !accountPattern.MatchString(input.Account) {
		return LoginResult{}, ErrInvalidAccount
	}
	if err := validatePassword(input.Password); err != nil {
		return LoginResult{}, err
	}
	if input.Nickname == "" {
		input.Nickname = input.Account
	}
	if s.deps.Users == nil || s.deps.Tokens == nil || s.deps.Sessions == nil {
		return LoginResult{}, ErrAuthNotConfigured
	}
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return LoginResult{}, fmt.Errorf("hash password: %w", err)
	}
	user, err := s.deps.Users.RegisterAccount(ctx, input.Nickname, input.Account, passwordHash, time.Now())
	if err != nil {
		if errors.Is(err, ErrAccountExists) {
			// A lost response can leave the account persisted. Repeating the same
			// registration is safe only when the supplied password proves ownership.
			credentials, lookupErr := s.deps.Users.FindCredentialsByAccount(ctx, input.Account)
			if lookupErr == nil && credentials.User.Status == "active" && verifyPassword(credentials.PasswordHash, input.Password) == nil {
				result, issueErr := s.issueLoginResult(ctx, credentials.User, LoginMeta{IP: input.IP, UserAgent: input.UserAgent})
				if issueErr != nil {
					return LoginResult{}, issueErr
				}
				return result, nil
			}
			return LoginResult{}, ErrAccountExists
		}
		return LoginResult{}, fmt.Errorf("register account: %w", err)
	}
	if user.Status != "active" {
		return LoginResult{}, ErrUserDisabled
	}
	result, err := s.issueLoginResult(ctx, user, LoginMeta{IP: input.IP, UserAgent: input.UserAgent})
	if err != nil {
		return LoginResult{}, err
	}
	result.IsNewUser = true
	return result, nil
}

func (s *Service) loginWithPassword(ctx context.Context, input LoginInput) (LoginResult, error) {
	if !accountPattern.MatchString(input.Account) {
		return LoginResult{}, ErrInvalidCredentials
	}
	if strings.TrimSpace(input.Password) == "" {
		return LoginResult{}, ErrInvalidCredentials
	}
	if s.deps.Users == nil || s.deps.Tokens == nil || s.deps.Sessions == nil {
		return LoginResult{}, ErrAuthNotConfigured
	}
	credentials, err := s.deps.Users.FindCredentialsByAccount(ctx, input.Account)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}
	if err := verifyPassword(credentials.PasswordHash, input.Password); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}
	if credentials.User.Status != "active" {
		return LoginResult{}, ErrUserDisabled
	}
	return s.issueLoginResult(ctx, credentials.User, LoginMeta{IP: input.IP, UserAgent: input.UserAgent})
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (LoginResult, error) {
	if s.deps.Sessions == nil || s.deps.Users == nil || s.deps.Tokens == nil {
		return LoginResult{}, ErrAuthNotConfigured
	}
	userID, err := s.deps.Sessions.Consume(ctx, strings.TrimSpace(refreshToken))
	if err != nil {
		return LoginResult{}, err
	}
	user, err := s.deps.Users.FindByID(ctx, userID)
	if err != nil {
		return LoginResult{}, err
	}
	if user.Status != "active" {
		return LoginResult{}, ErrUserDisabled
	}
	accessToken, expiresAt, err := s.deps.Tokens.IssueAccess(user)
	if err != nil {
		return LoginResult{}, err
	}
	nextRefreshToken, err := s.deps.Sessions.Create(ctx, user.ID, s.deps.RefreshTTL)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		User:               user,
		AccessToken:        accessToken,
		AccessTokenExpires: expiresAt,
		RefreshToken:       nextRefreshToken,
	}, nil
}

func (s *Service) issueLoginResult(ctx context.Context, user User, meta LoginMeta) (LoginResult, error) {
	if err := s.deps.Users.RecordLogin(ctx, user.ID, meta); err != nil {
		return LoginResult{}, fmt.Errorf("record login: %w", err)
	}
	accessToken, expiresAt, err := s.deps.Tokens.IssueAccess(user)
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue access token: %w", err)
	}
	refreshToken, err := s.deps.Sessions.Create(ctx, user.ID, s.deps.RefreshTTL)
	if err != nil {
		return LoginResult{}, fmt.Errorf("create refresh session: %w", err)
	}
	return LoginResult{
		User:               user,
		AccessToken:        accessToken,
		AccessTokenExpires: expiresAt,
		RefreshToken:       refreshToken,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if s.deps.Sessions == nil {
		return nil
	}
	return s.deps.Sessions.Delete(ctx, strings.TrimSpace(refreshToken))
}

func (s *Service) CurrentUser(ctx context.Context, userID int64) (User, error) {
	if s.deps.Users == nil {
		return User{}, ErrAuthNotConfigured
	}
	user, err := s.deps.Users.FindByID(ctx, userID)
	if err != nil {
		return User{}, err
	}
	if user.Status != "active" {
		return User{}, ErrUserDisabled
	}
	return user, nil
}

func generateSixDigitCode() (string, error) {
	max := big.NewInt(1_000_000)
	number, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", number.Int64()), nil
}

func normalizeAccount(account string) string {
	return strings.ToLower(strings.TrimSpace(account))
}

func validatePassword(password string) error {
	if len(password) < 6 || len(password) > 72 {
		return ErrInvalidPassword
	}
	return nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func verifyPassword(passwordHash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
}
