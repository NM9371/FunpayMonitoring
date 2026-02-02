package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/NM9371/FunpayMonitoring/internal/app/usecase"
	"github.com/NM9371/FunpayMonitoring/internal/db"
	"github.com/NM9371/FunpayMonitoring/internal/funpay"
	"github.com/NM9371/FunpayMonitoring/internal/telegram"
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

	fp := funpay.NewClient(http.DefaultClient)

	monitor := usecase.NewMonitorService(pg, fp, tg)

	log.Println("⏱ Monitor started")

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		_ = monitor.CheckOnce(ctx)
		cancel()

		time.Sleep(60 * time.Second)
	}
}
