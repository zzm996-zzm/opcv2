package content

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type postgresDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const query = `
		SELECT role = 'admin'
		FROM users
		WHERE id = $1 AND status = 'active'
	`
	var isAdmin bool
	err := r.db.QueryRow(ctx, query, userID).Scan(&isAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return isAdmin, err
}

func (r *PostgresRepository) ListArticles(ctx context.Context, filters ArticleFilters) ([]Article, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, title, summary, '' AS body, status,
		       source_name, source_url, author, category, tags, citations,
		       source_published_at, published_at, created_at, updated_at
		FROM content_articles
		WHERE status = 'published'
		  AND ($1 = '' OR category = $1)
		  AND ($2 = '' OR title ILIKE '%' || $2 || '%' OR summary ILIKE '%' || $2 || '%' OR tags::TEXT ILIKE '%' || $2 || '%')
		ORDER BY published_at DESC NULLS LAST, created_at DESC
		LIMIT $3
	`, filters.Category, filters.Query, filters.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		article, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}
	return articles, rows.Err()
}

func (r *PostgresRepository) GetArticle(ctx context.Context, slug string) (Article, error) {
	article, err := scanArticle(r.db.QueryRow(ctx, `
		SELECT id, slug, title, summary, body, status,
		       source_name, source_url, author, category, tags, citations,
		       source_published_at, published_at, created_at, updated_at
		FROM content_articles
		WHERE slug = $1 AND status = 'published'
	`, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return Article{}, ErrArticleNotFound
	}
	return article, err
}

func (r *PostgresRepository) UpsertArticle(ctx context.Context, article Article) (Article, error) {
	tags, err := json.Marshal(article.Tags)
	if err != nil {
		return Article{}, err
	}
	citations, err := json.Marshal(article.Citations)
	if err != nil {
		return Article{}, err
	}
	return scanArticle(r.db.QueryRow(ctx, `
		INSERT INTO content_articles (
			slug, title, summary, body, status, source_name, source_url, author, category,
			tags, citations, source_published_at, published_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
		        NULLIF($13, TIMESTAMPTZ '0001-01-01 00:00:00+00'), $14, $15)
		ON CONFLICT (slug) DO UPDATE SET
			title = EXCLUDED.title,
			summary = EXCLUDED.summary,
			body = EXCLUDED.body,
			status = EXCLUDED.status,
			source_name = EXCLUDED.source_name,
			source_url = EXCLUDED.source_url,
			author = EXCLUDED.author,
			category = EXCLUDED.category,
			tags = EXCLUDED.tags,
			citations = EXCLUDED.citations,
			source_published_at = EXCLUDED.source_published_at,
			published_at = COALESCE(EXCLUDED.published_at, content_articles.published_at),
			updated_at = EXCLUDED.updated_at
		RETURNING id, slug, title, summary, body, status,
		          source_name, source_url, author, category, tags, citations,
		          source_published_at, published_at, created_at, updated_at
	`,
		article.Slug,
		article.Title,
		article.Summary,
		article.Body,
		article.Status,
		article.SourceName,
		article.SourceURL,
		article.Author,
		article.Category,
		tags,
		citations,
		article.SourcePublishedAt,
		article.PublishedAt,
		article.CreatedAt,
		article.UpdatedAt,
	))
}

func (r *PostgresRepository) BookmarkArticle(ctx context.Context, userID int64, slug string) (BookmarkResult, error) {
	_, err := r.db.Exec(ctx, `
		INSERT INTO content_article_bookmarks (user_id, article_slug)
		VALUES ($1, $2)
		ON CONFLICT (user_id, article_slug) DO NOTHING
	`, userID, slug)
	if err != nil {
		return BookmarkResult{}, err
	}
	return BookmarkResult{Slug: slug, Bookmarked: true}, nil
}

func (r *PostgresRepository) UnbookmarkArticle(ctx context.Context, userID int64, slug string) (BookmarkResult, error) {
	_, err := r.db.Exec(ctx, `
		DELETE FROM content_article_bookmarks
		WHERE user_id = $1 AND article_slug = $2
	`, userID, slug)
	if err != nil {
		return BookmarkResult{}, err
	}
	return BookmarkResult{Slug: slug, Bookmarked: false}, nil
}

func (r *PostgresRepository) ListTools(ctx context.Context, filters ToolFilters) ([]Tool, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, name, description, url, status, COALESCE(category, ''), tags,
		       provider_name, price_label, platforms, features, use_cases, limitations,
		       source_url, source_updated_at, sort_weight, created_at, updated_at
		FROM content_tools
		WHERE status = 'published'
		  AND ($1 = '' OR category = $1)
		  AND ($2 = '' OR name ILIKE '%' || $2 || '%' OR description ILIKE '%' || $2 || '%'
		       OR tags::TEXT ILIKE '%' || $2 || '%' OR features::TEXT ILIKE '%' || $2 || '%'
		       OR use_cases::TEXT ILIKE '%' || $2 || '%')
		ORDER BY
		  CASE WHEN $3 = 'hot' THEN sort_weight ELSE 0 END DESC,
		  created_at DESC
		LIMIT $4
	`, filters.Category, filters.Query, filters.Sort, filters.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []Tool
	for rows.Next() {
		tool, err := scanTool(rows)
		if err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}
	return tools, rows.Err()
}

func (r *PostgresRepository) GetTool(ctx context.Context, slug string) (Tool, error) {
	tool, err := scanTool(r.db.QueryRow(ctx, `
		SELECT id, slug, name, description, url, status, COALESCE(category, ''), tags,
		       provider_name, price_label, platforms, features, use_cases, limitations,
		       source_url, source_updated_at, sort_weight, created_at, updated_at
		FROM content_tools
		WHERE slug = $1 AND status = 'published'
	`, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return Tool{}, ErrToolNotFound
	}
	return tool, err
}

func (r *PostgresRepository) FavoriteTool(ctx context.Context, userID int64, slug string) (FavoriteResult, error) {
	_, err := r.db.Exec(ctx, `
		INSERT INTO content_tool_favorites (user_id, tool_slug)
		VALUES ($1, $2)
		ON CONFLICT (user_id, tool_slug) DO NOTHING
	`, userID, slug)
	if err != nil {
		return FavoriteResult{}, err
	}
	return FavoriteResult{Slug: slug, Favorited: true}, nil
}

func (r *PostgresRepository) UnfavoriteTool(ctx context.Context, userID int64, slug string) (FavoriteResult, error) {
	_, err := r.db.Exec(ctx, `
		DELETE FROM content_tool_favorites
		WHERE user_id = $1 AND tool_slug = $2
	`, userID, slug)
	if err != nil {
		return FavoriteResult{}, err
	}
	return FavoriteResult{Slug: slug, Favorited: false}, nil
}

func (r *PostgresRepository) UpsertTool(ctx context.Context, tool Tool) (Tool, error) {
	tags, err := json.Marshal(tool.Tags)
	if err != nil {
		return Tool{}, err
	}
	platforms, err := json.Marshal(tool.Platforms)
	if err != nil {
		return Tool{}, err
	}
	features, err := json.Marshal(tool.Features)
	if err != nil {
		return Tool{}, err
	}
	useCases, err := json.Marshal(tool.UseCases)
	if err != nil {
		return Tool{}, err
	}
	limitations, err := json.Marshal(tool.Limitations)
	if err != nil {
		return Tool{}, err
	}
	return scanTool(r.db.QueryRow(ctx, `
		INSERT INTO content_tools (
			slug, name, description, url, status, category, tags, provider_name, price_label,
			platforms, features, use_cases, limitations, source_url, source_updated_at,
			sort_weight, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			url = EXCLUDED.url,
			status = EXCLUDED.status,
			category = EXCLUDED.category,
			tags = EXCLUDED.tags,
			provider_name = EXCLUDED.provider_name,
			price_label = EXCLUDED.price_label,
			platforms = EXCLUDED.platforms,
			features = EXCLUDED.features,
			use_cases = EXCLUDED.use_cases,
			limitations = EXCLUDED.limitations,
			source_url = EXCLUDED.source_url,
			source_updated_at = EXCLUDED.source_updated_at,
			sort_weight = EXCLUDED.sort_weight,
			updated_at = EXCLUDED.updated_at
		RETURNING id, slug, name, description, url, status, COALESCE(category, ''), tags,
		          provider_name, price_label, platforms, features, use_cases, limitations,
		          source_url, source_updated_at, sort_weight, created_at, updated_at
	`,
		tool.Slug,
		tool.Name,
		tool.Description,
		tool.URL,
		tool.Status,
		tool.Category,
		tags,
		tool.ProviderName,
		tool.PriceLabel,
		platforms,
		features,
		useCases,
		limitations,
		tool.SourceURL,
		tool.SourceUpdatedAt,
		tool.SortWeight,
		tool.CreatedAt,
		tool.UpdatedAt,
	))
}

func (r *PostgresRepository) GetCommunityConfig(ctx context.Context) (CommunityConfig, error) {
	config, err := scanCommunityConfig(r.db.QueryRow(ctx, `
		SELECT id, headline, description, join_url, qr_variants, created_at, updated_at
		FROM community_config
		WHERE id = 1
	`))
	if errors.Is(err, pgx.ErrNoRows) {
		return CommunityConfig{}, nil
	}
	return config, err
}

func (r *PostgresRepository) UpsertCommunityConfig(ctx context.Context, config CommunityConfig) (CommunityConfig, error) {
	variants, err := json.Marshal(config.QRVariants)
	if err != nil {
		return CommunityConfig{}, err
	}
	return scanCommunityConfig(r.db.QueryRow(ctx, `
		INSERT INTO community_config (id, headline, description, join_url, qr_variants, created_at, updated_at)
		VALUES (1, $1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			headline = EXCLUDED.headline,
			description = EXCLUDED.description,
			join_url = EXCLUDED.join_url,
			qr_variants = EXCLUDED.qr_variants,
			updated_at = EXCLUDED.updated_at
		RETURNING id, headline, description, join_url, qr_variants, created_at, updated_at
	`,
		config.Headline,
		config.Description,
		config.JoinURL,
		variants,
		config.CreatedAt,
		config.UpdatedAt,
	))
}

func (r *PostgresRepository) CreateCommunityJoinRequest(ctx context.Context, userID int64, input CommunityJoinInput) (CommunityJoinRequest, error) {
	var request CommunityJoinRequest
	err := r.db.QueryRow(ctx, `
		INSERT INTO community_join_requests (user_id, community, contact, note)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, community, contact, note, status, created_at
	`, userID, input.Community, input.Contact, input.Note).Scan(
		&request.ID,
		&request.UserID,
		&request.Community,
		&request.Contact,
		&request.Note,
		&request.Status,
		&request.CreatedAt,
	)
	return request, err
}

func (r *PostgresRepository) ListHelpTopics(ctx context.Context) ([]HelpTopic, error) {
	rows, err := r.db.Query(ctx, `
		SELECT key, name
		FROM help_topics
		ORDER BY display_order, key
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var topics []HelpTopic
	for rows.Next() {
		var topic HelpTopic
		if err := rows.Scan(&topic.Key, &topic.Name); err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}
	return topics, rows.Err()
}

func (r *PostgresRepository) ListHelpArticles(ctx context.Context, filters HelpArticleFilters) ([]HelpArticle, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, topic, title, summary, '' AS body, created_at, updated_at
		FROM help_articles
		WHERE status = 'published'
		  AND ($1 = '' OR topic = $1)
		  AND ($2 = '' OR title ILIKE '%' || $2 || '%' OR summary ILIKE '%' || $2 || '%')
		ORDER BY created_at DESC
		LIMIT $3
	`, filters.Topic, filters.Query, filters.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var articles []HelpArticle
	for rows.Next() {
		article, err := scanHelpArticle(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}
	return articles, rows.Err()
}

func (r *PostgresRepository) GetHelpArticle(ctx context.Context, slug string) (HelpArticle, error) {
	article, err := scanHelpArticle(r.db.QueryRow(ctx, `
		SELECT id, slug, topic, title, summary, body, created_at, updated_at
		FROM help_articles
		WHERE slug = $1 AND status = 'published'
	`, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return HelpArticle{}, ErrHelpArticleNotFound
	}
	return article, err
}

func (r *PostgresRepository) ListBrandMetrics(ctx context.Context) ([]BrandMetric, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, key, label, value, created_at, updated_at
		FROM brand_metrics
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []BrandMetric
	for rows.Next() {
		metric, err := scanBrandMetric(rows)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, rows.Err()
}

func (r *PostgresRepository) UpsertBrandMetric(ctx context.Context, metric BrandMetric) (BrandMetric, error) {
	return scanBrandMetric(r.db.QueryRow(ctx, `
		INSERT INTO brand_metrics (key, label, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (key) DO UPDATE SET
			label = EXCLUDED.label,
			value = EXCLUDED.value,
			updated_at = EXCLUDED.updated_at
		RETURNING id, key, label, value, created_at, updated_at
	`,
		metric.Key,
		metric.Label,
		metric.Value,
		metric.CreatedAt,
		metric.UpdatedAt,
	))
}

func (r *PostgresRepository) ListBrandCases(ctx context.Context) ([]BrandCase, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, title, summary, url, created_at, updated_at
		FROM brand_cases
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []BrandCase
	for rows.Next() {
		brandCase, err := scanBrandCase(rows)
		if err != nil {
			return nil, err
		}
		cases = append(cases, brandCase)
	}
	return cases, rows.Err()
}

func (r *PostgresRepository) UpsertBrandCase(ctx context.Context, brandCase BrandCase) (BrandCase, error) {
	return scanBrandCase(r.db.QueryRow(ctx, `
		INSERT INTO brand_cases (slug, title, summary, url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (slug) DO UPDATE SET
			title = EXCLUDED.title,
			summary = EXCLUDED.summary,
			url = EXCLUDED.url,
			updated_at = EXCLUDED.updated_at
		RETURNING id, slug, title, summary, url, created_at, updated_at
	`,
		brandCase.Slug,
		brandCase.Title,
		brandCase.Summary,
		brandCase.URL,
		brandCase.CreatedAt,
		brandCase.UpdatedAt,
	))
}

type scanner interface {
	Scan(dest ...any) error
}

func scanArticle(scanner scanner) (Article, error) {
	var article Article
	var publishedAt sql.NullTime
	var sourcePublishedAt sql.NullTime
	var tags, citations []byte
	if err := scanner.Scan(
		&article.ID,
		&article.Slug,
		&article.Title,
		&article.Summary,
		&article.Body,
		&article.Status,
		&article.SourceName,
		&article.SourceURL,
		&article.Author,
		&article.Category,
		&tags,
		&citations,
		&sourcePublishedAt,
		&publishedAt,
		&article.CreatedAt,
		&article.UpdatedAt,
	); err != nil {
		return Article{}, err
	}
	if publishedAt.Valid {
		article.PublishedAt = publishedAt.Time
	}
	if sourcePublishedAt.Valid {
		article.SourcePublishedAt = &sourcePublishedAt.Time
	}
	if err := json.Unmarshal(tags, &article.Tags); err != nil {
		return Article{}, err
	}
	if err := json.Unmarshal(citations, &article.Citations); err != nil {
		return Article{}, err
	}
	return article, nil
}

func scanTool(scanner scanner) (Tool, error) {
	var tool Tool
	var tags, platforms, features, useCases, limitations []byte
	var sourceUpdatedAt sql.NullTime
	err := scanner.Scan(
		&tool.ID,
		&tool.Slug,
		&tool.Name,
		&tool.Description,
		&tool.URL,
		&tool.Status,
		&tool.Category,
		&tags,
		&tool.ProviderName,
		&tool.PriceLabel,
		&platforms,
		&features,
		&useCases,
		&limitations,
		&tool.SourceURL,
		&sourceUpdatedAt,
		&tool.SortWeight,
		&tool.CreatedAt,
		&tool.UpdatedAt,
	)
	if err != nil {
		return Tool{}, err
	}
	if sourceUpdatedAt.Valid {
		tool.SourceUpdatedAt = &sourceUpdatedAt.Time
	}
	for data, destination := range map[*[]byte]*[]string{
		&tags: &tool.Tags, &platforms: &tool.Platforms, &features: &tool.Features,
		&useCases: &tool.UseCases, &limitations: &tool.Limitations,
	} {
		if err := json.Unmarshal(*data, destination); err != nil {
			return Tool{}, err
		}
	}
	return tool, nil
}

func scanHelpArticle(scanner scanner) (HelpArticle, error) {
	var article HelpArticle
	err := scanner.Scan(
		&article.ID,
		&article.Slug,
		&article.Topic,
		&article.Title,
		&article.Summary,
		&article.Body,
		&article.CreatedAt,
		&article.UpdatedAt,
	)
	return article, err
}

func scanCommunityConfig(scanner scanner) (CommunityConfig, error) {
	var config CommunityConfig
	var variants []byte
	err := scanner.Scan(
		&config.ID,
		&config.Headline,
		&config.Description,
		&config.JoinURL,
		&variants,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err != nil {
		return CommunityConfig{}, err
	}
	if err := json.Unmarshal(variants, &config.QRVariants); err != nil {
		return CommunityConfig{}, err
	}
	return config, nil
}

func scanBrandMetric(scanner scanner) (BrandMetric, error) {
	var metric BrandMetric
	err := scanner.Scan(
		&metric.ID,
		&metric.Key,
		&metric.Label,
		&metric.Value,
		&metric.CreatedAt,
		&metric.UpdatedAt,
	)
	return metric, err
}

func scanBrandCase(scanner scanner) (BrandCase, error) {
	var brandCase BrandCase
	err := scanner.Scan(
		&brandCase.ID,
		&brandCase.Slug,
		&brandCase.Title,
		&brandCase.Summary,
		&brandCase.URL,
		&brandCase.CreatedAt,
		&brandCase.UpdatedAt,
	)
	return brandCase, err
}
