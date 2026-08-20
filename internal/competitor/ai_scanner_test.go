package competitor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/ai"
)

type fakeAIGenerator struct {
	request ai.GenerateJSONRequest
	content []byte
	err     error
}

func (g *fakeAIGenerator) GenerateJSON(_ context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error) {
	g.request = request
	if g.err != nil {
		return ai.GenerateJSONResult{}, g.err
	}
	if err := request.Validate(g.content); err != nil {
		return ai.GenerateJSONResult{}, err
	}
	return ai.GenerateJSONResult{Content: g.content}, nil
}

func TestAIAnalysisScannerGeneratesStructuredResult(t *testing.T) {
	generator := &fakeAIGenerator{content: []byte(`{
		"competitors":[{"name":"小鹅通","category":"知识付费平台","score":86,"signal":"产品能力覆盖内容交付与私域运营。","risk":"high","tags":["内容交付","私域运营"]}],
		"conclusions":[{"title":"能力对比","detail":"优先比较课程交付和客户运营链路。"},{"title":"验证方向","detail":"通过客户访谈验证功能差异。"}]
	}`)}
	scanner := NewAIAnalysisScanner(generator)
	now := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
	scanner.now = func() time.Time { return now }

	result, err := scanner.Scan(context.Background(), Scan{ID: 99, UserID: 42, Targets: []string{"小鹅通"}, Focus: "产品能力"}, nil)

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if generator.request.Feature != "competitor.analysis" || generator.request.PromptVersion != "competitor_analysis_v1" || generator.request.UserID != 42 {
		t.Fatalf("request = %+v", generator.request)
	}
	if len(result.Competitors) != 1 || result.Competitors[0].Risk != "强" || len(result.Conclusions) != 2 {
		t.Fatalf("result = %+v", result)
	}
	if len(result.EvidenceSources) != 0 || len(result.RawSnapshots) != 1 || result.RawSnapshots[0].Platform != "ai" || !result.RawSnapshots[0].CapturedAt.Equal(now) {
		t.Fatalf("artifacts = %+v", result)
	}
}

func TestAIAnalysisScannerRejectsInvalidResult(t *testing.T) {
	generator := &fakeAIGenerator{content: []byte(`{"competitors":[],"conclusions":[]}`)}

	_, err := NewAIAnalysisScanner(generator).Scan(context.Background(), Scan{UserID: 42, Targets: []string{"小鹅通"}}, nil)

	if !errors.Is(err, ErrInvalidAIAnalysis) {
		t.Fatalf("Scan() error = %v, want ErrInvalidAIAnalysis", err)
	}
}

func TestAIAnalysisScannerRejectsUnexpectedCompetitorCount(t *testing.T) {
	generator := &fakeAIGenerator{content: []byte(`{
		"competitors":[
			{"name":"小鹅通","category":"知识付费平台","score":86,"signal":"能力重合。","risk":"中","tags":["产品"]},
			{"name":"额外对象","category":"其他","score":70,"signal":"额外生成。","risk":"弱","tags":["其他"]}
		],
		"conclusions":[{"title":"能力对比","detail":"比较核心能力。"},{"title":"验证方向","detail":"完成客户访谈。"}]
	}`)}

	_, err := NewAIAnalysisScanner(generator).Scan(context.Background(), Scan{UserID: 42, Targets: []string{"小鹅通"}}, nil)

	if !errors.Is(err, ErrInvalidAIAnalysis) {
		t.Fatalf("Scan() error = %v, want ErrInvalidAIAnalysis", err)
	}
}
