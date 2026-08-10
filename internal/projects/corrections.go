package projects

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"
)

var ErrInvalidCorrection = errors.New("invalid content correction")

type ContentCorrectionInput struct {
	TargetType  string `json:"target_type"`
	TargetID    int64  `json:"target_id"`
	Reason      string `json:"reason"`
	EvidenceURL string `json:"evidence_url,omitempty"`
	Contact     string `json:"contact,omitempty"`
}

type ContentCorrection struct {
	ID          int64     `json:"id"`
	TargetType  string    `json:"target_type"`
	TargetID    int64     `json:"target_id"`
	Reason      string    `json:"reason"`
	EvidenceURL string    `json:"evidence_url,omitempty"`
	Contact     string    `json:"contact,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type ContentCorrectionApplication interface {
	SubmitContentCorrection(context.Context, ContentCorrectionInput) (ContentCorrection, error)
}

type ContentCorrectionRepository interface {
	CreateContentCorrection(context.Context, ContentCorrection) (ContentCorrection, error)
}

func (s *Service) SubmitContentCorrection(ctx context.Context, input ContentCorrectionInput) (ContentCorrection, error) {
	repository, ok := s.repository.(ContentCorrectionRepository)
	if !ok {
		return ContentCorrection{}, ErrServiceNotReady
	}
	input.TargetType = strings.TrimSpace(input.TargetType)
	input.Reason = strings.TrimSpace(input.Reason)
	input.EvidenceURL = strings.TrimSpace(input.EvidenceURL)
	if input.TargetID <= 0 || input.Reason == "" || !map[string]bool{"failure": true, "plan": true, "case": true, "project": true}[input.TargetType] {
		return ContentCorrection{}, ErrInvalidCorrection
	}
	if input.EvidenceURL != "" {
		parsed, err := url.Parse(input.EvidenceURL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return ContentCorrection{}, ErrInvalidCorrection
		}
	}
	return repository.CreateContentCorrection(ctx, ContentCorrection{TargetType: input.TargetType, TargetID: input.TargetID, Reason: input.Reason, EvidenceURL: input.EvidenceURL, Contact: strings.TrimSpace(input.Contact), Status: "pending", CreatedAt: s.now()})
}
