package main

import (
	"context"
	"testing"

	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/competitor"
	"github.com/zzm/opcv2/internal/platform/config"
)

type fakeCompetitorAIGenerator struct{}

func (fakeCompetitorAIGenerator) GenerateJSON(_ context.Context, _ ai.GenerateJSONRequest) (ai.GenerateJSONResult, error) {
	return ai.GenerateJSONResult{}, nil
}

func TestNewCompetitorScannerReturnsNilWhenProviderUnset(t *testing.T) {
	scanner, err := newCompetitorScanner(config.Config{}, nil)
	if err != nil {
		t.Fatalf("newCompetitorScanner() error = %v", err)
	}
	if scanner != nil {
		t.Fatalf("scanner = %T, want nil", scanner)
	}
}

func TestNewCompetitorScannerUsesDevelopmentProvider(t *testing.T) {
	scanner, err := newCompetitorScanner(config.Config{CompetitorScannerProvider: "development"}, nil)
	if err != nil {
		t.Fatalf("newCompetitorScanner() error = %v", err)
	}
	result, err := scanner.Scan(context.Background(), competitor.Scan{Targets: []string{"小鹅通"}, Focus: "价格变化"}, nil)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(result.Competitors) != 1 || len(result.EvidenceSources) != 1 {
		t.Fatalf("result = %+v", result)
	}
}

func TestNewCompetitorScannerRejectsUnsupportedProvider(t *testing.T) {
	if _, err := newCompetitorScanner(config.Config{CompetitorScannerProvider: "external"}, nil); err == nil {
		t.Fatal("newCompetitorScanner() error = nil, want unsupported provider error")
	}
}

func TestNewCompetitorScannerUsesAIProvider(t *testing.T) {
	scanner, err := newCompetitorScanner(config.Config{CompetitorScannerProvider: "ai"}, fakeCompetitorAIGenerator{})
	if err != nil {
		t.Fatalf("newCompetitorScanner() error = %v", err)
	}
	if _, ok := scanner.(*competitor.AIAnalysisScanner); !ok {
		t.Fatalf("scanner = %T, want *competitor.AIAnalysisScanner", scanner)
	}
}
