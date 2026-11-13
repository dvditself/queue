package telegram

import (
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
)

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
name := strings.TrimSpace(message.CommandArguments())
if name == "" {
return s.bot.SendJoinUsage(chatID)
}
return s.wrapUserError(s.bot.JoinQueueFromCommand(chatID, userID, name), chatID)
case LeaveCommand:
return s.wrapUserError(s.bot.LeaveQueueFromCommand(chatID, userID), chatID)
case KickFirstCommand:
return s.wrapUserError(s.bot.KickFirstFromCommand(chatID, userID), chatID)
case CloseCommand:
return s.wrapUserError(s.bot.CloseQueueFromCommand(chatID, userID), chatID)
}
return nil
}

func (s BotServer) wrapUserError(err error, chatID int64) error {
if err == nil {
return nil
}
var text string
switch {
case err == client.ErrNoQueue:
text = client.NoActiveQueueMessage
case err == client.ErrNotInQueue:
text = client.NotInQueueMessage
case err == client.ErrNotAdmin:
text = client.OnlyAdminMessage
default:
return err
}
_, _ = s.bot.TgBot.Send(tgbotapi.NewMessage(chatID, text))
return nil
}