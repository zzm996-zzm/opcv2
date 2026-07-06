package competitor

import (
	"context"
	"testing"
)

func TestDevelopmentScannerReturnsCompetitorsConclusionsAndEvidence(t *testing.T) {
	scanner := NewDevelopmentScanner()

	result, err := scanner.Scan(context.Background(), Scan{
		ID:      99,
		UserID:  42,
		Targets: []string{"小鹅通", "有赞教育"},
		Focus:   "价格变化",
	})

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(result.Competitors) != 2 || result.Competitors[0].Name != "小鹅通" {
		t.Fatalf("competitors = %+v", result.Competitors)
	}
	if len(result.Conclusions) == 0 || result.Conclusions[0].Title == "" {
		t.Fatalf("conclusions = %+v", result.Conclusions)
	}
	if len(result.EvidenceSources) != 2 {
		t.Fatalf("evidence = %+v, want one source per target", result.EvidenceSources)
	}
	if result.EvidenceSources[0].Title == "" || result.EvidenceSources[0].URL == "" || result.EvidenceSources[0].CapturedAt.IsZero() {
		t.Fatalf("first evidence = %+v", result.EvidenceSources[0])
	}
}
