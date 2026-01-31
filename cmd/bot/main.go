package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/NM9371/FunpayMonitoring/internal/db"
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

		// Инициализация состояния пользователя
		if _, ok := states[chatID]; !ok {
			states[chatID] = &userState{Step: 0}
		}

		state := states[chatID]

		// Если это callback от кнопки
		if update.CallbackQuery != nil {
			switch text {
			case "add":
				state.Step = 1
				state.Category = ""
				state.LotName = ""
				state.MinPrice = 0
				bot.SendMessage(chatID, "Введите категорию (например 210):")
				continue
			case "list":
				subs, err := pg.GetSubscriptions()
				if err != nil {
					bot.SendMessage(chatID, "Ошибка получения подписок: "+err.Error())
					continue
				}
				var sb strings.Builder
				for _, s := range subs {
					if s.UserID == chatID {
						sb.WriteString(s.LotName)
						sb.WriteString(" | Категория: ")
						sb.WriteString(s.Category) // можно хранить категорию отдельно
						sb.WriteString(" | Мин. цена: ")
						sb.WriteString(strconv.FormatFloat(s.MinPrice, 'f', 2, 64))
						sb.WriteString("\n")
					}
				}
				if sb.Len() == 0 {
					bot.SendMessage(chatID, "У вас нет подписок")
				} else {
					bot.SendMessage(chatID, sb.String())
				}
				continue
			case "remove":
				bot.SendMessage(chatID, "Введите название подписки для удаления:")
				state.Step = -1 // шаг удаления
				continue
			}
		}

		// Обработка пошагового ввода
		switch state.Step {
		case 1: // ввод категории
			state.Category = text
			state.Step = 2
			bot.SendMessage(chatID, "Введите название лота для подписки:")
		case 2: // ввод названия
			state.LotName = text
			state.Step = 3
			bot.SendMessage(chatID, "Введите минимальную цену (только число):")
		case 3: // ввод минимальной цены
			price, err := strconv.ParseFloat(text, 64)
			if err != nil {
				bot.SendMessage(chatID, "Ошибка: введите число")
				continue
			}
			state.MinPrice = price

			// Сохраняем подписку
			sub := db.Subscription{
				UserID:   chatID,
				LotName:  state.LotName,
				MinPrice: state.MinPrice,
				Category: state.Category, // пока хранится категория в Category
			}
			if err := pg.InsertSubscription(sub); err != nil {
				bot.SendMessage(chatID, "Ошибка добавления подписки: "+err.Error())
			} else {
				bot.SendMessage(chatID, "✅ Подписка добавлена!")
			}
			state.Step = 0

		case -1: // удаление подписки
			// Удаляем подписку по имени
			err := pg.DeleteSubscription(chatID, text)
			if err != nil {
				bot.SendMessage(chatID, "❌ Ошибка при удалении подписки")
				log.Println("Failed to delete subscription:", err)
			} else {
				bot.SendMessage(chatID, "✅ Подписка удалена")
			}

			state.Step = 0

		default:
			// Показываем главные кнопки
			buttons := []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить подписку", "add"),
				tgbotapi.NewInlineKeyboardButtonData("📄 Просмотреть подписки", "list"),
				tgbotapi.NewInlineKeyboardButtonData("❌ Удалить подписку", "remove"),
			}
			kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
			bot.SendMessage(chatID, "Выберите действие:", kb)
		}
	}
}
