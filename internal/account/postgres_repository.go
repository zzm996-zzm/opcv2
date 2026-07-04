package account

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type postgresDB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetProfile(ctx context.Context, userID int64) (ProfilePayload, error) {
	const query = `
		SELECT u.id, u.nickname, COALESCE(u.phone, ''), COALESCE(p.email, ''), COALESCE(u.wechat, ''),
		       COALESCE(p.company, ''), COALESCE(p.industry, ''), COALESCE(p.role, ''),
		       COALESCE(o.completed, false), u.created_at, COALESCE(p.updated_at, u.updated_at)
		FROM users u
		LEFT JOIN user_profiles p ON p.user_id = u.id
		LEFT JOIN user_onboarding o ON o.user_id = u.id
		WHERE u.id = $1 AND u.status = 'active'
	`
	var profile Profile
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&profile.ID,
		&profile.Nickname,
		&profile.Phone,
		&profile.Email,
		&profile.Wechat,
		&profile.Company,
		&profile.Industry,
		&profile.Role,
		&profile.OnboardingCompleted,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProfilePayload{}, ErrProfileNotFound
	}
	if err != nil {
		return ProfilePayload{}, err
	}
	return ProfilePayload{Profile: profile, Bindings: bindingsFor(profile)}, nil
}

func (r *PostgresRepository) UpdateProfile(ctx context.Context, userID int64, update ProfileUpdate) (ProfilePayload, error) {
	if update.Nickname != nil || update.Wechat != nil {
		if _, err := r.db.Exec(ctx, `
			UPDATE users
			SET nickname = COALESCE($2, nickname),
			    wechat = COALESCE($3, wechat),
			    updated_at = NOW()
			WHERE id = $1 AND status = 'active'
		`, userID, update.Nickname, update.Wechat); err != nil {
			return ProfilePayload{}, err
		}
	}
	if _, err := r.db.Exec(ctx, `
		INSERT INTO user_profiles (user_id, email, company, industry, role, updated_at)
		VALUES ($1, COALESCE($2, ''), COALESCE($3, ''), COALESCE($4, ''), COALESCE($5, ''), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			email = COALESCE($2, user_profiles.email),
			company = COALESCE($3, user_profiles.company),
			industry = COALESCE($4, user_profiles.industry),
			role = COALESCE($5, user_profiles.role),
			updated_at = NOW()
	`, userID, update.Email, update.Company, update.Industry, update.Role); err != nil {
		return ProfilePayload{}, err
	}
	return r.GetProfile(ctx, userID)
}

func (r *PostgresRepository) GetOnboarding(ctx context.Context, userID int64) (OnboardingState, error) {
	const query = `
		INSERT INTO user_onboarding (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
		RETURNING completed, sections
	`
	var completed bool
	var raw []byte
	if err := r.db.QueryRow(ctx, query, userID).Scan(&completed, &raw); err != nil {
		return OnboardingState{}, err
	}
	return decodeOnboarding(completed, raw)
}

func (r *PostgresRepository) SaveOnboarding(ctx context.Context, userID int64, state OnboardingState) (OnboardingState, error) {
	raw, err := json.Marshal(httpapi.EnsureSlice(state.Sections))
	if err != nil {
		return OnboardingState{}, err
	}
	const query = `
		INSERT INTO user_onboarding (user_id, completed, sections, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			completed = EXCLUDED.completed,
			sections = EXCLUDED.sections,
			updated_at = NOW()
		RETURNING completed, sections
	`
	var completed bool
	var returned []byte
	if err := r.db.QueryRow(ctx, query, userID, state.Completed, raw).Scan(&completed, &returned); err != nil {
		return OnboardingState{}, err
	}
	return decodeOnboarding(completed, returned)
}

func (r *PostgresRepository) CompleteOnboarding(ctx context.Context, userID int64) (OnboardingState, error) {
	const query = `
		INSERT INTO user_onboarding (user_id, completed)
		VALUES ($1, true)
		ON CONFLICT (user_id) DO UPDATE SET completed = true, updated_at = NOW()
		RETURNING completed, sections
	`
	var completed bool
	var raw []byte
	if err := r.db.QueryRow(ctx, query, userID).Scan(&completed, &raw); err != nil {
		return OnboardingState{}, err
	}
	return decodeOnboarding(completed, raw)
}

func (r *PostgresRepository) GetPreferences(ctx context.Context, userID int64) (Preferences, error) {
	const query = `
		INSERT INTO user_preferences (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
		RETURNING notifications_enabled, default_model, language, timezone
	`
	var preferences Preferences
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&preferences.NotificationsEnabled,
		&preferences.DefaultModel,
		&preferences.Language,
		&preferences.Timezone,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return r.getExistingPreferences(ctx, userID)
	}
	return preferences, err
}

func (r *PostgresRepository) getExistingPreferences(ctx context.Context, userID int64) (Preferences, error) {
	const query = `
		SELECT notifications_enabled, default_model, language, timezone
		FROM user_preferences
		WHERE user_id = $1
	`
	var preferences Preferences
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&preferences.NotificationsEnabled,
		&preferences.DefaultModel,
		&preferences.Language,
		&preferences.Timezone,
	)
	return preferences, err
}

func (r *PostgresRepository) UpdatePreferences(ctx context.Context, userID int64, update PreferencesUpdate) (Preferences, error) {
	const query = `
		INSERT INTO user_preferences (user_id, notifications_enabled, default_model, language, timezone, updated_at)
		VALUES ($1, COALESCE($2, true), COALESCE($3, ''), COALESCE($4, 'zh-CN'), COALESCE($5, 'Asia/Shanghai'), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			notifications_enabled = COALESCE($2, user_preferences.notifications_enabled),
			default_model = COALESCE($3, user_preferences.default_model),
			language = COALESCE($4, user_preferences.language),
			timezone = COALESCE($5, user_preferences.timezone),
			updated_at = NOW()
		RETURNING notifications_enabled, default_model, language, timezone
	`
	var preferences Preferences
	err := r.db.QueryRow(ctx, query, userID, update.NotificationsEnabled, update.DefaultModel, update.Language, update.Timezone).Scan(
		&preferences.NotificationsEnabled,
		&preferences.DefaultModel,
		&preferences.Language,
		&preferences.Timezone,
	)
	return preferences, err
}

func (r *PostgresRepository) ListQuotas(_ context.Context, _ int64) ([]Quota, error) {
	return []Quota{}, nil
}

func (r *PostgresRepository) ListContent(_ context.Context, _ int64, _ int) ([]ContentItem, error) {
	return []ContentItem{}, nil
}

func (r *PostgresRepository) DeleteAccount(ctx context.Context, userID int64) (DeletionStatus, error) {
	_, err := r.db.Exec(ctx, `
		UPDATE users
		SET status = 'pending_deletion', updated_at = NOW()
		WHERE id = $1 AND status = 'active'
	`, userID)
	if err != nil {
		return DeletionStatus{}, err
	}
	return DeletionStatus{Status: "pending_deletion"}, nil
}

func decodeOnboarding(completed bool, raw []byte) (OnboardingState, error) {
	var sections []OnboardingSection
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &sections); err != nil {
			return OnboardingState{}, err
		}
	}
	return OnboardingState{Completed: completed, Sections: httpapi.EnsureSlice(sections)}, nil
}

func bindingsFor(profile Profile) []Binding {
	return []Binding{
		{Type: "phone", MaskedValue: maskValue(profile.Phone), Bound: strings.TrimSpace(profile.Phone) != ""},
		{Type: "email", MaskedValue: maskValue(profile.Email), Bound: strings.TrimSpace(profile.Email) != ""},
		{Type: "wechat", MaskedValue: maskValue(profile.Wechat), Bound: strings.TrimSpace(profile.Wechat) != ""},
	}
}

func maskValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= 4 {
		return value
	}
	return string(runes[:3]) + "****" + string(runes[len(runes)-4:])
}
