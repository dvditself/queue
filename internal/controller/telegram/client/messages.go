package client

import (
"fmt"

tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

"QueueBot/internal/entity"
)

const (
QueueDescription = "В очереди состоят:"
EndedQueue       = "Участники закончились, значит и очередь тоже. Что делаем дальше?"
FinishedQueue    = "'Очередь' окончена 🎉"
)

func getMessageContentBeforeStart(title string, users []entity.User) string {
return fmt.Sprintf(
"*%s*\n%s\n%s",
title,
QueueDescription,
cutStringByLines(entity.ListToString(users), 26),
)
}

func getMessageContentAfterStart(title string, users []entity.User, currentPersonIndex int) string {
return fmt.Sprintf(
"*%s*\n%s\n%s",
title,
QueueDescription,
cutStringByLinesWithCurrent(entity.ListToStringWithCurrent(users, currentPersonIndex), 13, currentPersonIndex),
)
}

func GetQueueMessageContent(description string) tgbotapi.InputTextMessageContent {
return tgbotapi.InputTextMessageContent{
Text:      getMessageContentBeforeStart(description, nil),
ParseMode: tgbotapi.ModeMarkdown,
}
}

func GetQueueMessage(chatID int64, messageID int, users []entity.User, description string) tgbotapi.EditMessageTextConfig {
keyboard := GetBeforeStartKeyboard()
return tgbotapi.EditMessageTextConfig{
BaseEdit: tgbotapi.BaseEdit{
ChatID:      chatID,
MessageID:   messageID,
ReplyMarkup: &keyboard,
},
Text:      getMessageContentBeforeStart(description, users),
ParseMode: tgbotapi.ModeMarkdown,
}
}

func GetUpdatedQueueMessage(chatID int64, messageID int, description string, users []entity.User) tgbotapi.EditMessageTextConfig {
keyboard := GetBeforeStartKeyboard()
return tgbotapi.EditMessageTextConfig{
BaseEdit: tgbotapi.BaseEdit{
ChatID:      chatID,
MessageID:   messageID,
ReplyMarkup: &keyboard,
},
Text:      getMessageContentBeforeStart(description, users),
ParseMode: tgbotapi.ModeMarkdown,
}
}

func GetQueueAfterStartMessage(chatID int64, messageID int, description string, users []entity.User, currentPersonIndex int) tgbotapi.EditMessageTextConfig {
keyboard := GetAfterStartKeyboard()
return tgbotapi.EditMessageTextConfig{
BaseEdit: tgbotapi.BaseEdit{
ChatID:      chatID,
MessageID:   messageID,
ReplyMarkup: &keyboard,
},
Text:      getMessageContentAfterStart(description, users, currentPersonIndex),
ParseMode: tgbotapi.ModeMarkdown,
}
}

func GetEndQueueMessage(chatID int64, messageID int) tgbotapi.EditMessageTextConfig {
keyboard := GetEndedQueueKeyboard()
return tgbotapi.EditMessageTextConfig{
BaseEdit: tgbotapi.BaseEdit{
ChatID:      chatID,
MessageID:   messageID,
ReplyMarkup: &keyboard,
},
Text: EndedQueue,
}
}

func GetFinishedMessage(chatID int64, messageID int) tgbotapi.EditMessageTextConfig {
return tgbotapi.EditMessageTextConfig{
BaseEdit: tgbotapi.BaseEdit{
ChatID:    chatID,
MessageID: messageID,
},
Text: FinishedQueue,
}
}
