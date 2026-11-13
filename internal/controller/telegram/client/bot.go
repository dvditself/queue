package client

import (
"context"
"fmt"
"log/slog"

tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

"QueueBot/internal/entity"
"QueueBot/internal/usecase"
)

const HelloMessage = `Привет! Я бот для создания очередей.

Как пользоваться:
• /create <название> — создать очередь (например, /create ЛР3)
• /help — справка
• /start — это сообщение`

const CreateUsageMessage = `Использование: /create <название>
Пример: /create ЛР3`

const HelpMessage = `Команды бота:
/create <название> — создать очередь
/start — приветствие
/help — эта справка

После создания очереди под сообщением появятся кнопки:
• Добавиться/выйти из очереди
• Старт в порядке очереди
• Старт в случайном порядке`

type TelegramBot struct {
TgBot *tgbotapi.BotAPI
u     usecase.Bot
}

func NewTelegramBot(tgBot *tgbotapi.BotAPI, u usecase.Bot) *TelegramBot {
return &TelegramBot{TgBot: tgBot, u: u}
}

func queueKey(chatID int64, messageID int) string {
return fmt.Sprintf("%d:%d", chatID, messageID)
}

func keyFromCallback(cb *tgbotapi.CallbackQuery) string {
return queueKey(cb.Message.Chat.ID, cb.Message.MessageID)
}

func (b TelegramBot) SendHelloMessage(message *tgbotapi.Message) error {
msg := tgbotapi.NewMessage(message.Chat.ID, HelloMessage)
if _, err := b.TgBot.Send(msg); err != nil {
return fmt.Errorf("couldn't send hello message: %w", err)
}
return nil
}

func (b TelegramBot) SendCreateUsage(chatID int64) error {
msg := tgbotapi.NewMessage(chatID, CreateUsageMessage)
if _, err := b.TgBot.Send(msg); err != nil {
return fmt.Errorf("couldn't send create usage: %w", err)
}
return nil
}

func (b TelegramBot) SendHelpMessage(chatID int64) error {
msg := tgbotapi.NewMessage(chatID, HelpMessage)
if _, err := b.TgBot.Send(msg); err != nil {
return fmt.Errorf("couldn't send help message: %w", err)
}
return nil
}

func (b TelegramBot) CreateQueueFromCommand(chatID int64, description string) error {
msg := tgbotapi.NewMessage(chatID, getMessageContentBeforeStart(description, nil))
msg.ParseMode = tgbotapi.ModeMarkdown
keyboard := GetBeforeStartKeyboard()
msg.ReplyMarkup = keyboard

sent, err := b.TgBot.Send(msg)
if err != nil {
return fmt.Errorf("couldn't send queue message: %w", err)
}

key := queueKey(chatID, sent.MessageID)
if err := b.u.CreateQueue(context.Background(), key, description); err != nil {
return fmt.Errorf("couldn't create queue: %w", err)
}

slog.Info("Queue created", "key", key, "description", description)
return nil
}

func (b TelegramBot) LogInOurOut(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) error {
key := keyFromCallback(callbackQuery)

if err := b.u.LogInOutToQueue(
ctx,
key,
entity.New(callbackQuery.From.ID, callbackQuery.From.LastName, callbackQuery.From.FirstName),
); err != nil {
return fmt.Errorf("couldn't add user to queue: %w", err)
}

queue, err := b.u.GetQueue(ctx, key)
if err != nil {
return fmt.Errorf("couldn't get queue: %w", err)
}

updatedMessage := GetUpdatedQueueMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, queue.Description, queue.Users)
if _, err := b.TgBot.Request(updatedMessage); err != nil {
return fmt.Errorf("couldn't update message: %w", err)
}

return nil
}

func (b TelegramBot) Start(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, isShuffled bool) error {
key := keyFromCallback(callbackQuery)
if err := b.u.StartQueue(ctx, key, isShuffled); err != nil {
return fmt.Errorf("couldn't start queue: %w", err)
}
slog.Info("Started queue", "key", key, "isShuffled", isShuffled)
return b.sendQueueStatusMessage(ctx, callbackQuery)
}

func (b TelegramBot) Next(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) error {
key := keyFromCallback(callbackQuery)
if err := b.u.SetNextPersonToQueue(ctx, key); err != nil {
return fmt.Errorf("couldn't increment current person: %w", err)
}
slog.Info("Set next person", "key", key)
return b.sendQueueStatusMessage(ctx, callbackQuery)
}

func (b TelegramBot) GoToMenu(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) error {
key := keyFromCallback(callbackQuery)
queue, err := b.u.GetQueue(ctx, key)
if err != nil {
return fmt.Errorf("couldn't get queue: %w", err)
}

updatedMessage := GetQueueMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, queue.Users, queue.Description)
if _, err := b.TgBot.Request(updatedMessage); err != nil {
return fmt.Errorf("couldn't go to menu: %w", err)
}
slog.Info("Went to menu", "key", key)
return nil
}

func (b TelegramBot) FinishQueue(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) error {
key := keyFromCallback(callbackQuery)
updatedMessage := GetFinishedMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
if _, err := b.TgBot.Request(updatedMessage); err != nil {
return fmt.Errorf("couldn't send finish queue: %w", err)
}
if err := b.u.FinishQueue(ctx, key); err != nil {
return fmt.Errorf("couldn't finish queue: %w", err)
}
slog.Info("Finished queue", "key", key)
return nil
}

func (b TelegramBot) sendQueueStatusMessage(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) error {
key := keyFromCallback(callbackQuery)
queue, err := b.u.GetQueue(ctx, key)
if err != nil {
return fmt.Errorf("couldn't get queue: %w", err)
}

var updatedMessage tgbotapi.EditMessageTextConfig
if queue.CurrentPersonIdx == len(queue.Users) {
updatedMessage = GetEndQueueMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
} else {
updatedMessage = GetQueueAfterStartMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, queue.Description, queue.Users, queue.CurrentPersonIdx)
}

if _, err := b.TgBot.Request(updatedMessage); err != nil {
return fmt.Errorf("couldn't update message after starting queue: %w", err)
}
return nil
}
