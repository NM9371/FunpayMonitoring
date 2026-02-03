package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/NM9371/FunpayMonitoring/internal/domain/model"
)

type fakeRepo struct {
	listAllResult []model.Subscription
	listAllErr    error

	removed []struct {
		userID   int64
		category string
		lotName  string
	}
	removeErr error
}

func (f *fakeRepo) ListAll(ctx context.Context) ([]model.Subscription, error) {
	return f.listAllResult, f.listAllErr
}
func (f *fakeRepo) ListByUser(ctx context.Context, userID int64) ([]model.Subscription, error) {
	panic("not used in these tests")
}
func (f *fakeRepo) Add(ctx context.Context, sub model.Subscription) error {
	panic("not used in these tests")
}
func (f *fakeRepo) Remove(ctx context.Context, userID int64, category string, lotName string) error {
	f.removed = append(f.removed, struct {
		userID   int64
		category string
		lotName  string
	}{userID: userID, category: category, lotName: lotName})
	return f.removeErr
}

type fakeLots struct {
	byKey map[string][]model.Lot
	err   error
}

func (f *fakeLots) FindLots(ctx context.Context, category string, query string) ([]model.Lot, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.byKey[category+"|"+query], nil
}

type fakeNotifier struct {
	sent []struct {
		userID  int64
		message string
	}
	err error
}

func (f *fakeNotifier) Notify(ctx context.Context, userID int64, message string) error {
	f.sent = append(f.sent, struct {
		userID  int64
		message string
	}{userID: userID, message: message})
	return f.err
}

func TestMonitorService_CheckOnce_ReturnsErrorWhenRepoFails(t *testing.T) {
	repo := &fakeRepo{listAllErr: errors.New("db down")}
	lots := &fakeLots{byKey: map[string][]model.Lot{}}
	ntf := &fakeNotifier{}

	svc := NewMonitorService(repo, lots, ntf)

	err := svc.CheckOnce(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestMonitorService_CheckOnce_SendsNotificationAndRemovesSubscription(t *testing.T) {
	repo := &fakeRepo{
		listAllResult: []model.Subscription{
			{UserID: 10, Category: "cat", LotName: "item", MinPrice: 100},
		},
	}
	lots := &fakeLots{
		byKey: map[string][]model.Lot{
			"cat|item": {
				{Name: "item AAA", Price: 150, URL: "u1", Category: "cat"},
				{Name: "item BBB", Price: 99, URL: "u2", Category: "cat"}, // cheapest, подходит
			},
		},
	}
	ntf := &fakeNotifier{}

	svc := NewMonitorService(repo, lots, ntf)

	err := svc.CheckOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ntf.sent) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(ntf.sent))
	}
	if ntf.sent[0].userID != 10 {
		t.Fatalf("expected userID=10, got %d", ntf.sent[0].userID)
	}
	if len(repo.removed) != 1 {
		t.Fatalf("expected 1 removal, got %d", len(repo.removed))
	}
	if repo.removed[0].userID != 10 || repo.removed[0].category != "cat" || repo.removed[0].lotName != "item" {
		t.Fatalf("unexpected removed key: %+v", repo.removed[0])
	}
}

func TestMonitorService_CheckOnce_DoesNothingWhenPriceHigherThanMin(t *testing.T) {
	repo := &fakeRepo{
		listAllResult: []model.Subscription{
			{UserID: 10, Category: "cat", LotName: "item", MinPrice: 50},
		},
	}
	lots := &fakeLots{
		byKey: map[string][]model.Lot{
			"cat|item": {
				{Name: "item", Price: 60, URL: "u", Category: "cat"},
			},
		},
	}
	ntf := &fakeNotifier{}

	svc := NewMonitorService(repo, lots, ntf)

	_ = svc.CheckOnce(context.Background())

	if len(ntf.sent) != 0 {
		t.Fatalf("expected 0 notifications, got %d", len(ntf.sent))
	}
	if len(repo.removed) != 0 {
		t.Fatalf("expected 0 removals, got %d", len(repo.removed))
	}
}

func TestMonitorService_CheckOnce_SkipsSubscriptionWhenLotProviderFails(t *testing.T) {
	repo := &fakeRepo{
		listAllResult: []model.Subscription{
			{UserID: 10, Category: "cat", LotName: "item", MinPrice: 100},
		},
	}
	lots := &fakeLots{err: errors.New("funpay unavailable")}
	ntf := &fakeNotifier{}

	svc := NewMonitorService(repo, lots, ntf)

	_ = svc.CheckOnce(context.Background())

	if len(ntf.sent) != 0 {
		t.Fatalf("expected 0 notifications, got %d", len(ntf.sent))
	}
	if len(repo.removed) != 0 {
		t.Fatalf("expected 0 removals, got %d", len(repo.removed))
	}
}
