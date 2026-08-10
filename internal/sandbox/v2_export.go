package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type V2ExportRepository interface {
	V2Repository
	CreateV2Export(ctx context.Context, export V2Export, userID int64, payload []byte) (V2Export, error)
	GetV2Export(ctx context.Context, userID, exportID int64) (V2Export, []byte, error)
}

func (s *Service) CreateV2Export(ctx context.Context, input CreateV2ExportInput) (V2Export, error) {
	repository, ok := s.repository.(V2ExportRepository)
	if !ok || repository == nil {
		return V2Export{}, ErrServiceNotReady
	}
	format := strings.ToLower(strings.TrimSpace(input.Format))
	if format != "json" && format != "pdf" && format != "link" {
		return V2Export{}, ErrV2InvalidRequest
	}
	run, err := repository.GetV2Run(ctx, input.UserID, input.RunID)
	if err != nil {
		return V2Export{}, err
	}
	if run.Report == nil || !isV2TerminalStatus(run.Status) {
		return V2Export{}, ErrV2InvalidRequest
	}
	var payload []byte
	if format == "pdf" {
		if s.renderPDF == nil {
			return V2Export{}, ErrServiceNotReady
		}
		payload, err = s.renderPDF(run)
	} else {
		payload, err = json.Marshal(map[string]any{"run": run, "report": run.Report, "format": format})
	}
	if err != nil {
		return V2Export{}, err
	}
	now := s.now()
	export := V2Export{RunID: input.RunID, Format: format, ExpiresAt: now.Add(7 * 24 * time.Hour), CreatedAt: now}
	export.DownloadURL = fmt.Sprintf("/api/v1/sandbox-runs/%d/exports/{id}/download", input.RunID)
	return repository.CreateV2Export(ctx, export, input.UserID, payload)
}

func (s *Service) GetV2Export(ctx context.Context, userID, exportID int64) (V2Export, []byte, error) {
	repository, ok := s.repository.(V2ExportRepository)
	if !ok || repository == nil {
		return V2Export{}, nil, ErrServiceNotReady
	}
	export, payload, err := repository.GetV2Export(ctx, userID, exportID)
	if err != nil {
		return V2Export{}, nil, err
	}
	if !export.ExpiresAt.After(s.now()) {
		return V2Export{}, nil, ErrV2ExportExpired
	}
	return export, payload, nil
}
