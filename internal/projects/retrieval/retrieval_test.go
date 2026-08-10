package retrieval

import (
	"context"
	"testing"
)

func TestDevelopmentRetrievalAndEmbedding(t *testing.T) {
	provider := DevelopmentProvider{Documents: []Document{{ID: "1", Title: "AI销售顾问", Text: "B端销售"}, {ID: "2", Title: "内容工作室", Text: "短视频"}}}
	items, err := provider.Search(context.Background(), SearchRequest{Query: "销售", Limit: 5})
	if err != nil || len(items) != 1 || items[0].ID != "1" {
		t.Fatalf("items/error = %+v/%v", items, err)
	}
	embedding := DevelopmentEmbedding{Dimension: 3}
	vectors, err := embedding.Embed(context.Background(), []string{"abc"})
	if err != nil || len(vectors) != 1 || len(vectors[0]) != 3 {
		t.Fatalf("vectors/error = %+v/%v", vectors, err)
	}
}

func TestCanonicalURLNormalizesFragmentAndTrailingSlash(t *testing.T) {
	canonical, err := CanonicalURL("HTTPS://Example.COM/path/#section")
	if err != nil || canonical != "https://example.com/path" {
		t.Fatalf("canonical/error = %q/%v", canonical, err)
	}
}
