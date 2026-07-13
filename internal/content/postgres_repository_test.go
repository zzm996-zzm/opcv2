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
	db.ExpectQuery(regexp.QuoteMeta("SELECT id, slug, title, summary, '' AS body, status,")).
		WithArgs("", "", 20).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "slug", "title", "summary", "body", "status", "source_name", "source_url", "author", "category", "tags", "citations", "source_published_at", "published_at", "created_at", "updated_at",
		}).AddRow(
			int64(7),
			"growth-playbook",
			"增长手册",
			"摘要",
			"",
			StatusPublished,
			"智活AI研究院",
			"https://example.com/report",
			"研究团队",
			"行业趋势",
			[]byte(`["AI创业"]`),
			[]byte(`[]`),
			now,
			now,
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	articles, err := repository.ListArticles(context.Background(), ArticleFilters{Limit: 20})
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
	db.ExpectQuery(regexp.QuoteMeta("INSERT INTO content_articles (")).
		WithArgs(
			"growth-playbook", "增长手册", "摘要", "正文", StatusPublished,
			"智活AI研究院", "https://example.com/report", "研究团队", "行业趋势",
			[]byte(`["AI创业"]`),
			[]byte(`[{"id":"source-1","label":"行业报告","source_name":"研究机构","source_url":"https://example.com/source"}]`),
			&now, now, now, now,
		).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "slug", "title", "summary", "body", "status", "source_name", "source_url", "author", "category", "tags", "citations", "source_published_at", "published_at", "created_at", "updated_at",
		}).AddRow(
			int64(7),
			"growth-playbook",
			"增长手册",
			"摘要",
			"正文",
			StatusPublished,
			"智活AI研究院",
			"https://example.com/report",
			"研究团队",
			"行业趋势",
			[]byte(`["AI创业"]`),
			[]byte(`[{"id":"source-1","label":"行业报告","source_name":"研究机构","source_url":"https://example.com/source"}]`),
			now,
			now,
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	article, err := repository.UpsertArticle(context.Background(), Article{
		Slug:       "growth-playbook",
		Title:      "增长手册",
		Summary:    "摘要",
		Body:       "正文",
		Status:     StatusPublished,
		SourceName: "智活AI研究院",
		SourceURL:  "https://example.com/report",
		Author:     "研究团队",
		Category:   "行业趋势",
		Tags:       []string{"AI创业"},
		Citations: []ArticleCitation{{
			ID: "source-1", Label: "行业报告", SourceName: "研究机构", SourceURL: "https://example.com/source",
		}},
		SourcePublishedAt: &now,
		PublishedAt:       now,
		CreatedAt:         now,
		UpdatedAt:         now,
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
