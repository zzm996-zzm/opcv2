package competitor

import (
	"context"
	"encoding/json"
	"net/url"
	"time"
)

type DevelopmentScanner struct {
	now func() time.Time
}

func NewDevelopmentScanner() *DevelopmentScanner {
	return &DevelopmentScanner{now: time.Now}
}

func (s *DevelopmentScanner) Platform() string { return "" }

func (s *DevelopmentScanner) Scan(_ context.Context, scan Scan, _ *ScriptAccount) (ScanResult, error) {
	competitors := defaultCompetitors(scan.Targets)
	evidence := make([]EvidenceSource, 0, len(competitors))
	snapshots := make([]RawSnapshot, 0, len(competitors))
	capturedAt := s.now()
	for _, item := range competitors {
		evidence = append(evidence, EvidenceSource{
			SourceType: "development",
			Platform:   "development",
			Title:      item.Name + "公开信号样例",
			URL:        "https://example.com/competitor-scan?target=" + url.QueryEscape(item.Name),
			Summary:    item.Signal,
			CapturedAt: capturedAt,
		})
		payload, err := json.Marshal(map[string]any{
			"development_sample": true,
			"target":             item.Name,
			"signal":             item.Signal,
			"tags":               item.Tags,
		})
		if err != nil {
			return ScanResult{}, err
		}
		snapshots = append(snapshots, RawSnapshot{Platform: "development", Payload: payload, CapturedAt: capturedAt})
	}
	return ScanResult{
		Competitors:     competitors,
		Conclusions:     defaultConclusions(),
		EvidenceSources: evidence,
		RawSnapshots:    snapshots,
	}, nil
}
