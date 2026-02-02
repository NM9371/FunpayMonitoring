package telegram

import (
	"context"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	API *tgbotapi.BotAPI
}

func NewBot() (*Bot, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, ErrTokenNotSet
	}

	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Telegram bot authorized as %s", botAPI.Self.UserName)
	return &Bot{API: botAPI}, nil
}

func (b *Bot) SendMessage(chatID int64, text string, keyboard ...tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	if len(keyboard) > 0 {
		msg.ReplyMarkup = keyboard[0]
	}
	_, err := b.API.Send(msg)
	if err != nil {
		log.Println("Failed to send Telegram message:", err)
	}
}

func (b *Bot) Notify(ctx context.Context, userID int64, message string) error {
	_ = ctx
	msg := tgbotapi.NewMessage(userID, message)
	_, err := b.API.Send(msg)
	return err
}

var ErrTokenNotSet = &BotError{"TELEGRAM_BOT_TOKEN is not set"}

type BotError struct{ Msg string }

func (e *BotError) Error() string { return e.Msg }
