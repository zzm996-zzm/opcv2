package content

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryChecksAdminRole(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT role = 'admin'
		FROM users
		WHERE id = $1 AND status = 'active'
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"is_admin"}).AddRow(true))

	repository := NewPostgresRepository(db)
	isAdmin, err := repository.IsAdmin(context.Background(), 42)
	if err != nil {
		t.Fatalf("IsAdmin() error = %v", err)
	}
	if !isAdmin {
		t.Fatal("isAdmin = false, want true")
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsPublishedArticles(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, slug, title, summary, '' AS body, status, published_at, created_at, updated_at
		FROM content_articles
		WHERE status = 'published'
		ORDER BY published_at DESC NULLS LAST, created_at DESC
	`)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "slug", "title", "summary", "body", "status", "published_at", "created_at", "updated_at",
		}).AddRow(
			int64(7),
			"growth-playbook",
			"增长手册",
			"摘要",
			"",
			StatusPublished,
			now,
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	articles, err := repository.ListArticles(context.Background())
	if err != nil {
		t.Fatalf("ListArticles() error = %v", err)
	}
	if len(articles) != 1 || articles[0].Body != "" || articles[0].Slug != "growth-playbook" {
		t.Fatalf("articles = %+v", articles)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryUpsertsArticle(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO content_articles (slug, title, summary, body, status, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, TIMESTAMPTZ '0001-01-01 00:00:00+00'), $7, $8)
		ON CONFLICT (slug) DO UPDATE SET
			title = EXCLUDED.title,
			summary = EXCLUDED.summary,
			body = EXCLUDED.body,
			status = EXCLUDED.status,
			published_at = COALESCE(EXCLUDED.published_at, content_articles.published_at),
			updated_at = EXCLUDED.updated_at
		RETURNING id, slug, title, summary, body, status, published_at, created_at, updated_at
	`)).
		WithArgs("growth-playbook", "增长手册", "摘要", "正文", StatusPublished, now, now, now).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "slug", "title", "summary", "body", "status", "published_at", "created_at", "updated_at",
		}).AddRow(
			int64(7),
			"growth-playbook",
			"增长手册",
			"摘要",
			"正文",
			StatusPublished,
			now,
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	article, err := repository.UpsertArticle(context.Background(), Article{
		Slug:        "growth-playbook",
		Title:       "增长手册",
		Summary:     "摘要",
		Body:        "正文",
		Status:      StatusPublished,
		PublishedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("UpsertArticle() error = %v", err)
	}
	if article.ID != 7 || article.Body != "正文" {
		t.Fatalf("article = %+v", article)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
