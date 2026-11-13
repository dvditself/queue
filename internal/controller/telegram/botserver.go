package telegram

import (
"fmt"
"log/slog"

tgbotapi "github.com/OvyFlash/telegram-bot-api"

"QueueBot/internal/controller/telegram/client"
)

type BotServer struct {
bot *client.TelegramBot
}

func NewBotServer(bot *client.TelegramBot) *BotServer {
return &BotServer{bot: bot}
}

func (s BotServer) Listen(config tgbotapi.UpdateConfig, errChan chan<- error) {
updates := s.bot.TgBot.GetUpdatesChan(config)
slog.Info("Started listening update channel")

for update := range updates {
go func(update tgbotapi.Update) {
switch {
case update.Message != nil:
if err := s.HandleMessage(update.Message); err != nil {
errChan <- fmt.Errorf("couldn't handle message: %w", err)
}
case update.CallbackQuery != nil:
if err := s.HandleCallbackQuery(update.CallbackQuery); err != nil {
errChan <- fmt.Errorf("couldn't handle callback query: %w", err)
}
}
}(update)
}
}
