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
PositionsCommand = "positions"
DebugCommand     = "debug"
)

// parseJoinArgs:
//   "/join Иванов Иван"        -> name="Иванов Иван", pos=0
//   "/join Иванов Иван 3"      -> name="Иванов Иван", pos=3
//   "/join Иванов Иван #3"     -> name="Иванов Иван", pos=3
//   "/join Иванов Иван 0"      -> name="Иванов Иван 0", pos=0 (невалидная позиция - уйдёт в имя)
//   "/join #3"                 -> name="", pos=0 (сработает usage)
func parseJoinArgs(args string) (fullName string, position int) {
args = strings.TrimSpace(args)
if args == "" {
return "", 0
}
parts := strings.Fields(args)
if len(parts) >= 2 {
last := parts[len(parts)-1]
// вариант с #: "#3"
if strings.HasPrefix(last, "#") && len(last) > 1 {
if p, err := strconv.Atoi(last[1:]); err == nil && p > 0 {
return strings.Join(parts[:len(parts)-1], " "), p
}
}
// вариант без #: "3"
if p, err := strconv.Atoi(last); err == nil && p > 0 {
return strings.Join(parts[:len(parts)-1], " "), p
}
}
return args, 0
}

func (s BotServer) HandleMessage(message *tgbotapi.Message) error {
if !message.IsCommand() {
return nil
}
chatID := message.Chat.ID
userID := message.From.ID

switch message.Command() {
case StartCommand:
return s.bot.SendHelloMessage(message)
case HelpCommand:
return s.bot.SendHelpMessage(chatID)
case CreateCommand:
desc := strings.TrimSpace(message.CommandArguments())
if desc == "" {
return s.bot.SendCreateUsage(chatID)
}
return s.bot.CreateQueueFromCommand(chatID, desc)
case JoinCommand:
name, pos := parseJoinArgs(message.CommandArguments())
if name == "" {
return s.bot.SendJoinUsage(chatID)
}
err := s.bot.JoinQueueFromCommand(chatID, userID, name, pos)
if err == nil {
return nil
}
switch {
case errors.Is(err, client.ErrPositionOccupied):
_, _ = s.bot.TgBot.Send(tgbotapi.NewMessage(chatID,
"Позиция "+strconv.Itoa(pos)+" уже занята. Свободные: /positions"))
return nil
case errors.Is(err, client.ErrPositionTooBig):
_, _ = s.bot.TgBot.Send(tgbotapi.NewMessage(chatID,
"Позиция должна быть в диапазоне 1..999. Свободные: /positions"))
return nil
case errors.Is(err, client.ErrAlreadyInQueue):
_, _ = s.bot.TgBot.Send(tgbotapi.NewMessage(chatID,
"Вы уже в очереди. Чтобы сменить позицию — укажите номер: /join Иванов Иван 3 или /join Иванов Иван #3"))
return nil
}
return s.wrapUserError(err, chatID)
case LeaveCommand:
return s.wrapUserError(s.bot.LeaveQueueFromCommand(chatID, userID), chatID)
case KickFirstCommand:
return s.wrapUserError(s.bot.KickFirstFromCommand(chatID, userID), chatID)
case CloseCommand:
return s.wrapUserError(s.bot.CloseQueueFromCommand(chatID, userID), chatID)
case PositionsCommand:
return s.wrapUserError(s.bot.SendFreePositions(chatID), chatID)
case DebugCommand:
return s.wrapUserError(s.bot.SendDebug(chatID, userID), chatID)
}
return nil
}

func (s BotServer) wrapUserError(err error, chatID int64) error {
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
_, _ = s.bot.TgBot.Send(tgbotapi.NewMessage(chatID, text))
return nil
}