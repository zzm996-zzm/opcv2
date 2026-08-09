package projects

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"
)

var ErrInvalidDictionaryKind = errors.New("invalid project dictionary kind")

var dictionaryKinds = map[string]bool{
	"category":     true,
	"sector":       true,
	"product_type": true,
	"antipattern":  true,
	"country":      true,
	"cn_channel":   true,
}

type PublicConfig struct {
	FeaturePaywallEnabled bool `json:"feature_paywall_enabled"`
}

type DictionaryItem struct {
	Code   string `json:"code"`
	Kind   string `json:"kind"`
	NameZH string `json:"name_zh"`
	NameEN string `json:"name_en,omitempty"`
	Sort   int    `json:"sort"`
}

type ProjectFilters struct {
	Keyword    string
	Category   string
	Track      string
	Budget     string
	Difficulty string
	Resource   string
	Sort       string
	Featured   *bool
	Page       int
	PageSize   int
}

type Project struct {
	ID                   int64           `json:"id"`
	Slug                 string          `json:"slug"`
	Title                string          `json:"title"`
	CoverURL             string          `json:"cover_url,omitempty"`
	Category             string          `json:"category,omitempty"`
	Track                string          `json:"track,omitempty"`
	Difficulty           string          `json:"difficulty"`
	InvestCents          *int64          `json:"invest_cents,omitempty"`
	BudgetBand           string          `json:"budget_band,omitempty"`
	RevenueRange         string          `json:"revenue_range,omitempty"`
	IsReal               bool            `json:"is_real"`
	SourceURL            string          `json:"source_url,omitempty"`
	Summary              string          `json:"summary"`
	Tags                 []string        `json:"tags"`
	Heat                 int             `json:"heat"`
	IsFeatured           bool            `json:"is_featured"`
	IsFavorited          bool            `json:"is_favorited"`
	FitPct               *int            `json:"fit_pct,omitempty"`
	ResourceRequirements []string        `json:"resource_requirements,omitempty"`
	Detail               json.RawMessage `json:"detail,omitempty"`
	LockedBlocks         []string        `json:"locked_blocks"`
	IsUnlocked           bool            `json:"is_unlocked"`
	PublishedAt          *time.Time      `json:"published_at,omitempty"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

type ProjectPage struct {
	Items    []Project `json:"items"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Total    int       `json:"total"`
}

type ProjectHero struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Desc     string `json:"desc"`
	ImageURL string `json:"image_url,omitempty"`
}

type ProjectQuickTag struct {
	Code    string         `json:"code"`
	Name    string         `json:"name"`
	Filters map[string]any `json:"filters"`
}

type ProjectEntry struct {
	Code          string `json:"code"`
	Title         string `json:"title"`
	Desc          string `json:"desc"`
	ImageURL      string `json:"image_url,omitempty"`
	Route         string `json:"route"`
	IsRecommended bool   `json:"is_recommended"`
}

type ProjectHome struct {
	Hero      ProjectHero       `json:"hero"`
	QuickTags []ProjectQuickTag `json:"quick_tags"`
	Entries   []ProjectEntry    `json:"entries"`
	Featured  []Project         `json:"featured"`
}

type CatalogRepository interface {
	ListDictionaryItems(ctx context.Context, kind string) ([]DictionaryItem, error)
	ListProjects(ctx context.Context, filters ProjectFilters) (ProjectPage, error)
	GetProject(ctx context.Context, ref string) (Project, error)
}

type CatalogApplication interface {
	GetPublicConfig(context.Context) PublicConfig
	ListDictionaryItems(context.Context, string) ([]DictionaryItem, error)
	GetProjectHome(context.Context) (ProjectHome, error)
	ListProjects(context.Context, ProjectFilters) (ProjectPage, error)
	GetProject(context.Context, string) (Project, error)
}

func (s *Service) GetPublicConfig(context.Context) PublicConfig {
	return PublicConfig{FeaturePaywallEnabled: s.featurePaywallEnabled}
}

func (s *Service) ListDictionaryItems(ctx context.Context, kind string) ([]DictionaryItem, error) {
	kind = strings.TrimSpace(kind)
	if !dictionaryKinds[kind] {
		return nil, ErrInvalidDictionaryKind
	}
	repository, ok := s.repository.(CatalogRepository)
	if !ok {
		return nil, ErrServiceNotReady
	}
	items, err := repository.ListDictionaryItems(ctx, kind)
	if items == nil {
		items = []DictionaryItem{}
	}
	return items, err
}

func (s *Service) GetProjectHome(ctx context.Context) (ProjectHome, error) {
	featured := true
	page, err := s.ListProjects(ctx, ProjectFilters{Featured: &featured, Sort: "heat", Page: 1, PageSize: 8})
	if err != nil {
		return ProjectHome{}, err
	}
	return ProjectHome{
		Hero: ProjectHero{
			Title:    "项目超市",
			Subtitle: "发现下一个可落地机会",
			Desc:     "用历史案例看清坑，用当前证据找到能执行的机会",
		},
		QuickTags: defaultProjectQuickTags(),
		Entries: []ProjectEntry{
			{Code: "match", Title: "AI 匹配", Desc: "根据你的能力、预算和时间筛选项目", Route: "/projects/match", IsRecommended: true},
			{Code: "explore", Title: "机会探索", Desc: "浏览经过研究和核验的机会方向", Route: "/projects?tab=explore"},
			{Code: "cases", Title: "真实案例库", Desc: "从真实案例中查看结果和教训", Route: "/project-cases"},
		},
		Featured: page.Items,
	}, nil
}

func defaultProjectQuickTags() []ProjectQuickTag {
	return []ProjectQuickTag{
		{Code: "blue_ocean", Name: "蓝海机会", Filters: map[string]any{"group": "blue_ocean"}},
		{Code: "low_competition", Name: "低竞争赛道", Filters: map[string]any{"group": "low_competition"}},
		{Code: "low_budget", Name: "小成本启动", Filters: map[string]any{"group": "low_budget"}},
		{Code: "trending", Name: "近期爆发", Filters: map[string]any{"group": "trending"}},
		{Code: "solo", Name: "一人公司", Filters: map[string]any{"resource": "solo"}},
		{Code: "failure_lessons", Name: "失败教训", Filters: map[string]any{"group": "failure_lessons"}},
	}
}

func (s *Service) ListProjects(ctx context.Context, filters ProjectFilters) (ProjectPage, error) {
	repository, ok := s.repository.(CatalogRepository)
	if !ok {
		return ProjectPage{}, ErrServiceNotReady
	}
	filters.Keyword = strings.TrimSpace(filters.Keyword)
	filters.Category = strings.TrimSpace(filters.Category)
	filters.Track = strings.TrimSpace(filters.Track)
	filters.Budget = strings.TrimSpace(filters.Budget)
	filters.Difficulty = strings.TrimSpace(filters.Difficulty)
	filters.Resource = strings.TrimSpace(filters.Resource)
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 12
	}
	if filters.PageSize > 100 {
		filters.PageSize = 100
	}
	switch filters.Sort {
	case "latest", "heat":
	default:
		filters.Sort = "heat"
	}
	page, err := repository.ListProjects(ctx, filters)
	if err != nil {
		return ProjectPage{}, err
	}
	page.Page = filters.Page
	page.PageSize = filters.PageSize
	if page.Items == nil {
		page.Items = []Project{}
	}
	for index := range page.Items {
		s.normalizeProject(&page.Items[index])
	}
	return page, nil
}

func (s *Service) GetProject(ctx context.Context, ref string) (Project, error) {
	repository, ok := s.repository.(CatalogRepository)
	if !ok {
		return Project{}, ErrServiceNotReady
	}
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return Project{}, ErrOpportunityNotFound
	}
	project, err := repository.GetProject(ctx, ref)
	if err != nil {
		return Project{}, err
	}
	s.normalizeProject(&project)
	return project, nil
}

func (s *Service) normalizeProject(project *Project) {
	if project.Tags == nil {
		project.Tags = []string{}
	}
	if project.ResourceRequirements == nil {
		project.ResourceRequirements = []string{}
	}
	project.LockedBlocks = []string{}
	project.IsUnlocked = !s.featurePaywallEnabled
	project.SourceURL = eligibleProjectSourceURL(project.SourceURL)
	project.IsReal = project.IsReal && project.SourceURL != ""
	if len(project.Detail) == 0 {
		project.Detail = json.RawMessage(`{}`)
	}
}

func eligibleProjectSourceURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return ""
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "loot-drop.io" || strings.HasSuffix(host, ".loot-drop.io") {
		return ""
	}
	return parsed.String()
}
