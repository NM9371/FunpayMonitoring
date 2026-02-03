package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/NM9371/FunpayMonitoring/internal/domain/model"
)

type fakeSubsRepo struct {
	addCalled      int
	listByUserArgs []int64
	removeArgs     []struct {
		userID   int64
		category string
		lotName  string
	}

	addErr       error
	listByUser   []model.Subscription
	listByUserErr error
	removeErr    error
}

func (f *fakeSubsRepo) ListAll(ctx context.Context) ([]model.Subscription, error) {
	panic("not used")
}
func (f *fakeSubsRepo) ListByUser(ctx context.Context, userID int64) ([]model.Subscription, error) {
	f.listByUserArgs = append(f.listByUserArgs, userID)
	return f.listByUser, f.listByUserErr
}
func (f *fakeSubsRepo) Add(ctx context.Context, sub model.Subscription) error {
	f.addCalled++
	return f.addErr
}
func (f *fakeSubsRepo) Remove(ctx context.Context, userID int64, category string, lotName string) error {
	f.removeArgs = append(f.removeArgs, struct {
		userID   int64
		category string
		lotName  string
	}{userID: userID, category: category, lotName: lotName})
	return f.removeErr
}

func TestSubscriptionsService_Add_DelegatesToRepo(t *testing.T) {
	repo := &fakeSubsRepo{}
	svc := NewSubscriptionsService(repo)

	err := svc.Add(context.Background(), model.Subscription{UserID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.addCalled != 1 {
		t.Fatalf("expected Add to be called once, got %d", repo.addCalled)
	}
}

func TestSubscriptionsService_ListByUser_DelegatesToRepo(t *testing.T) {
	repo := &fakeSubsRepo{
		listByUser: []model.Subscription{{UserID: 7, Category: "c", LotName: "l"}},
	}
	svc := NewSubscriptionsService(repo)

	subs, err := svc.ListByUser(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 sub, got %d", len(subs))
	}
	if len(repo.listByUserArgs) != 1 || repo.listByUserArgs[0] != 7 {
		t.Fatalf("expected repo.ListByUser called with 7, got %+v", repo.listByUserArgs)
	}
}

func TestSubscriptionsService_Remove_DelegatesToRepo(t *testing.T) {
	repo := &fakeSubsRepo{}
	svc := NewSubscriptionsService(repo)

	err := svc.Remove(context.Background(), 9, "cat", "lot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.removeArgs) != 1 {
		t.Fatalf("expected 1 remove call, got %d", len(repo.removeArgs))
	}
	if repo.removeArgs[0].userID != 9 || repo.removeArgs[0].category != "cat" || repo.removeArgs[0].lotName != "lot" {
		t.Fatalf("unexpected args: %+v", repo.removeArgs[0])
	}
}

func TestSubscriptionsService_Add_ReturnsRepoError(t *testing.T) {
	repo := &fakeSubsRepo{addErr: errors.New("insert failed")}
	svc := NewSubscriptionsService(repo)

	err := svc.Add(context.Background(), model.Subscription{UserID: 1})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
