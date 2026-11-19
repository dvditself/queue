package telegram

import (
"errors"
"strconv"
"strings"

tgbotapi "github.com/OvyFlash/telegram-bot-api"

"QueueBot/internal/controller/telegram/client"
)

const (
StartCommand     = "start"
CreateCommand    = "create"
HelpCommand      = "help"
JoinCommand      = "join"
LeaveCommand     = "leave"
CloseCommand     = "close"
KickFirstCommand = "kickfirst"
KickCommand      = "kick"
SwitchCommand    = "switch"
Swap12Command    = "swap12"
PositionsCommand = "positions"
DebugCommand     = "debug"
)

func parseJoinArgs(args string) (fullName string, position int) {
args = strings.TrimSpace(args)
if args == "" {
return "", 0
}
parts := strings.Fields(args)
if len(parts) >= 2 {
last := parts[len(parts)-1]
if strings.HasPrefix(last, "#") && len(last) > 1 {
if p, err := strconv.Atoi(last[1:]); err == nil && p > 0 {
return strings.Join(parts[:len(parts)-1], " "), p
}
}
if p, err := strconv.Atoi(last); err == nil && p > 0 {
return strings.Join(parts[:len(parts)-1], " "), p
}
}
return args, 0
}

func parsePositionArg(args string) (int, bool) {
args = strings.TrimSpace(args)
if args == "" {
return 0, false
}
if p, err := strconv.Atoi(args); err == nil && p > 0 {
return p, true
}
return 0, false
}

func (s BotServer) deleteUserCommand(message *tgbotapi.Message) {
if message == nil {
return
}
_, _ = s.bot.TgBot.Request(tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID))
}

func (s BotServer) HandleMessage(message *tgbotapi.Message) error {
if !message.IsCommand() {
return nil
}
defer s.deleteUserCommand(message)

chatID := message.Chat.ID
userID := message.From.ID
threadID := message.MessageThreadID
if threadID == 1 {
threadID = 0
}

switch message.Command() {
case StartCommand:
return s.bot.SendHelloMessage(message)
case HelpCommand:
return s.bot.SendHelpMessage(chatID, threadID)
case CreateCommand:
desc := strings.TrimSpace(message.CommandArguments())
if desc == "" {
return s.bot.SendCreateUsage(chatID, threadID)
}
err := s.bot.CreateQueueFromCommand(chatID, threadID, desc)
if err == nil {
return nil
}
if errors.Is(err, client.ErrQueueAlreadyOpen) {
m := tgbotapi.NewMessage(chatID, "В этой теме уже открыта очередь. Закройте её через /close, чтобы создать новую.")
m.MessageThreadID = threadID
_, _ = s.bot.TgBot.Send(m)
return nil
}
return err
case JoinCommand:
name, pos := parseJoinArgs(message.CommandArguments())
if name == "" {
return s.bot.SendJoinUsage(chatID, threadID)
}
err := s.bot.JoinQueueFromCommand(chatID, userID, threadID, name, pos)
if err == nil {
return nil
}
switch {
case errors.Is(err, client.ErrNoQueue):
m := tgbotapi.NewMessage(chatID, "В этой теме нет активной очереди. Создайте через /create.")
m.MessageThreadID = threadID
_, _ = s.bot.TgBot.Send(m)
return nil
case errors.Is(err, client.ErrPositionOccupied):
m := tgbotapi.NewMessage(chatID, "Позиция "+strconv.Itoa(pos)+" уже занята. Свободные: /positions")
m.MessageThreadID = threadID
_, _ = s.bot.TgBot.Send(m)
return nil
case errors.Is(err, client.ErrPositionTooBig):
m := tgbotapi.NewMessage(chatID, "Позиция должна быть в диапазоне 1..999.")
m.MessageThreadID = threadID
_, _ = s.bot.TgBot.Send(m)
return nil
case errors.Is(err, client.ErrAlreadyInQueue):
m := tgbotapi.NewMessage(chatID, "Вы уже в очереди. Сменить позицию: /join Иванов Иван #3")
m.MessageThreadID = threadID
_, _ = s.bot.TgBot.Send(m)
return nil
}
return s.wrapUserError(err, chatID, threadID)
case LeaveCommand:
return s.wrapUserError(s.bot.LeaveQueueFromCommand(chatID, userID, threadID), chatID, threadID)
case KickFirstCommand:
return s.wrapUserError(s.bot.KickFirstFromCommand(chatID, userID, threadID), chatID, threadID)
case KickCommand:
pos, ok := parsePositionArg(message.CommandArguments())
if !ok {
return s.bot.SendKickUsage(chatID, threadID)
}
err := s.bot.KickAtFromCommand(chatID, userID, threadID, pos)
if err == nil {
return nil
}
if errors.Is(err, client.ErrPositionEmpty) {
m := tgbotapi.NewMessage(chatID, "Позиция "+strconv.Itoa(pos)+" пуста.")
m.MessageThreadID = threadID
_, _ = s.bot.TgBot.Send(m)
return nil
}
return s.wrapUserError(err, chatID, threadID)
case SwitchCommand:
pos, ok := parsePositionArg(message.CommandArguments())
if !ok {
return s.bot.SendSwitchUsage(chatID, threadID)
}
err := s.bot.SwitchFromCommand(chatID, userID, threadID, pos)
if err == nil {
return nil
}
switch {
case errors.Is(err, client.ErrNotAdmin):
return s.wrapUserError(err, chatID, threadID)
case errors.Is(err, client.ErrNoQueue):
return s.wrapUserError(err, chatID, threadID)
case errors.Is(err, client.ErrNotInQueue):
m := tgbotapi.NewMessage(chatID, "Вас нет в очереди этой темы. Сначала /join.")
m.MessageThreadID = threadID
_, _ = s.bot.TgBot.Send(m)
return nil
case errors.Is(err, client.ErrPositionEmpty):
m := tgbotapi.NewMessage(chatID, "Позиция "+strconv.Itoa(pos)+" пуста — там никого нет.")
m.MessageThreadID = threadID
_, _ = s.bot.TgBot.Send(m)
return nil
case errors.Is(err, client.ErrSamePosition):
m := tgbotapi.NewMessage(chatID, "Вы уже на позиции "+strconv.Itoa(pos)+".")
m.MessageThreadID = threadID
_, _ = s.bot.TgBot.Send(m)
return nil
}
return s.wrapUserError(err, chatID, threadID)
case Swap12Command:
// Проверка админа внутри SwapFirstTwoFromCommand
err := s.bot.SwapFirstTwoFromCommand(chatID, userID, threadID)
if err == nil {
return nil
}
switch {
case errors.Is(err, client.ErrNotAdmin):
return s.wrapUserError(err, chatID, threadID)
case errors.Is(err, client.ErrNoQueue):
return s.wrapUserError(err, chatID, threadID)
case errors.Is(err, client.ErrPositionEmpty):
m := tgbotapi.NewMessage(chatID, "Нужны хотя бы 2 участника: первый и второй должны быть на местах.")
m.MessageThreadID = threadID
_, _ = s.bot.TgBot.Send(m)
return nil
}
return s.wrapUserError(err, chatID, threadID)
case CloseCommand:
return s.wrapUserError(s.bot.CloseQueueFromCommand(chatID, userID, threadID), chatID, threadID)
case PositionsCommand:
return s.wrapUserError(s.bot.SendFreePositions(chatID, threadID), chatID, threadID)
case DebugCommand:
return s.wrapUserError(s.bot.SendDebug(chatID, userID, threadID), chatID, threadID)
}
return nil
}

func (s BotServer) wrapUserError(err error, chatID int64, threadID int) error {
if err == nil {
return nil
}
var text string
switch {
case errors.Is(err, client.ErrNoQueue):
text = client.NoActiveQueueMessage
case errors.Is(err, client.ErrNotInQueue):
text = client.NotInQueueMessage
case errors.Is(err, client.ErrNotAdmin):
text = client.OnlyAdminMessage
default:
return err
}
m := tgbotapi.NewMessage(chatID, text)
m.MessageThreadID = threadID
_, _ = s.bot.TgBot.Send(m)
return nil
}