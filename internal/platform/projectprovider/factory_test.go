package projectprovider

import (
	"context"
	"errors"
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

func TestNewRejectsDevelopmentProvidersInProduction(t *testing.T) {
	for _, field := range []string{"file", "retrieval", "research"} {
		t.Run(field, func(t *testing.T) {
			cfg := config.Config{Environment: "production", ProjectFileProvider: "s3", ProjectRetrievalProvider: "pgvector", ProjectResearchProvider: "serper"}
			switch field {
			case "file":
				cfg.ProjectFileProvider = "development"
			case "retrieval":
				cfg.ProjectRetrievalProvider = "development"
			case "research":
				cfg.ProjectResearchProvider = "development"
			}
			if _, err := New(cfg); err == nil {
				t.Fatal("New() error = nil, want production development provider error")
			}
		})
	}
}

func TestNewBuildsUnavailableAdaptersForNamedProductionProviders(t *testing.T) {
	_, err := New(config.Config{Environment: "production", ProjectFileProvider: "s3", ProjectRetrievalProvider: "pgvector", ProjectResearchProvider: "bing"})
	if !errors.Is(err, ErrUnsupportedProvider) {
		t.Fatalf("New() error = %v, want unsupported provider", err)
	}
}

func TestNewBuildsConfiguredProductionAdapters(t *testing.T) {
	retriever := retrieval.DevelopmentProvider{}
	bundle, err := New(config.Config{Environment: "production", ProjectFileProvider: "local", ProjectFileStoragePath: t.TempDir(), ProjectRetrievalProvider: "postgres", ProjectResearchProvider: "serper", SerperAPIKey: "key", SerperTimeoutSeconds: 1}, WithRetrievalProvider(retriever))
	if err != nil || bundle.Files == nil || bundle.Retrieval == nil || bundle.Research == nil {
		t.Fatalf("New() = %+v, %v", bundle, err)
	}
}
