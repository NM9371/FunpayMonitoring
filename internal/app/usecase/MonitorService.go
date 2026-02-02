package usecase

import (
	"context"
	"fmt"

	"github.com/NM9371/FunpayMonitoring/internal/domain/model"
	"github.com/NM9371/FunpayMonitoring/internal/domain/port"
)

type MonitorService struct {
	repo      port.SubscriptionRepository
	lots      port.LotProvider
	notifier  port.Notifier
}

func NewMonitorService(
	repo port.SubscriptionRepository,
	lots port.LotProvider,
	notifier port.Notifier,
) *MonitorService {
	return &MonitorService{
		repo:     repo,
		lots:     lots,
		notifier: notifier,
	}
}

func (m *MonitorService) CheckOnce(ctx context.Context) error {
	subs, err := m.repo.ListAll(ctx)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		found, err := m.lots.FindLots(ctx, sub.Category, sub.LotName)
		if err != nil {
			continue
		}

		lot := cheapest(found)
		if lot == nil {
			continue
		}

		if lot.Price <= sub.MinPrice {
			msg := fmt.Sprintf(
				"💰 Найден лот!\n\n%s\nЦена: %.2f\n%s",
				lot.Name,
				lot.Price,
				lot.URL,
			)

			_ = m.notifier.Notify(ctx, sub.UserID, msg)

			_ = m.repo.Remove(ctx, sub.UserID, sub.Category, sub.LotName)
		}
	}

	return nil
}

func cheapest(lots []model.Lot) *model.Lot {
	if len(lots) == 0 {
		return nil
	}
	min := lots[0]
	for _, l := range lots[1:] {
		if l.Price < min.Price {
			min = l
		}
	}
	return &min
}
