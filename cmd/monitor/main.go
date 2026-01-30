package main

import (
	"fmt"
	"log"

	"github.com/NM9371/FunpayMonitoring/internal/fetcher"
	"github.com/NM9371/FunpayMonitoring/internal/parser"
)

func main() {
	url := "https://funpay.com/lots/210/"

	html, err := fetcher.FetchPage(url)
	if err != nil {
		log.Fatal(err)
	}

	minPrice, err := parser.MinPriceFromHTML(html)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Минимальная цена: %.2f ₽\n", minPrice)
}
