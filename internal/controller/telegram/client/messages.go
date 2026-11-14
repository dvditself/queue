package client

import (
"fmt"

tgbotapi "github.com/OvyFlash/telegram-bot-api"

"QueueBot/internal/entity"
)

const (
QueueDescription     = "В очереди состоят:"
EndedQueue           = "Участники закончились, значит и очередь тоже. Что делаем дальше?"
FinishedQueue        = "'Очередь' окончена 🎉"
OnlyAdminMessage     = "Только администратор чата может это сделать"
NoActiveQueueMessage = "В этой теме нет активной очереди"
NotInQueueMessage    = "Вас нет в очереди этой темы"
)

func getMessageContentBeforeStart(title string, users []entity.User) string {
return fmt.Sprintf("*%s*\n%s\n%s", title, QueueDescription, cutStringByLines(entity.ListToString(users), 26))
}

func getMessageContentAfterStart(title string, users []entity.User, currentPersonIndex int) string {
return fmt.Sprintf("*%s*\n%s\n%s", title, QueueDescription,
cutStringByLinesWithCurrent(entity.ListToStringWithCurrent(users, currentPersonIndex), 13, currentPersonIndex))
}

func GetQueueMessageContent(description string) tgbotapi.InputTextMessageContent {
return tgbotapi.InputTextMessageContent{
Text:      getMessageContentBeforeStart(description, nil),
ParseMode: tgbotapi.ModeMarkdown,
}
}

func GetQueueMessage(chatID int64, messageID int, users []entity.User, description string) tgbotapi.EditMessageTextConfig {
kb := GetBeforeStartKeyboard()
return tgbotapi.EditMessageTextConfig{
BaseEdit:  tgbotapi.BaseEdit{ChatID: chatID, MessageID: messageID, ReplyMarkup: &kb},
Text:      getMessageContentBeforeStart(description, users),
ParseMode: tgbotapi.ModeMarkdown,
}
}

func GetUpdatedQueueMessage(chatID int64, messageID int, description string, users []entity.User) tgbotapi.EditMessageTextConfig {
kb := GetBeforeStartKeyboard()
return tgbotapi.EditMessageTextConfig{
BaseEdit:  tgbotapi.BaseEdit{ChatID: chatID, MessageID: messageID, ReplyMarkup: &kb},
Text:      getMessageContentBeforeStart(description, users),
ParseMode: tgbotapi.ModeMarkdown,
}
}

func GetQueueAfterStartMessage(chatID int64, messageID int, description string, users []entity.User, currentPersonIndex int) tgbotapi.EditMessageTextConfig {
kb := GetAfterStartKeyboard()
return tgbotapi.EditMessageTextConfig{
BaseEdit:  tgbotapi.BaseEdit{ChatID: chatID, MessageID: messageID, ReplyMarkup: &kb},
Text:      getMessageContentAfterStart(description, users, currentPersonIndex),
ParseMode: tgbotapi.ModeMarkdown,
}
}

func GetEndQueueMessage(chatID int64, messageID int) tgbotapi.EditMessageTextConfig {
kb := GetEndedQueueKeyboard()
return tgbotapi.EditMessageTextConfig{
BaseEdit:  tgbotapi.BaseEdit{ChatID: chatID, MessageID: messageID, ReplyMarkup: &kb},
Text:      EndedQueue,
}
}

func GetFinishedMessage(chatID int64, messageID int) tgbotapi.EditMessageTextConfig {
return tgbotapi.EditMessageTextConfig{
BaseEdit: tgbotapi.BaseEdit{ChatID: chatID, MessageID: messageID},
Text:     FinishedQueue,
}
}