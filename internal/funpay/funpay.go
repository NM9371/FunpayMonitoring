package funpay

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/NM9371/FunpayMonitoring/internal/domain/model"
	"github.com/PuerkitoBio/goquery"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    "https://funpay.com",
	}
}

func (c *Client) FindLots(ctx context.Context, category string, query string) ([]model.Lot, error) {
	url := fmt.Sprintf("%s/lots/%s/", c.baseURL, category)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
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
	var lots []model.Lot

	doc.Find(".tc-item").Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Find(".tc-desc-text").Text())
		if name == "" {
			return
		}

		if query != "" && !strings.Contains(strings.ToLower(name), query) {
			return
		}

		priceStr, exists := s.Find(".tc-price").Attr("data-s")
		if !exists {
			return
		}
		price := parsePrice(priceStr)
		if price <= 0 {
			return
		}

		href, ok := s.Attr("href")
		if !ok || href == "" {
			return
		}
		if strings.HasPrefix(href, "/") {
			href = c.baseURL + href
		}

		lots = append(lots, model.Lot{
			Name:     name,
			Price:    price,
			URL:      href,
			Category: category,
		})
	})

	if len(lots) == 0 {
		return nil, fmt.Errorf("no matching lots found")
	}

	sort.Slice(lots, func(i, j int) bool {
		return lots[i].Price < lots[j].Price
	})

	return lots, nil
}

func parsePrice(s string) float64 {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", ".")
	var price float64
	fmt.Sscanf(s, "%f", &price)
	return price
}
