package main

import (
	"log"

	"github.com/NM9371/FunpayMonitoring/internal/db"
	"github.com/NM9371/FunpayMonitoring/internal/telegram"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	pg, err := db.NewPostgres()
	if err != nil {
		log.Fatal(err)
	}

	bot, err := telegram.NewBot()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Bot is running. Send a message to it to get your chat ID...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.API.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		log.Printf("Received message from chat %d: %s", chatID, update.Message.Text)

		// Создаём подписку на тестовый лот
		sub := db.Subscription{
			UserID:   chatID,
			LotName:  "sins of the",
			MinPrice: 500.0,
			URL:      "https://funpay.com/lots/210/",
		}

		if err := pg.InsertSubscription(sub); err != nil {
			log.Println("Failed to insert subscription:", err)
		}

		bot.SendMessage(chatID, "✅ Подписка добавлена!")
	}
}
