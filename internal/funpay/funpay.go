package funpay

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func FetchLowestPrice(url, lotName string) (float64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, err
	}

	minPrice := 1e9

	doc.Find(".tc-lot-card").Each(func(i int, s *goquery.Selection) {
		name := s.Find(".tc-desc-text").Text()
		priceStr := s.Find(".tc-price").Text()

		if !strings.Contains(strings.ToLower(name), strings.ToLower(lotName)) {
			return
		}

		priceStr = strings.ReplaceAll(priceStr, "₽", "")
		priceStr = strings.ReplaceAll(priceStr, " ", "")
		priceStr = strings.ReplaceAll(priceStr, ",", ".")

		var price float64
		fmt.Sscanf(priceStr, "%f", &price)

		if price < minPrice {
			minPrice = price
		}
	})

	if minPrice == 1e9 {
		return 0, fmt.Errorf("no matching lots found")
	}

	return minPrice, nil
}
