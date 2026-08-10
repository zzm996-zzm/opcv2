package projectprovider

import (
	"errors"

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
		return nil, ErrUnsupportedProvider
	default:
		return nil, ErrUnsupportedProvider
	}
}

func newRetrieval(provider string) (retrieval.Provider, error) {
	switch provider {
	case "development":
		return retrieval.DevelopmentProvider{}, nil
	case "postgres", "pgvector", "qdrant":
		return nil, ErrUnsupportedProvider
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
		return nil, ErrUnsupportedProvider
	default:
		return nil, ErrUnsupportedProvider
	}
}
