package account

import "time"

type Profile struct {
	ID                  int64     `json:"id"`
	Nickname            string    `json:"nickname"`
	Phone               string    `json:"phone"`
	Email               string    `json:"email"`
	Wechat              string    `json:"wechat"`
	Company             string    `json:"company"`
	Industry            string    `json:"industry"`
	Role                string    `json:"role"`
	OnboardingCompleted bool      `json:"onboarding_completed"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type Binding struct {
	Type        string `json:"type"`
	MaskedValue string `json:"masked_value"`
	Bound       bool   `json:"bound"`
}

type ProfilePayload struct {
	Profile  Profile   `json:"profile"`
	Bindings []Binding `json:"bindings"`
}

type ProfileUpdate struct {
	Nickname *string `json:"nickname"`
	Email    *string `json:"email"`
	Wechat   *string `json:"wechat"`
	Company  *string `json:"company"`
	Industry *string `json:"industry"`
	Role     *string `json:"role"`
}

type OnboardingSection struct {
	Key    string            `json:"key"`
	Title  string            `json:"title"`
	Fields map[string]string `json:"fields"`
}

type OnboardingState struct {
	Completed bool                `json:"completed"`
	Sections  []OnboardingSection `json:"sections"`
}

const (
	ProfileGroupIdentity    = "identity"
	ProfileGroupBusiness    = "business"
	ProfileGroupProducts    = "products"
	ProfileGroupResources   = "resources"
	ProfileGroupGoals       = "goals"
	ProfileGroupPreferences = "preferences"
)

type ProfileGroup struct {
	Key    string            `json:"key"`
	Title  string            `json:"title"`
	Fields map[string]string `json:"fields"`
}

type ProfileContext struct {
	UserID    int64          `json:"user_id"`
	Completed bool           `json:"completed"`
	Groups    []ProfileGroup `json:"groups"`
}

type Preferences struct {
	NotificationsEnabled bool   `json:"notifications_enabled"`
	DefaultModel         string `json:"default_model"`
	Language             string `json:"language"`
	Timezone             string `json:"timezone"`
}

type PreferencesUpdate struct {
	NotificationsEnabled *bool   `json:"notifications_enabled"`
	DefaultModel         *string `json:"default_model"`
	Language             *string `json:"language"`
	Timezone             *string `json:"timezone"`
}

type Quota struct {
	Key     string     `json:"key"`
	Label   string     `json:"label"`
	Used    int        `json:"used"`
	Limit   int        `json:"limit"`
	Unit    string     `json:"unit"`
	ResetAt *time.Time `json:"reset_at,omitempty"`
}

type ContentItem struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type DeletionStatus struct {
	Status string `json:"status"`
}
