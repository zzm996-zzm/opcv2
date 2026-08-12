package projects

import (
	"context"
	"encoding/json"
)

func (s *Service) ProcessProjectExport(ctx context.Context, userID, exportID int64) error {
	item, err := s.repository.GetExport(ctx, userID, exportID)
	if err != nil {
		return err
	}
	if item.Status == "ready" {
		return nil
	}
	item.Status, item.ErrorCode, item.UpdatedAt = "running", "", s.now()
	item, err = s.repository.UpdateExport(ctx, item)
	if err != nil {
		return err
	}
	var payload []byte
	switch item.Format {
	case "pdf":
		if s.renderProjectPDF == nil {
			err = ErrServiceNotReady
		} else {
			payload, err = s.renderProjectPDF(item)
		}
	case "json", "link":
		var snapshot any
		if err = json.Unmarshal(item.Snapshot, &snapshot); err == nil {
			payload, err = json.Marshal(map[string]any{"format": item.Format, "snapshot": snapshot})
		}
	default:
		err = ErrInvalidExport
	}
	if err != nil {
		item.Status, item.ErrorCode, item.UpdatedAt = "failed", "render_failed", s.now()
		_, _ = s.repository.UpdateExport(ctx, item)
		return err
	}
	item.Status, item.Payload, item.ErrorCode, item.UpdatedAt = "ready", payload, "", s.now()
	_, err = s.repository.UpdateExport(ctx, item)
	return err
}
