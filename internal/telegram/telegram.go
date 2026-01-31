package telegram

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	API *tgbotapi.BotAPI
}

// NewBot создаёт и возвращает Telegram Bot
func NewBot() (*Bot, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, ErrTokenNotSet
	}

	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Telegram bot authorized on account %s", botAPI.Self.UserName)

	return &Bot{API: botAPI}, nil
}

// SendMessage отправляет сообщение пользователю
func (b *Bot) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.API.Send(msg)
	if err != nil {
		log.Println("Failed to send Telegram message:", err)
	}
}

// Ошибка, если токен не установлен
var ErrTokenNotSet = &BotError{"TELEGRAM_BOT_TOKEN is not set"}

type BotError struct{ Msg string }

func (e *BotError) Error() string { return e.Msg }
