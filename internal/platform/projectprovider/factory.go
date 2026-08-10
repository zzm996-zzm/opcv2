package projectprovider

import (
	"context"
	"errors"
	"fmt"

	"github.com/zzm/opcv2/internal/platform/config"
	"github.com/zzm/opcv2/internal/projects/files"
	"github.com/zzm/opcv2/internal/projects/research"
	"github.com/zzm/opcv2/internal/projects/retrieval"
)

var ErrUnsupportedProvider = errors.New("unsupported project provider")

type Bundle struct {
	Files     *files.Manager
	Retrieval retrieval.Provider
	Research  *research.Service
}

// New builds all project-market dependencies from the same startup config.
// Development implementations are intentionally available only outside production.
func New(cfg config.Config) (Bundle, error) {
	if cfg.Environment == "production" && (cfg.ProjectFileProvider == "development" || cfg.ProjectRetrievalProvider == "development" || cfg.ProjectResearchProvider == "development") {
		return Bundle{}, fmt.Errorf("%w: development providers are not allowed in production", ErrUnsupportedProvider)
	}
	fileManager, err := newFiles(cfg.ProjectFileProvider)
	if err != nil {
		return Bundle{}, err
	}
	retriever, err := newRetrieval(cfg.ProjectRetrievalProvider)
	if err != nil {
		return Bundle{}, err
	}
	researchService, err := newResearch(cfg.ProjectResearchProvider)
	if err != nil {
		return Bundle{}, err
	}
	return Bundle{Files: fileManager, Retrieval: retriever, Research: researchService}, nil
}

func newFiles(provider string) (*files.Manager, error) {
	switch provider {
	case "development":
		return files.NewManager(files.NewDevelopmentStorage(), files.DevelopmentScanner{}, files.DevelopmentParser{}, files.DefaultMaxFileSize), nil
	case "s3", "oss", "minio":
		return files.NewManager(unavailableStorage{}, unavailableScanner{}, unavailableParser{}, files.DefaultMaxFileSize), nil
	default:
		return nil, ErrUnsupportedProvider
	}
}

func newRetrieval(provider string) (retrieval.Provider, error) {
	switch provider {
	case "development":
		return retrieval.DevelopmentProvider{}, nil
	case "postgres", "pgvector", "qdrant":
		return unavailableRetrieval{}, nil
	default:
		return nil, ErrUnsupportedProvider
	}
}

func newResearch(provider string) (*research.Service, error) {
	switch provider {
	case "development":
		dev := &research.DevelopmentProvider{Pages: map[string]research.Page{}}
		return research.NewService(dev, dev, dev, 0.5), nil
	case "serper", "bing":
		dev := unavailableResearchProvider{}
		return research.NewService(dev, dev, dev, 0.5), nil
	default:
		return nil, ErrUnsupportedProvider
	}
}

type unavailableStorage struct{}

func (unavailableStorage) Put(context.Context, string, []byte) error { return ErrUnsupportedProvider }
func (unavailableStorage) Get(context.Context, string) ([]byte, error) {
	return nil, ErrUnsupportedProvider
}
func (unavailableStorage) Delete(context.Context, string) error { return ErrUnsupportedProvider }

type unavailableScanner struct{}

func (unavailableScanner) Scan(context.Context, []byte) error { return ErrUnsupportedProvider }

type unavailableParser struct{}

func (unavailableParser) Parse(context.Context, string, []byte) (files.ParsedDocument, error) {
	return files.ParsedDocument{}, ErrUnsupportedProvider
}

type unavailableRetrieval struct{}

func (unavailableRetrieval) Search(context.Context, retrieval.SearchRequest) ([]retrieval.Document, error) {
	return nil, ErrUnsupportedProvider
}

type unavailableResearchProvider struct{}

func (unavailableResearchProvider) Search(context.Context, string, int) ([]research.SearchResult, error) {
	return nil, ErrUnsupportedProvider
}
func (unavailableResearchProvider) Fetch(context.Context, string) (research.Page, error) {
	return research.Page{}, ErrUnsupportedProvider
}
func (unavailableResearchProvider) Extract(context.Context, research.Page, research.SearchResult) (research.Evidence, error) {
	return research.Evidence{}, ErrUnsupportedProvider
}
