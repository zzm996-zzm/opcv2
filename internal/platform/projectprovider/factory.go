package projectprovider

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

type Option func(*options)

type options struct {
	retrieval retrieval.Provider
}

func WithRetrievalProvider(provider retrieval.Provider) Option {
	return func(value *options) { value.retrieval = provider }
}

// New builds all project-market dependencies from the same startup config.
// Development implementations are intentionally available only outside production.
func New(cfg config.Config, optionValues ...Option) (Bundle, error) {
	settings := options{}
	for _, option := range optionValues {
		option(&settings)
	}
	if cfg.Environment == "production" && (cfg.ProjectFileProvider == "development" || cfg.ProjectRetrievalProvider == "development" || cfg.ProjectResearchProvider == "development") {
		return Bundle{}, fmt.Errorf("%w: development providers are not allowed in production", ErrUnsupportedProvider)
	}
	fileManager, err := newFiles(cfg)
	if err != nil {
		return Bundle{}, err
	}
	retriever, err := newRetrieval(cfg.ProjectRetrievalProvider, settings.retrieval)
	if err != nil {
		return Bundle{}, err
	}
	researchService, err := newResearch(cfg.ProjectResearchProvider, cfg)
	if err != nil {
		return Bundle{}, err
	}
	return Bundle{Files: fileManager, Retrieval: retriever, Research: researchService}, nil
}

func newFiles(cfg config.Config) (*files.Manager, error) {
	switch cfg.ProjectFileProvider {
	case "development", "local":
		storagePath := strings.TrimSpace(cfg.ProjectFileStoragePath)
		if storagePath == "" {
			storagePath = filepath.Join(os.TempDir(), "opcv2-project-match-files")
		}
		storage, err := files.NewLocalStorage(storagePath)
		if err != nil {
			return nil, err
		}
		return files.NewManager(storage, files.DevelopmentScanner{}, files.DevelopmentParser{}, files.DefaultMaxFileSize), nil
	default:
		return nil, ErrUnsupportedProvider
	}
}

func newRetrieval(provider string, configured retrieval.Provider) (retrieval.Provider, error) {
	switch provider {
	case "development":
		if configured != nil {
			return configured, nil
		}
		return retrieval.DevelopmentProvider{}, nil
	case "postgres":
		if configured == nil {
			return nil, fmt.Errorf("%w: postgres retrieval repository is required", ErrUnsupportedProvider)
		}
		return configured, nil
	default:
		return nil, ErrUnsupportedProvider
	}
}

func newResearch(provider string, cfg ...config.Config) (*research.Service, error) {
	var settings config.Config
	if len(cfg) > 0 {
		settings = cfg[0]
	}
	switch provider {
	case "development":
		dev := &research.DevelopmentProvider{Pages: map[string]research.Page{}}
		return research.NewService(dev, dev, dev, 0.5), nil
	case "serper":
		provider, err := research.NewSerperProvider(settings.SerperBaseURL, settings.SerperAPIKey, time.Duration(settings.SerperTimeoutSeconds)*time.Second)
		if err != nil {
			return nil, err
		}
		return research.NewService(provider, provider, provider, 0.5), nil
	default:
		return nil, ErrUnsupportedProvider
	}
}
