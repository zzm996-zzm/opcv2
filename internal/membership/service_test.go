package membership

import (
	"context"
	"errors"
	"testing"
	"time"
)

type memoryRepository struct {
	account       CreditAccount
	transactions  []CreditTransaction
	codes         map[string]*RedemptionCode
	redemptions   map[string]Redemption
	subscriptions map[int64]Subscription
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{
		codes:         map[string]*RedemptionCode{},
		redemptions:   map[string]Redemption{},
		subscriptions: map[int64]Subscription{},
	}
}

func (r *memoryRepository) CurrentSnapshot(_ context.Context, userID int64, now time.Time) (Snapshot, error) {
	subscription := r.subscriptions[userID]
	planCode := PlanFree
	if subscription.Status == "active" && (subscription.EndsAt == nil || subscription.EndsAt.After(now)) {
		planCode = subscription.PlanCode
	}
	return Snapshot{
		Plan:          PlanCatalog[planCode],
		Subscription:  subscription,
		CreditBalance: r.account.Balance,
	}, nil
}

func (r *memoryRepository) RedeemCode(ctx context.Context, input RedeemInput, now time.Time) (RedeemResult, error) {
	code, ok := r.codes[input.Code]
	if !ok || !code.Active {
		return RedeemResult{}, ErrCodeNotFound
	}
	if code.ExpiresAt != nil && !code.ExpiresAt.After(now) {
		return RedeemResult{}, ErrCodeExpired
	}
	key := input.Code + ":" + string(rune(input.UserID))
	if redemption, ok := r.redemptions[key]; ok {
		snapshot, err := r.CurrentSnapshot(ctx, input.UserID, now)
		if err != nil {
			return RedeemResult{}, err
		}
		return RedeemResult{Snapshot: snapshot, Redemption: redemption, AlreadyRedeemed: true}, nil
	}
	if code.MaxRedemptions > 0 && code.RedeemedCount >= code.MaxRedemptions {
		return RedeemResult{}, ErrCodeExhausted
	}

	if code.CreditAmount > 0 {
		r.account.UserID = input.UserID
		r.account.Balance += code.CreditAmount
		r.transactions = append(r.transactions, CreditTransaction{
			UserID: input.UserID,
			Amount: code.CreditAmount,
			Reason: "redemption",
		})
	}
	if code.PlanCode != "" {
		r.subscriptions[input.UserID] = Subscription{
			UserID:   input.UserID,
			PlanCode: code.PlanCode,
			Status:   "active",
			StartsAt: now,
		}
	}
	code.RedeemedCount++
	redemption := Redemption{UserID: input.UserID, Code: input.Code, CreatedAt: now}
	r.redemptions[key] = redemption
	snapshot, err := r.CurrentSnapshot(ctx, input.UserID, now)
	if err != nil {
		return RedeemResult{}, err
	}
	return RedeemResult{Snapshot: snapshot, Redemption: redemption}, nil
}

func TestServiceReturnsFreeSnapshotForNewUser(t *testing.T) {
	service := NewService(newMemoryRepository())

	snapshot, err := service.CurrentSnapshot(context.Background(), 42)
	if err != nil {
		t.Fatalf("CurrentSnapshot() error = %v", err)
	}
	if snapshot.Plan.Code != PlanFree || snapshot.CreditBalance != 0 {
		t.Fatalf("snapshot = %+v", snapshot)
	}
}

func TestServiceRedeemsCreditCodeOnce(t *testing.T) {
	repository := newMemoryRepository()
	repository.codes["WELCOME100"] = &RedemptionCode{
		Code:         "WELCOME100",
		CreditAmount: 100,
		Active:       true,
	}
	service := NewService(repository)

	first, err := service.Redeem(context.Background(), RedeemInput{UserID: 42, Code: " welcome100 "})
	if err != nil {
		t.Fatalf("first Redeem() error = %v", err)
	}
	second, err := service.Redeem(context.Background(), RedeemInput{UserID: 42, Code: "WELCOME100"})
	if err != nil {
		t.Fatalf("second Redeem() error = %v", err)
	}

	if first.Snapshot.CreditBalance != 100 || second.Snapshot.CreditBalance != 100 {
		t.Fatalf("balances = %d/%d, want 100/100", first.Snapshot.CreditBalance, second.Snapshot.CreditBalance)
	}
	if !second.AlreadyRedeemed {
		t.Fatal("second redeem should be idempotent")
	}
	if len(repository.transactions) != 1 {
		t.Fatalf("transactions = %d, want 1", len(repository.transactions))
	}
}

func TestServiceRedeemsPlanCode(t *testing.T) {
	repository := newMemoryRepository()
	repository.codes["PRO30"] = &RedemptionCode{
		Code:     "PRO30",
		PlanCode: PlanPro,
		Active:   true,
	}
	service := NewService(repository)

	result, err := service.Redeem(context.Background(), RedeemInput{UserID: 42, Code: "PRO30"})
	if err != nil {
		t.Fatalf("Redeem() error = %v", err)
	}
	if result.Snapshot.Plan.Code != PlanPro {
		t.Fatalf("plan = %q, want %q", result.Snapshot.Plan.Code, PlanPro)
	}
}

func TestServiceRejectsInvalidCode(t *testing.T) {
	service := NewService(newMemoryRepository())

	_, err := service.Redeem(context.Background(), RedeemInput{UserID: 42, Code: "NOPE"})
	if !errors.Is(err, ErrCodeNotFound) {
		t.Fatalf("Redeem() error = %v, want ErrCodeNotFound", err)
	}
}
