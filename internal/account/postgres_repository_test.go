package account

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryGetsProfile(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT u.id, u.nickname, COALESCE(u.phone, ''), COALESCE(p.email, ''), COALESCE(u.wechat, ''),
		       COALESCE(p.company, ''), COALESCE(p.industry, ''), COALESCE(p.role, ''),
		       COALESCE(o.completed, false), u.created_at, COALESCE(p.updated_at, u.updated_at)
		FROM users u
		LEFT JOIN user_profiles p ON p.user_id = u.id
		LEFT JOIN user_onboarding o ON o.user_id = u.id
		WHERE u.id = $1 AND u.status = 'active'
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "nickname", "phone", "email", "wechat", "company", "industry", "role", "onboarding_completed", "created_at", "updated_at",
		}).AddRow(
			int64(42), "张晨", "13800138000", "founder@example.com", "zhangchen", "智活AI", "企业服务", "创始人", true, now, now,
		))

	repository := NewPostgresRepository(db)
	profile, err := repository.GetProfile(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if profile.Profile.ID != 42 || profile.Profile.Email != "founder@example.com" || len(profile.Bindings) != 3 {
		t.Fatalf("profile = %+v", profile)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositorySavesOnboardingJSON(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO user_onboarding (user_id, completed, sections, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			completed = EXCLUDED.completed,
			sections = EXCLUDED.sections,
			updated_at = NOW()
		RETURNING completed, sections
	`)).
		WithArgs(int64(42), false, []byte(`[{"key":"identity","title":"基本身份","fields":{"role":"创始人"}}]`)).
		WillReturnRows(pgxmock.NewRows([]string{"completed", "sections"}).AddRow(false, []byte(`[{"key":"identity","title":"基本身份","fields":{"role":"创始人"}}]`)))

	repository := NewPostgresRepository(db)
	state, err := repository.SaveOnboarding(context.Background(), 42, OnboardingState{
		Sections: []OnboardingSection{{Key: "identity", Title: "基本身份", Fields: map[string]string{"role": "创始人"}}},
	})
	if err != nil {
		t.Fatalf("SaveOnboarding() error = %v", err)
	}
	if len(state.Sections) != 1 || state.Sections[0].Fields["role"] != "创始人" {
		t.Fatalf("state = %+v", state)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryGetsDefaultPreferences(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO user_preferences (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
		RETURNING notifications_enabled, default_model, language, timezone
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"notifications_enabled", "default_model", "language", "timezone"}).
			AddRow(true, "", "zh-CN", "Asia/Shanghai"))

	repository := NewPostgresRepository(db)
	preferences, err := repository.GetPreferences(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetPreferences() error = %v", err)
	}
	if !preferences.NotificationsEnabled || preferences.Language != "zh-CN" {
		t.Fatalf("preferences = %+v", preferences)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
