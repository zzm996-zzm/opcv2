package competitor

import (
	"context"
	"net/url"
	"time"
)

type DevelopmentScanner struct {
	now func() time.Time
}

func NewDevelopmentScanner() *DevelopmentScanner {
	return &DevelopmentScanner{now: time.Now}
}

func (s *DevelopmentScanner) Scan(_ context.Context, scan Scan) (ScanResult, error) {
	competitors := defaultCompetitors(scan.Targets)
	evidence := make([]EvidenceSource, 0, len(competitors))
	capturedAt := s.now()
	for _, item := range competitors {
		evidence = append(evidence, EvidenceSource{
			SourceType: "development",
			Title:      item.Name + "公开信号样例",
			URL:        "https://example.com/competitor-scan?target=" + url.QueryEscape(item.Name),
			Summary:    item.Signal,
			CapturedAt: capturedAt,
		})
	}
	return ScanResult{
		Competitors:     competitors,
		Conclusions:     defaultConclusions(),
		EvidenceSources: evidence,
	}, nil
}
