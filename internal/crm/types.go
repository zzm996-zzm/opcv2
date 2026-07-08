package crm

import (
	"errors"
	"time"
)

const (
	StageNew         = "new"
	StageContacted   = "contacted"
	StageQualified   = "qualified"
	StageProposal    = "proposal"
	StageWon         = "won"
	StageLost        = "lost"
	SourceLead       = "lead"
	SourceEnterprise = "enterprise"
	defaultListLimit = 20
)

const (
	ActivityStageChanged     = "stage_changed"
	ActivityFollowUpRecorded = "follow_up_recorded"
	ActivityCustomerUpdated  = "customer_updated"
)

var (
	ErrServiceNotReady  = errors.New("crm service is not configured")
	ErrInvalidInput     = errors.New("invalid crm input")
	ErrCustomerNotFound = errors.New("crm customer not found")
	ErrInvalidAIResult  = errors.New("invalid crm ai result")
)

type ImportLeadInput struct {
	UserID       int64  `json:"-"`
	LeadResultID int64  `json:"lead_result_id"`
	Name         string `json:"name"`
	Phone        string `json:"phone,omitempty"`
	Email        string `json:"email,omitempty"`
	Website      string `json:"website,omitempty"`
}

type ImportEnterpriseInput struct {
	UserID             int64  `json:"-"`
	DiagnosisRequestID int64  `json:"diagnosis_request_id"`
	Need               string `json:"need"`
}

type UpdateStageInput struct {
	UserID     int64  `json:"-"`
	CustomerID int64  `json:"-"`
	Stage      string `json:"stage"`
	Note       string `json:"note,omitempty"`
}

type UpdateCustomerInput struct {
	UserID     int64   `json:"-"`
	CustomerID int64   `json:"-"`
	Name       *string `json:"name,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Email      *string `json:"email,omitempty"`
	Website    *string `json:"website,omitempty"`
}

type RecordFollowUpInput struct {
	UserID         int64     `json:"-"`
	CustomerID     int64     `json:"-"`
	Note           string    `json:"note"`
	NextFollowUpAt time.Time `json:"next_follow_up_at"`
}

type ListDueInput struct {
	UserID int64
	Limit  int
}

type ListCustomersInput struct {
	UserID int64
	Stage  string
	Source string
	Q      string
	Limit  int
}

type ListFollowUpsInput struct {
	UserID       int64
	CustomerID   int64
	Q            string
	Due          string
	HasDueFrom   bool
	DueFrom      time.Time
	HasDueBefore bool
	DueBefore    time.Time
	Limit        int
}

type FollowUpCopyInput struct {
	UserID     int64  `json:"-"`
	CustomerID int64  `json:"-"`
	Goal       string `json:"goal"`
}

type FollowUpCopy struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Channel string `json:"channel"`
}

type Customer struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	ImportKey      string    `json:"import_key"`
	Name           string    `json:"name"`
	Phone          string    `json:"phone,omitempty"`
	Email          string    `json:"email,omitempty"`
	Website        string    `json:"website,omitempty"`
	Stage          string    `json:"stage"`
	Source         string    `json:"source"`
	NextFollowUpAt time.Time `json:"next_follow_up_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Activity struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	CustomerID int64     `json:"customer_id"`
	Type       string    `json:"type"`
	Note       string    `json:"note,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type FollowUp struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	CustomerID     int64     `json:"customer_id"`
	Note           string    `json:"note"`
	NextFollowUpAt time.Time `json:"next_follow_up_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type PipelineStats struct {
	Total     int `json:"total"`
	New       int `json:"new"`
	Contacted int `json:"contacted"`
	Qualified int `json:"qualified"`
	Proposal  int `json:"proposal"`
	Won       int `json:"won"`
	Lost      int `json:"lost"`
	DueToday  int `json:"due_today"`
}
