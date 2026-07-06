package account

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	profile     ProfilePayload
	update      ProfileUpdate
	onboarding  OnboardingState
	preferences Preferences
	content     []ContentItem
	deletion    DeletionStatus
	err         error
	userID      int64
	limit       int
}

func (r *fakeRepository) GetProfile(_ context.Context, userID int64) (ProfilePayload, error) {
	r.userID = userID
	return r.profile, r.err
}

func (r *fakeRepository) UpdateProfile(_ context.Context, userID int64, update ProfileUpdate) (ProfilePayload, error) {
	r.userID = userID
	r.update = update
	return r.profile, r.err
}

func (r *fakeRepository) GetOnboarding(_ context.Context, userID int64) (OnboardingState, error) {
	r.userID = userID
	return r.onboarding, r.err
}

func (r *fakeRepository) SaveOnboarding(_ context.Context, userID int64, state OnboardingState) (OnboardingState, error) {
	r.userID = userID
	r.onboarding = state
	return state, r.err
}

func (r *fakeRepository) CompleteOnboarding(_ context.Context, userID int64) (OnboardingState, error) {
	r.userID = userID
	r.onboarding.Completed = true
	return r.onboarding, r.err
}

func (r *fakeRepository) GetPreferences(_ context.Context, userID int64) (Preferences, error) {
	r.userID = userID
	return r.preferences, r.err
}

func (r *fakeRepository) UpdatePreferences(_ context.Context, userID int64, update PreferencesUpdate) (Preferences, error) {
	r.userID = userID
	if update.NotificationsEnabled != nil {
		r.preferences.NotificationsEnabled = *update.NotificationsEnabled
	}
	if update.DefaultModel != nil {
		r.preferences.DefaultModel = *update.DefaultModel
	}
	if update.Language != nil {
		r.preferences.Language = *update.Language
	}
	if update.Timezone != nil {
		r.preferences.Timezone = *update.Timezone
	}
	return r.preferences, r.err
}

func (r *fakeRepository) ListQuotas(_ context.Context, userID int64) ([]Quota, error) {
	r.userID = userID
	return nil, r.err
}

func (r *fakeRepository) ListContent(_ context.Context, userID int64, limit int) ([]ContentItem, error) {
	r.userID = userID
	r.limit = limit
	return r.content, r.err
}

func (r *fakeRepository) DeleteAccount(_ context.Context, userID int64) (DeletionStatus, error) {
	r.userID = userID
	return r.deletion, r.err
}

func TestServiceRejectsMissingUserID(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.GetProfile(context.Background(), 0)

	if !errors.Is(err, ErrUserIDRequired) {
		t.Fatalf("err = %v, want ErrUserIDRequired", err)
	}
}

func TestServiceTrimsProfileUpdateAndRejectsInvalidEmail(t *testing.T) {
	nickname := "  张晨  "
	email := "not-an-email"
	service := NewService(&fakeRepository{})

	_, err := service.UpdateProfile(context.Background(), 42, ProfileUpdate{
		Nickname: &nickname,
		Email:    &email,
	})

	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("err = %v, want ErrInvalidEmail", err)
	}

	email = " founder@example.com "
	repository := &fakeRepository{profile: ProfilePayload{Profile: Profile{ID: 42, Nickname: "张晨"}}}
	service = NewService(repository)

	_, err = service.UpdateProfile(context.Background(), 42, ProfileUpdate{
		Nickname: &nickname,
		Email:    &email,
	})

	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if repository.update.Nickname == nil || *repository.update.Nickname != "张晨" {
		t.Fatalf("nickname update = %+v", repository.update.Nickname)
	}
	if repository.update.Email == nil || *repository.update.Email != "founder@example.com" {
		t.Fatalf("email update = %+v", repository.update.Email)
	}
}

func TestServiceCompletesOnboarding(t *testing.T) {
	repository := &fakeRepository{onboarding: OnboardingState{
		Sections: []OnboardingSection{{Key: "identity", Title: "基本身份", Fields: map[string]string{"role": "创始人"}}},
	}}
	service := NewService(repository)

	state, err := service.CompleteOnboarding(context.Background(), 42)

	if err != nil {
		t.Fatalf("CompleteOnboarding() error = %v", err)
	}
	if !state.Completed || repository.userID != 42 {
		t.Fatalf("state/userID = %+v/%d", state, repository.userID)
	}
}

func TestServiceBuildsProfileContextWithSixGroups(t *testing.T) {
	repository := &fakeRepository{
		profile: ProfilePayload{Profile: Profile{
			ID:       42,
			Nickname: "张晨",
			Company:  "智活AI",
			Industry: "企业服务",
			Role:     "创始人",
		}},
		onboarding: OnboardingState{
			Completed: true,
			Sections: []OnboardingSection{
				{Key: "business", Title: "我的业务/公司", Fields: map[string]string{"stage": "启动", "channels": "私域"}},
				{Key: "goals", Title: "目标与诉求", Fields: map[string]string{"short_term": "验证项目"}},
			},
		},
	}
	service := NewService(repository)

	context, err := service.GetProfileContext(context.Background(), 42)

	if err != nil {
		t.Fatalf("GetProfileContext() error = %v", err)
	}
	if context.UserID != 42 || !context.Completed {
		t.Fatalf("context = %+v", context)
	}
	if len(context.Groups) != 6 {
		t.Fatalf("groups = %+v, want 6 groups", context.Groups)
	}
	identity := context.Groups[0]
	if identity.Key != ProfileGroupIdentity || identity.Fields["nickname"] != "张晨" || identity.Fields["industry"] != "企业服务" {
		t.Fatalf("identity = %+v", identity)
	}
	business := context.Groups[1]
	if business.Key != ProfileGroupBusiness || business.Fields["company"] != "智活AI" || business.Fields["stage"] != "启动" {
		t.Fatalf("business = %+v", business)
	}
	if context.Groups[2].Key != ProfileGroupProducts || context.Groups[3].Key != ProfileGroupResources ||
		context.Groups[4].Key != ProfileGroupGoals || context.Groups[5].Key != ProfileGroupPreferences {
		t.Fatalf("group order = %+v", context.Groups)
	}
}

func TestServiceReturnsDefaultPreferences(t *testing.T) {
	service := NewService(&fakeRepository{preferences: DefaultPreferences()})

	preferences, err := service.GetPreferences(context.Background(), 42)

	if err != nil {
		t.Fatalf("GetPreferences() error = %v", err)
	}
	if !preferences.NotificationsEnabled || preferences.Language != "zh-CN" || preferences.Timezone != "Asia/Shanghai" {
		t.Fatalf("preferences = %+v", preferences)
	}
}

func TestServiceListsContentWithCappedLimit(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	repository := &fakeRepository{content: []ContentItem{{ID: "analysis:99", Type: "analysis", Title: "报告", CreatedAt: now}}}
	service := NewService(repository)

	items, err := service.ListContent(context.Background(), 42, 500)

	if err != nil {
		t.Fatalf("ListContent() error = %v", err)
	}
	if repository.limit != 100 || len(items) != 1 {
		t.Fatalf("limit/items = %d/%+v", repository.limit, items)
	}
}

func TestServiceDeleteAccountReturnsPendingDeletion(t *testing.T) {
	repository := &fakeRepository{deletion: DeletionStatus{Status: "pending_deletion"}}
	service := NewService(repository)

	status, err := service.DeleteAccount(context.Background(), 42)

	if err != nil {
		t.Fatalf("DeleteAccount() error = %v", err)
	}
	if status.Status != "pending_deletion" || repository.userID != 42 {
		t.Fatalf("status/userID = %+v/%d", status, repository.userID)
	}
}
