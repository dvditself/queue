package client

import (
tgbotapi "github.com/OvyFlash/telegram-bot-api"
)

const (
StartQueueButton        = "Старт в порядке очереди"
StartQueueShuffleButton = "Старт в случайном порядке"
CloseQueueButton        = "Закрыть очередь (только админ)"
NextButton              = "Следующий"
GoToMenuButton          = "Перейти в меню"
FinishQueueButton       = "Закончить"
)

const (
StartQueueData        = "start_queue"
StartQueueShuffleData = "start_queue_shuffle"
NextData              = "next_user"
GoToMenuData          = "go_to_menu"
FinishQueueData       = "finish_queue"
)

func GetBeforeStartKeyboard() tgbotapi.InlineKeyboardMarkup {
return tgbotapi.NewInlineKeyboardMarkup(
tgbotapi.NewInlineKeyboardRow(startQueueButton()),
tgbotapi.NewInlineKeyboardRow(startQueueShuffleButton()),
tgbotapi.NewInlineKeyboardRow(closeQueueButton()),
)
}

func GetAfterStartKeyboard() tgbotapi.InlineKeyboardMarkup {
return tgbotapi.NewInlineKeyboardMarkup(
tgbotapi.NewInlineKeyboardRow(nextButton()),
tgbotapi.NewInlineKeyboardRow(closeQueueButton()),
)
}

func GetEndedQueueKeyboard() tgbotapi.InlineKeyboardMarkup {
return tgbotapi.NewInlineKeyboardMarkup(
tgbotapi.NewInlineKeyboardRow(goToMenuButton()),
tgbotapi.NewInlineKeyboardRow(endQueueButton()),
)
}

func startQueueButton() tgbotapi.InlineKeyboardButton {
return tgbotapi.NewInlineKeyboardButtonData(StartQueueButton, StartQueueData)
}
func startQueueShuffleButton() tgbotapi.InlineKeyboardButton {
return tgbotapi.NewInlineKeyboardButtonData(StartQueueShuffleButton, StartQueueShuffleData)
}
func nextButton() tgbotapi.InlineKeyboardButton {
return tgbotapi.NewInlineKeyboardButtonData(NextButton, NextData)
}
func goToMenuButton() tgbotapi.InlineKeyboardButton {
return tgbotapi.NewInlineKeyboardButtonData(GoToMenuButton, GoToMenuData)
}
func endQueueButton() tgbotapi.InlineKeyboardButton {
return tgbotapi.NewInlineKeyboardButtonData(FinishQueueButton, FinishQueueData)
}
func closeQueueButton() tgbotapi.InlineKeyboardButton {
return tgbotapi.NewInlineKeyboardButtonData(CloseQueueButton, FinishQueueData)
}