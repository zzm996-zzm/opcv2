package content

import (
	"context"
	"testing"
	"time"
)

type fakeRepository struct {
	admin    bool
	article  Article
	articles []Article
	err      error
}

func (r *fakeRepository) IsAdmin(context.Context, int64) (bool, error) {
	return r.admin, r.err
}

func (r *fakeRepository) ListArticles(context.Context) ([]Article, error) {
	return r.articles, r.err
}

func (r *fakeRepository) GetArticle(context.Context, string) (Article, error) {
	return r.article, r.err
}

func (r *fakeRepository) UpsertArticle(_ context.Context, article Article) (Article, error) {
	r.article = article
	return article, r.err
}

func (r *fakeRepository) ListTools(context.Context) ([]Tool, error) {
	return nil, r.err
}

func (r *fakeRepository) UpsertTool(context.Context, Tool) (Tool, error) {
	return Tool{}, r.err
}

func (r *fakeRepository) GetCommunityConfig(context.Context) (CommunityConfig, error) {
	return CommunityConfig{}, r.err
}

func (r *fakeRepository) UpsertCommunityConfig(context.Context, CommunityConfig) (CommunityConfig, error) {
	return CommunityConfig{}, r.err
}

func (r *fakeRepository) ListBrandMetrics(context.Context) ([]BrandMetric, error) {
	return nil, r.err
}

func (r *fakeRepository) UpsertBrandMetric(context.Context, BrandMetric) (BrandMetric, error) {
	return BrandMetric{}, r.err
}

func (r *fakeRepository) ListBrandCases(context.Context) ([]BrandCase, error) {
	return nil, r.err
}

func (r *fakeRepository) UpsertBrandCase(context.Context, BrandCase) (BrandCase, error) {
	return BrandCase{}, r.err
}

func TestServiceRejectsNonAdminArticleWrite(t *testing.T) {
	service := NewService(&fakeRepository{admin: false})

	_, err := service.CreateArticle(context.Background(), 42, ArticleInput{
		Slug:  "growth-playbook",
		Title: "增长手册",
		Body:  "正文",
	})

	if err != ErrAdminRequired {
		t.Fatalf("err = %v, want ErrAdminRequired", err)
	}
}

func TestServiceNormalizesArticleAndAllowsAdmin(t *testing.T) {
	repository := &fakeRepository{admin: true}
	service := NewService(repository)
	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	article, err := service.CreateArticle(context.Background(), 42, ArticleInput{
		Slug:   " growth-playbook ",
		Title:  " 增长手册 ",
		Body:   " 正文 ",
		Status: StatusPublished,
	})

	if err != nil {
		t.Fatalf("CreateArticle() error = %v", err)
	}
	if article.Slug != "growth-playbook" || article.Title != "增长手册" || !article.PublishedAt.Equal(now) {
		t.Fatalf("article = %+v", article)
	}
}

func TestServiceEmptyListsAreArrays(t *testing.T) {
	service := NewService(&fakeRepository{})

	articles, err := service.ListArticles(context.Background())

	if err != nil {
		t.Fatalf("ListArticles() error = %v", err)
	}
	if articles == nil || len(articles) != 0 {
		t.Fatalf("articles = %#v, want empty slice", articles)
	}
}
