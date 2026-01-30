package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/NM9371/FunpayMonitoring/internal/fetcher"
	"github.com/NM9371/FunpayMonitoring/internal/parser"
)

func main() {
	url := flag.String(
		"url",
		"",
		"FunPay lots page URL (e.g. https://funpay.com/lots/210/)",
	)

	query := flag.String(
		"query",
		"",
		"Search text (part of item name)",
	)

	flag.Parse()

	if *url == "" || *query == "" {
		log.Fatal("usage: -url <funpay url> -query <search text>")
	}

	html, err := fetcher.FetchPage(*url)
	if err != nil {
		log.Fatal(err)
	}

	minPrice, err := parser.MinPriceByName(html, *query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"Минимальная цена для \"%s\": %.2f ₽\n",
		*query,
		minPrice,
	)
}
