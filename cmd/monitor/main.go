package main

import (
	"fmt"
	"log"
	"net/http"
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
			url := fmt.Sprintf("https://funpay.com/lots/%s/", sub.Category)
			log.Println(url)
			lots, err := getLots(url, sub.LotName)
			log.Println("hello")
			if err != nil {
				log.Println(err)
				continue
			}
			log.Printf("Found lot: '%s' price: %.2f", lots[0].Name, lots[0].Price)
			for _, lot := range lots {
				log.Printf("Found lot: '%s' price: %.2f", lot.Name, lot.Price)
				if lot.Price <= sub.MinPrice {
					msg := fmt.Sprintf(
						"💰 %s — %.2f\n%s",
						lot.Name, lot.Price, lot.Category,
					)
					tg.SendMessage(sub.UserID, msg)
				}
				pg.InsertPriceHistory(lot)
			}
		}

		time.Sleep(60 * time.Second)
	}
}

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

	query = strings.ToLower(query)
	var lots []db.Lot

	// Ищем все лоты с нужными классами
	doc.Find(".tc-item.offer-promo, .tc-item.lazyload-hidden.hidden").Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Find(".tc-desc .tc-desc-text").Text())
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

		lots = append(lots, db.Lot{
			Name:     name,
			Price:    price,
			Category: url, // можно заменить на отдельное поле CategoryID, если нужно
		})
	})

	if len(lots) == 0 {
		return nil, fmt.Errorf("parsing error: no matching lots found")
	}

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
