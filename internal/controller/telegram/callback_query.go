package telegram

import (
"context"
"errors"
"fmt"
"log/slog"

tgbotapi "github.com/OvyFlash/telegram-bot-api"

"QueueBot/internal/controller/telegram/client"
)

func (s BotServer) handleCallbackData(cb *tgbotapi.CallbackQuery) error {
ctx := context.Background()
chatID := cb.Message.Chat.ID
userID := cb.From.ID

isAdmin, err := s.bot.IsChatAdmin(ctx, chatID, userID)
if err != nil {
return fmt.Errorf("couldn't check admin: %w", err)
}
if !isAdmin {
return client.ErrNotAdmin
}

switch cb.Data {
case client.StartQueueData:
return s.bot.Start(ctx, cb, false)
case client.StartQueueShuffleData:
return s.bot.Start(ctx, cb, true)
case client.KickFirstData:
return s.bot.KickFirst(ctx, cb)
case client.NextData:
return s.bot.Next(ctx, cb)
case client.GoToMenuData:
return s.bot.GoToMenu(ctx, cb)
case client.FinishQueueData:
return s.bot.FinishQueue(ctx, cb)
}
return nil
}

func (s BotServer) HandleCallbackQuery(cb *tgbotapi.CallbackQuery) error {
if cb.Message == nil {
return nil
}
err := s.handleCallbackData(cb)

alertText := ""
if err != nil {
switch {
case errors.Is(err, client.ErrNotAdmin):
alertText = client.OnlyAdminMessage
case errors.Is(err, client.ErrNoQueue):
alertText = client.NoActiveQueueMessage
default:
slog.Error("Couldn't handle callback", "reason", err, "data", cb.Data, "user_id", cb.From.ID)
alertText = "Ошибка"
}
}

callback := tgbotapi.NewCallback(cb.ID, alertText)
if _, err := s.bot.TgBot.Request(callback); err != nil {
return fmt.Errorf("couldn't process callback: %w", err)
}
return nil
}