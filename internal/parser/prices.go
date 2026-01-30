package parser

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func MinPriceFromHTML(html string) (float64, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return 0, err
	}

	minPrice := math.MaxFloat64
	found := false

	// ВАЖНО: селектор может поменяться
	doc.Find(".tc-price").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		price := parsePrice(text)
		if price > 0 {
			found = true
			if price < minPrice {
				minPrice = price
			}
		}
	})

	if !found {
		return 0, errors.New("no prices found on page")
	}

	return minPrice, nil
}

func parsePrice(text string) float64 {
	// "12.34 ₽" → "12.34"
	clean := strings.ReplaceAll(text, "₽", "")
	clean = strings.ReplaceAll(clean, ",", ".")
	clean = strings.TrimSpace(clean)

	price, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0
	}

	return price
}

func MinPriceByName(html string, search string) (float64, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return 0, err
	}

	search = strings.ToLower(search)

	minPrice := math.MaxFloat64
	found := false

	doc.Find(".tc-item").Each(func(i int, s *goquery.Selection) {
		title := strings.ToLower(
			strings.TrimSpace(
				s.Find(".tc-desc-text").Text(),
			),
		)

		if !strings.Contains(title, search) {
			return
		}

		priceText := s.Find(".tc-price").Text()
		price := parsePrice(priceText)

		if price > 0 {
			found = true
			if price < minPrice {
				minPrice = price
			}
		}
	})

	if !found {
		return 0, errors.New("no matching lots found")
	}

	return minPrice, nil
}
