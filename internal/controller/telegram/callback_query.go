package telegram

import (
"context"
"errors"
"fmt"
"log/slog"
"strings"

tgbotapi "github.com/OvyFlash/telegram-bot-api"

"QueueBot/internal/controller/telegram/client"
)

func (s BotServer) handleCallbackData(cb *tgbotapi.CallbackQuery) error {
ctx := context.Background()

// ConfirmClose приходит как "close_confirm_yes:<key>"
if strings.HasPrefix(cb.Data, client.ConfirmCloseData+":") {
key := strings.TrimPrefix(cb.Data, client.ConfirmCloseData+":")
return s.bot.ConfirmClose(ctx, cb, key)
}
if cb.Data == client.CancelCloseData {
return s.bot.CancelClose(ctx, cb)
}

// всё остальное — только для админов
chatID := cb.Message.Chat.ID
userID := cb.From.ID
isAdmin, err := s.bot.IsChatAdmin(ctx, chatID, userID)
if err != nil {
return err
}
if !isAdmin {
return client.ErrNotAdmin
}

switch cb.Data {
case client.StartQueueData:
return s.bot.Start(ctx, cb)
case client.KickFirstData:
return s.bot.KickFirst(ctx, cb)
case client.NextData:
return s.bot.Next(ctx, cb)
case client.GoToMenuData:
return s.bot.GoToMenu(ctx, cb)
case client.FinishQueueData:
// "Закрыть очередь" / "Закончить" — отправляем подтверждение
return s.bot.RequestCloseConfirm(ctx, cb)
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
} else if cb.Data == client.FinishQueueData {
alertText = "Подтвердите закрытие"
}

callback := tgbotapi.NewCallback(cb.ID, alertText)
if _, err := s.bot.TgBot.Request(callback); err != nil {
return fmt.Errorf("couldn't process callback: %w", err)
}
return nil
}