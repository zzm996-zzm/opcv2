package leads

import "context"

type DevelopmentProvider struct{}

func NewDevelopmentProvider() *DevelopmentProvider {
	return &DevelopmentProvider{}
}

func (p *DevelopmentProvider) SearchLeads(_ context.Context, input SearchInput) ([]Lead, error) {
	return []Lead{
		{
			Name:    "成都启明星教育咨询有限公司",
			Phone:   "028-12345678",
			Website: "https://example.com",
			Evidence: []Evidence{
				{Type: "website", Title: "官网", URL: "https://example.com"},
			},
		},
	}, nil
}

type DevelopmentCreditLedger struct{}

func NewDevelopmentCreditLedger() *DevelopmentCreditLedger {
	return &DevelopmentCreditLedger{}
}

func (l *DevelopmentCreditLedger) Charge(context.Context, int64, int, string, int64) error {
	return nil
}

func (l *DevelopmentCreditLedger) Refund(context.Context, int64, int, string, int64) error {
	return nil
}
