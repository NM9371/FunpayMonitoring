package main

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/NM9371/FunpayMonitoring/internal/app/usecase"
	"github.com/NM9371/FunpayMonitoring/internal/db"
	"github.com/NM9371/FunpayMonitoring/internal/domain/model"
	"github.com/NM9371/FunpayMonitoring/internal/telegram"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// userState хранит состояние пользователя при пошаговом вводе
type userState struct {
	Step     int // 0 - нет действий, 1 - ввод категории, 2 - ввод названия, 3 - ввод цены
	Category string
	LotName  string
	MinPrice float64
}

var states = map[int64]*userState{}

func main() {
	pg, err := db.NewPostgres()
	if err != nil {
		log.Fatal(err)
	}

	bot, err := telegram.NewBot()
	if err != nil {
		log.Fatal(err)
	}

	subsService := usecase.NewSubscriptionsService(pg)

	log.Println("Bot is running...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.API.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		var chatID int64
		var text string
		if update.Message != nil {
			chatID = update.Message.Chat.ID
			text = update.Message.Text
		} else if update.CallbackQuery != nil {
			chatID = update.CallbackQuery.Message.Chat.ID
			text = update.CallbackQuery.Data
		}

		if _, ok := states[chatID]; !ok {
			states[chatID] = &userState{Step: 0}
		}
		state := states[chatID]

		if update.CallbackQuery != nil {
			switch text {
			case "add":
				state.Step = 1
				state.Category = ""
				state.LotName = ""
				state.MinPrice = 0
				bot.SendMessage(chatID, "Введите категорию из адресной строки:")
				continue

			case "list":
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				subs, err := subsService.ListByUser(ctx, chatID)
				cancel()

				if err != nil {
					bot.SendMessage(chatID, "Ошибка получения подписок: "+err.Error())
					continue
				}

				var sb strings.Builder
				for _, s := range subs {
					sb.WriteString(s.LotName)
					sb.WriteString(" | Категория: ")
					sb.WriteString(s.Category)
					sb.WriteString(" | Мин. цена: ")
					sb.WriteString(strconv.FormatFloat(s.MinPrice, 'f', 2, 64))
					sb.WriteString("\n")
				}

				if sb.Len() == 0 {
					bot.SendMessage(chatID, "У вас нет подписок")
				} else {
					bot.SendMessage(chatID, sb.String())
				}
				continue

			case "remove":
				bot.SendMessage(chatID, "Введите название подписки для удаления:")
				state.Step = -1
				continue
			}
		}

		switch state.Step {
		case 1:
			state.Category = text
			state.Step = 2
			bot.SendMessage(chatID, "Введите название лота для подписки:")

		case 2:
			state.LotName = text
			state.Step = 3
			bot.SendMessage(chatID, "Введите минимальную цену (только число):")

		case 3:
			price, err := strconv.ParseFloat(text, 64)
			if err != nil {
				bot.SendMessage(chatID, "Ошибка: введите число")
				continue
			}
			state.MinPrice = price

			sub := model.Subscription{
				UserID:   chatID,
				LotName:  state.LotName,
				MinPrice: state.MinPrice,
				Category: state.Category,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = subsService.Add(ctx, sub)
			cancel()

			if err != nil {
				bot.SendMessage(chatID, "Ошибка добавления подписки: "+err.Error())
			} else {
				bot.SendMessage(chatID, "✅ Подписка добавлена!")
			}
			state.Step = 0

		case -1:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := subsService.Remove(ctx, chatID, state.Category, text)
			cancel()

			if err != nil {
				bot.SendMessage(chatID, "❌ Ошибка при удалении подписки")
				log.Println("Failed to delete subscription:", err)
			} else {
				bot.SendMessage(chatID, "✅ Подписка удалена")
			}

			state.Step = 0

		default:
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("📄 Активные подписки", "list"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("➕ Добавить", "add"),
					tgbotapi.NewInlineKeyboardButtonData("❌ Удалить", "remove"),
				),
			)

			welcomeMessage := `Я отслеживаю цены на FunPay и отправляю уведомление,
когда появляется самый дешёвый лот по вашим условиям.

🔎 Как работает подписка:
• Вы указываете категорию (например: Dota 2 > Предметы > 210 (в адресной строке).
• Вводите текст для поиска в названии лота, аналогично как вы бы искали его на сайте.
• Задаёте минимальную цену, лоты с меньшей ценой будут отправлены уведомлением.

💡 Когда подходящий лот найден — я сразу присылаю название лота и ссылку.
❌ Подписка автоматически удаляется после уведомления.
`
			bot.SendMessage(chatID, welcomeMessage, keyboard)
		}
	}
}
