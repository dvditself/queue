package telegram

import (
"fmt"

tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const StartCommand = "start"

func (s BotServer) HandleMessage(message *tgbotapi.Message) error {
if message.IsCommand() {
if message.Command() == StartCommand {
if err := s.bot.SendHelloMessage(message); err != nil {
return fmt.Errorf("sendHelloMessage error occurred: %w", err)
}
}
return nil
}

if err := s.bot.SendForwardMessageButton(message); err != nil {
return fmt.Errorf("sendMessageToCreateMessage error occurred: %w", err)
}

return nil
}
