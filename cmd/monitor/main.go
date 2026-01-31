package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/NM9371/FunpayMonitoring/internal/db"
	"github.com/NM9371/FunpayMonitoring/internal/telegram"
	"github.com/PuerkitoBio/goquery"
	"net/http"
)

func main() {
	pg, err := db.NewPostgres()
	if err != nil {
		log.Fatal(err)
	}

	tg, err := telegram.NewBot()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("⏱ Monitor started")

	for {
		subs, err := pg.GetSubscriptions()
		if err != nil {
			log.Println("Failed to get subscriptions:", err)
			time.Sleep(30 * time.Second)
			continue
		}

		for _, sub := range subs {
			lots, err := getLots(sub.URL, sub.LotName)
			if err != nil {
				log.Println("Failed to fetch lots:", err)
				continue
			}

			for _, lot := range lots {
				// Если цена ниже минимальной, отправляем уведомление
				if lot.Price <= sub.MinPrice {
					msg := fmt.Sprintf(
						"💰 Найден лот '%s' по цене %.2f (минимальная: %.2f)\n%s",
						lot.Name, lot.Price, sub.MinPrice, lot.URL,
					)
					tg.SendMessage(sub.UserID, msg)
				}

				// Сохраняем цену в историю
				if err := pg.InsertPriceHistory(lot); err != nil {
					log.Println("Failed to insert price:", err)
				}
			}
		}

		// Проверяем каждые 60 секунд
		time.Sleep(60 * time.Second)
	}
}

// getLots парсит страницу и возвращает список лотов, подходящих под LotName
func getLots(url, query string) ([]db.Lot, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var lots []db.Lot
	doc.Find(".tc-list__item").Each(func(i int, s *goquery.Selection) {
		name := s.Find(".tc-desc-text").Text()
		priceStr := s.Find(".tc-price__value").Text()

		if strings.Contains(strings.ToLower(name), strings.ToLower(query)) {
			price := parsePrice(priceStr)
			lots = append(lots, db.Lot{
				Name:  name,
				Price: price,
				URL:   url,
			})
		}
	})

	return lots, nil
}

// parsePrice конвертирует строку с ценой в float64
func parsePrice(s string) float64 {
	s = strings.ReplaceAll(s, "₽", "")
	s = strings.ReplaceAll(s, " ", "")
	var price float64
	fmt.Sscanf(s, "%f", &price)
	return price
}
