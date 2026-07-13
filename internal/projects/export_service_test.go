package projects

import (
	"context"
	"testing"
)

func TestServiceCreatesExportFromOwnedMatchSnapshot(t *testing.T) {
	repository := &memoryRepository{sessions: []MatchSession{{ID: 99, UserID: 42, Intent: "AI项目", Status: StatusCompleted, Result: MatchResult{Status: StatusCompleted}}}}
	service := NewService(repository, nil)
	export, err := service.CreateExport(context.Background(), CreateExportInput{UserID: 42, SourceType: ExportSourceMatch, SourceID: 99})
	if err != nil || export.ID == 0 || len(export.Payload) == 0 {
		t.Fatalf("export = %+v err=%v", export, err)
	}
}
