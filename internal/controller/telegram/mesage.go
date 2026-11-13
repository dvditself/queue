package telegram

import (
"fmt"
"strings"

tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
StartCommand  = "start"
CreateCommand = "create"
HelpCommand   = "help"
)

func (s BotServer) HandleMessage(message *tgbotapi.Message) error {
if message.IsCommand() {
switch message.Command() {
case StartCommand:
if err := s.bot.SendHelloMessage(message); err != nil {
return fmt.Errorf("sendHelloMessage error occurred: %w", err)
}
case CreateCommand:
description := strings.TrimSpace(message.CommandArguments())
if description == "" {
if err := s.bot.SendCreateUsage(message.Chat.ID); err != nil {
return fmt.Errorf("sendCreateUsage error occurred: %w", err)
}
return nil
}
if err := s.bot.CreateQueueFromCommand(message.Chat.ID, description); err != nil {
return fmt.Errorf("createQueueFromCommand error occurred: %w", err)
}
case HelpCommand:
if err := s.bot.SendHelpMessage(message.Chat.ID); err != nil {
return fmt.Errorf("sendHelpMessage error occurred: %w", err)
}
}
return nil
}

return nil
}
