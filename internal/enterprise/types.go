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

type DiagnosisRequestInput struct {
	Need string `json:"need"`
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
