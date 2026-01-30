package main

import (
	"fmt"
	"log"
	"os"

	"github.com/NM9371/FunpayMonitoring/internal/fetcher"
)

func main() {
	url := "https://funpay.com/"

	html, err := fetcher.FetchPage(url)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("funpay.html", []byte(html), 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("FunPay page downloaded: funpay.html")
}
