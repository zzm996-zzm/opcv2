package account

import (
	"context"
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrUserIDRequired      = errors.New("user id required")
	ErrInvalidEmail        = errors.New("invalid email")
	ErrInvalidOnboarding   = errors.New("invalid onboarding")
	ErrProfileNotFound     = errors.New("profile not found")
	ErrServiceNotReady     = errors.New("account service is not configured")
	ErrPasswordUnsupported = errors.New("password update is not configured")
)

type Repository interface {
	GetProfile(ctx context.Context, userID int64) (ProfilePayload, error)
	UpdateProfile(ctx context.Context, userID int64, update ProfileUpdate) (ProfilePayload, error)
	GetOnboarding(ctx context.Context, userID int64) (OnboardingState, error)
	SaveOnboarding(ctx context.Context, userID int64, state OnboardingState) (OnboardingState, error)
	CompleteOnboarding(ctx context.Context, userID int64) (OnboardingState, error)
	GetPreferences(ctx context.Context, userID int64) (Preferences, error)
	UpdatePreferences(ctx context.Context, userID int64, update PreferencesUpdate) (Preferences, error)
	ListQuotas(ctx context.Context, userID int64) ([]Quota, error)
	ListContent(ctx context.Context, userID int64, limit int) ([]ContentItem, error)
	DeleteAccount(ctx context.Context, userID int64) (DeletionStatus, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func DefaultPreferences() Preferences {
	return Preferences{
		NotificationsEnabled: true,
		Language:             "zh-CN",
		Timezone:             "Asia/Shanghai",
	}
}

func (s *Service) GetProfile(ctx context.Context, userID int64) (ProfilePayload, error) {
	if err := s.ready(userID); err != nil {
		return ProfilePayload{}, err
	}
	return s.repository.GetProfile(ctx, userID)
}

func (s *Service) UpdateProfile(ctx context.Context, userID int64, update ProfileUpdate) (ProfilePayload, error) {
	if err := s.ready(userID); err != nil {
		return ProfilePayload{}, err
	}
	trimString(update.Nickname)
	trimString(update.Email)
	trimString(update.Wechat)
	trimString(update.Company)
	trimString(update.Industry)
	trimString(update.Role)
	if update.Email != nil && *update.Email != "" {
		if _, err := mail.ParseAddress(*update.Email); err != nil {
			return ProfilePayload{}, ErrInvalidEmail
		}
	}
	return s.repository.UpdateProfile(ctx, userID, update)
}

func (s *Service) GetOnboarding(ctx context.Context, userID int64) (OnboardingState, error) {
	if err := s.ready(userID); err != nil {
		return OnboardingState{}, err
	}
	return s.repository.GetOnboarding(ctx, userID)
}

func (s *Service) GetProfileContext(ctx context.Context, userID int64) (ProfileContext, error) {
	if err := s.ready(userID); err != nil {
		return ProfileContext{}, err
	}
	profile, err := s.repository.GetProfile(ctx, userID)
	if err != nil {
		return ProfileContext{}, err
	}
	onboarding, err := s.repository.GetOnboarding(ctx, userID)
	if err != nil {
		return ProfileContext{}, err
	}
	return buildProfileContext(profile.Profile, onboarding), nil
}

func (s *Service) SaveOnboarding(ctx context.Context, userID int64, state OnboardingState) (OnboardingState, error) {
	if err := s.ready(userID); err != nil {
		return OnboardingState{}, err
	}
	for index := range state.Sections {
		state.Sections[index].Key = strings.TrimSpace(state.Sections[index].Key)
		state.Sections[index].Title = strings.TrimSpace(state.Sections[index].Title)
		if state.Sections[index].Key == "" {
			return OnboardingState{}, ErrInvalidOnboarding
		}
		if state.Sections[index].Fields == nil {
			state.Sections[index].Fields = map[string]string{}
		}
		for key, value := range state.Sections[index].Fields {
			trimmedKey := strings.TrimSpace(key)
			if trimmedKey == "" {
				return OnboardingState{}, ErrInvalidOnboarding
			}
			delete(state.Sections[index].Fields, key)
			state.Sections[index].Fields[trimmedKey] = strings.TrimSpace(value)
		}
	}
	return s.repository.SaveOnboarding(ctx, userID, state)
}

func (s *Service) CompleteOnboarding(ctx context.Context, userID int64) (OnboardingState, error) {
	if err := s.ready(userID); err != nil {
		return OnboardingState{}, err
	}
	return s.repository.CompleteOnboarding(ctx, userID)
}

func (s *Service) GetPreferences(ctx context.Context, userID int64) (Preferences, error) {
	if err := s.ready(userID); err != nil {
		return Preferences{}, err
	}
	return s.repository.GetPreferences(ctx, userID)
}

func (s *Service) UpdatePreferences(ctx context.Context, userID int64, update PreferencesUpdate) (Preferences, error) {
	if err := s.ready(userID); err != nil {
		return Preferences{}, err
	}
	trimString(update.DefaultModel)
	trimString(update.Language)
	trimString(update.Timezone)
	return s.repository.UpdatePreferences(ctx, userID, update)
}

func (s *Service) ListQuotas(ctx context.Context, userID int64) ([]Quota, error) {
	if err := s.ready(userID); err != nil {
		return nil, err
	}
	return s.repository.ListQuotas(ctx, userID)
}

func (s *Service) ListContent(ctx context.Context, userID int64, limit int) ([]ContentItem, error) {
	if err := s.ready(userID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repository.ListContent(ctx, userID, limit)
}

func (s *Service) DeleteAccount(ctx context.Context, userID int64) (DeletionStatus, error) {
	if err := s.ready(userID); err != nil {
		return DeletionStatus{}, err
	}
	return s.repository.DeleteAccount(ctx, userID)
}

func (s *Service) ready(userID int64) error {
	if userID <= 0 {
		return ErrUserIDRequired
	}
	if s.repository == nil {
		return ErrServiceNotReady
	}
	return nil
}

func trimString(value *string) {
	if value != nil {
		*value = strings.TrimSpace(*value)
	}
}

func buildProfileContext(profile Profile, onboarding OnboardingState) ProfileContext {
	groups := defaultProfileGroups(profile)
	byKey := make(map[string]int, len(groups))
	for index, group := range groups {
		byKey[group.Key] = index
	}
	for _, section := range onboarding.Sections {
		key := strings.TrimSpace(section.Key)
		index, ok := byKey[key]
		if !ok {
			continue
		}
		if groups[index].Fields == nil {
			groups[index].Fields = map[string]string{}
		}
		for field, value := range section.Fields {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			groups[index].Fields[field] = strings.TrimSpace(value)
		}
	}
	return ProfileContext{
		UserID:    profile.ID,
		Completed: onboarding.Completed,
		Groups:    groups,
	}
}

func defaultProfileGroups(profile Profile) []ProfileGroup {
	return []ProfileGroup{
		{
			Key:   ProfileGroupIdentity,
			Title: "基本身份",
			Fields: map[string]string{
				"nickname": profile.Nickname,
				"role":     profile.Role,
				"industry": profile.Industry,
			},
		},
		{
			Key:   ProfileGroupBusiness,
			Title: "我的业务/公司",
			Fields: map[string]string{
				"company":  profile.Company,
				"industry": profile.Industry,
			},
		},
		{Key: ProfileGroupProducts, Title: "我的产品", Fields: map[string]string{}},
		{Key: ProfileGroupResources, Title: "能力与资源", Fields: map[string]string{}},
		{Key: ProfileGroupGoals, Title: "目标与诉求", Fields: map[string]string{}},
		{Key: ProfileGroupPreferences, Title: "偏好", Fields: map[string]string{}},
	}
}
