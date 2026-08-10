package projectprovider

import (
	"context"
	"testing"

	"github.com/zzm/opcv2/internal/platform/config"
	"github.com/zzm/opcv2/internal/projects/files"
	"github.com/zzm/opcv2/internal/projects/retrieval"
)

func TestNewBuildsDevelopmentBundle(t *testing.T) {
	bundle, err := New(config.Config{ProjectFileProvider: "development", ProjectRetrievalProvider: "development", ProjectResearchProvider: "development"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if bundle.Files == nil || bundle.Retrieval == nil || bundle.Research == nil {
		t.Fatalf("bundle = %+v", bundle)
	}
	if _, ok := bundle.Retrieval.(retrieval.DevelopmentProvider); !ok {
		t.Fatalf("retrieval = %T", bundle.Retrieval)
	}
	if _, err := bundle.Files.Upload(context.Background(), 42, "input.txt", files.MIMEText, []byte("hello"), 0); err != nil {
		t.Fatalf("development files = %v", err)
	}
}

func TestNewRejectsUnimplementedProductionProviders(t *testing.T) {
	for _, test := range []struct {
		name string
		cfg  config.Config
	}{
		{name: "s3", cfg: config.Config{ProjectFileProvider: "s3", ProjectRetrievalProvider: "development", ProjectResearchProvider: "development"}},
		{name: "pgvector", cfg: config.Config{ProjectFileProvider: "development", ProjectRetrievalProvider: "pgvector", ProjectResearchProvider: "development"}},
		{name: "serper", cfg: config.Config{ProjectFileProvider: "development", ProjectRetrievalProvider: "development", ProjectResearchProvider: "serper"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := New(test.cfg); err == nil {
				t.Fatal("New() error = nil, want unsupported provider")
			}
		})
	}
}
