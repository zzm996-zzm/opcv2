package enterprise

type Metric struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
}

type Plan struct {
	ID         int64    `json:"id"`
	Title      string   `json:"title"`
	Audience   string   `json:"audience,omitempty"`
	PriceLabel string   `json:"price_label,omitempty"`
	Focus      []string `json:"focus"`
	Result     string   `json:"result,omitempty"`
}

type DeliveryItem struct {
	Stage  string `json:"stage"`
	Count  int    `json:"count"`
	Detail string `json:"detail,omitempty"`
}

type Milestone struct {
	TimeLabel string `json:"time_label"`
	Title     string `json:"title"`
	Detail    string `json:"detail,omitempty"`
}

type Case struct {
	ID      int64  `json:"id"`
	Company string `json:"company"`
	Result  string `json:"result,omitempty"`
}

type PublicStat struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
	Note  string `json:"note,omitempty"`
}

type PublicProofPoint struct {
	Title  string `json:"title"`
	Detail string `json:"detail,omitempty"`
}

type PublicServiceStep struct {
	Title  string `json:"title"`
	Detail string `json:"detail,omitempty"`
}

type PublicOverview struct {
	Headline        string              `json:"headline"`
	Subheadline     string              `json:"subheadline,omitempty"`
	Description     string              `json:"description,omitempty"`
	ProofPoints     []PublicProofPoint  `json:"proof_points"`
	Stats           []PublicStat        `json:"stats"`
	ServiceSteps    []PublicServiceStep `json:"service_steps"`
	SourceName      string              `json:"source_name,omitempty"`
	SourceURL       string              `json:"source_url,omitempty"`
	SourceUpdatedAt string              `json:"source_updated_at,omitempty"`
	UpdatedAt       string              `json:"updated_at,omitempty"`
}

type PublicCaseMetric struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type PublicCase struct {
	ID              int64              `json:"id"`
	Slug            string             `json:"slug"`
	Company         string             `json:"company"`
	Title           string             `json:"title"`
	Summary         string             `json:"summary,omitempty"`
	Result          string             `json:"result,omitempty"`
	Industry        string             `json:"industry,omitempty"`
	Services        []string           `json:"services"`
	Metrics         []PublicCaseMetric `json:"metrics"`
	Body            string             `json:"body,omitempty"`
	SourceName      string             `json:"source_name,omitempty"`
	SourceURL       string             `json:"source_url,omitempty"`
	SourceUpdatedAt string             `json:"source_updated_at,omitempty"`
	UpdatedAt       string             `json:"updated_at,omitempty"`
}

type PublicCasesResponse struct {
	Cases []PublicCase `json:"cases"`
}

type ContactConfig struct {
	ConsultantName string `json:"consultant_name,omitempty"`
	Title          string `json:"title,omitempty"`
	Description    string `json:"description,omitempty"`
	Phone          string `json:"phone,omitempty"`
	Email          string `json:"email,omitempty"`
	Wechat         string `json:"wechat,omitempty"`
	QRImageURL     string `json:"qr_image_url,omitempty"`
	ContactURL     string `json:"contact_url,omitempty"`
	SourceName     string `json:"source_name,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

type InquiryInput struct {
	Company    string `json:"company,omitempty"`
	Name       string `json:"name"`
	Phone      string `json:"phone,omitempty"`
	Email      string `json:"email,omitempty"`
	Wechat     string `json:"wechat,omitempty"`
	Need       string `json:"need"`
	Budget     string `json:"budget,omitempty"`
	Timeline   string `json:"timeline,omitempty"`
	SourcePage string `json:"source_page,omitempty"`
}

type Inquiry struct {
	ID            int64  `json:"id"`
	Company       string `json:"company,omitempty"`
	Name          string `json:"name"`
	Phone         string `json:"phone,omitempty"`
	Email         string `json:"email,omitempty"`
	Wechat        string `json:"wechat,omitempty"`
	Need          string `json:"need"`
	Budget        string `json:"budget,omitempty"`
	Timeline      string `json:"timeline,omitempty"`
	SourcePage    string `json:"source_page,omitempty"`
	Status        string `json:"status"`
	CRMCustomerID int64  `json:"crm_customer_id,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

type DiagnosisRequestInput struct {
	Need string `json:"need"`
}

type DiagnosisRequestUpdateInput struct {
	Status string `json:"status"`
}

type DiagnosisRequest struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Need      string `json:"need"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type DiagnosisRequestsResponse struct {
	Requests []DiagnosisRequest `json:"requests"`
}

type Overview struct {
	Stats         []Metric       `json:"stats"`
	Plans         []Plan         `json:"plans"`
	DeliveryBoard []DeliveryItem `json:"delivery_board"`
	Milestones    []Milestone    `json:"milestones"`
	Cases         []Case         `json:"cases"`
}
