package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/NM9371/FunpayMonitoring/internal/db"
	"github.com/NM9371/FunpayMonitoring/internal/telegram"
	"github.com/PuerkitoBio/goquery"
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
		log.Printf("Fetched %d subscriptions from DB", len(subs))
		if err != nil {
			log.Println(err)
			time.Sleep(30 * time.Second)
			continue
		}

		for _, sub := range subs {

			lots, err := getLots(sub.Category, sub.LotName)
			if err != nil {
				log.Println("Parsing error:", err)
				continue
			}

			lot := cheapestLot(lots)
			if lot == nil {
				continue
			}

			log.Printf(
				"User %d | %s | price %.2f / min %.2f",
				sub.UserID,
				lot.Name,
				lot.Price,
				sub.MinPrice,
			)

			if lot.Price <= sub.MinPrice {

				msg := fmt.Sprintf(
					"💰 Найден лот!\n\n%s\nЦена: %.2f\n%s",
					lot.Name,
					lot.Price,
					lot.URL,
				)

				tg.SendMessage(sub.UserID, msg)

				if err := pg.DeleteSubscription(sub.UserID, sub.LotName); err != nil {
					log.Println("Failed to delete subscription:", err)
				} else {
					log.Println("Subscription removed after notification")
				}
			}
		}

		time.Sleep(60 * time.Second)
	}
}

func cheapestLot(lots []db.Lot) *db.Lot {
	if len(lots) == 0 {
		return nil
	}

	min := lots[0]
	for _, lot := range lots[1:] {
		if lot.Price < min.Price {
			min = lot
		}
	}
	return &min
}

func getLots(category, query string) ([]db.Lot, error) {
	url := fmt.Sprintf("https://funpay.com/lots/%s/", category)
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

	query = strings.ToLower(query)
	var lots []db.Lot

	// Ищем все лоты с нужными классами
	doc.Find(".tc-item").Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Find(".tc-desc-text").Text())
		if name == "" {
			return
		}

		// Если передан query, фильтруем по имени
		if query != "" && !strings.Contains(strings.ToLower(name), query) {
			return
		}

		// Получаем цену из атрибута data-s
		priceStr, exists := s.Find(".tc-price").Attr("data-s")
		if !exists {
			return
		}
		price := parsePrice(priceStr)
		if price <= 0 {
			return
		}

		// Берём ссылку на лот
		href, ok := s.Attr("href")
		if !ok || href == "" {
			return
		}
		if strings.HasPrefix(href, "/") {
			href = "https://funpay.com" + href
		}

		lots = append(lots, db.Lot{
			Name:     name,
			Price:    price,
			URL:      href,
			Category: url,
		})
	})

	if len(lots) == 0 {
		return nil, fmt.Errorf("parsing error: no matching lots found")
	}

	// ВАЖНО: делаем самый дешёвый лот первым (чтобы lots[0] == min price)
	sort.Slice(lots, func(i, j int) bool {
		return lots[i].Price < lots[j].Price
	})

	return lots, nil
}

// parsePrice конвертирует строку в float64
func parsePrice(s string) float64 {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", ".")
	var price float64
	fmt.Sscanf(s, "%f", &price)
	return price
}
