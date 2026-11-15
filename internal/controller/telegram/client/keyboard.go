package client

import (
tgbotapi "github.com/OvyFlash/telegram-bot-api"
)

const (
StartQueueButton  = "Старт в порядке очереди"
KickFirstButton   = "Убрать первого (админ)"
CloseQueueButton  = "Закрыть очередь (админ)"
NextButton        = "Следующий"
GoToMenuButton    = "Перейти в меню"
FinishQueueButton = "Закончить"

ConfirmYesButton = "Да, закрыть"
ConfirmNoButton  = "Отмена"
)

const (
StartQueueData   = "start_queue"
KickFirstData    = "kick_first"
NextData         = "next_user"
GoToMenuData     = "go_to_menu"
FinishQueueData  = "finish_queue"
ConfirmCloseData = "close_confirm_yes" // использование: ConfirmCloseData + ":" + key
CancelCloseData  = "close_confirm_no"
)

func GetBeforeStartKeyboard() tgbotapi.InlineKeyboardMarkup {
return tgbotapi.NewInlineKeyboardMarkup(
tgbotapi.NewInlineKeyboardRow(startQueueButton()),
tgbotapi.NewInlineKeyboardRow(kickFirstButton()),
tgbotapi.NewInlineKeyboardRow(closeQueueButton()),
)
}

func GetAfterStartKeyboard() tgbotapi.InlineKeyboardMarkup {
return tgbotapi.NewInlineKeyboardMarkup(
tgbotapi.NewInlineKeyboardRow(nextButton()),
tgbotapi.NewInlineKeyboardRow(kickFirstButton()),
tgbotapi.NewInlineKeyboardRow(closeQueueButton()),
)
}

func GetEndedQueueKeyboard() tgbotapi.InlineKeyboardMarkup {
return tgbotapi.NewInlineKeyboardMarkup(
tgbotapi.NewInlineKeyboardRow(goToMenuButton()),
tgbotapi.NewInlineKeyboardRow(endQueueButton()),
)
}

// GetCloseConfirmKeyboard — клавиатура подтверждения закрытия конкретной очереди.
// key кодируется прямо в callback data, чтобы при нажатии "Да" закрыть ИМЕННО эту очередь,
// а не последнюю в чате (race condition при создании новой, пока висело подтверждение).
func GetCloseConfirmKeyboard(key string) tgbotapi.InlineKeyboardMarkup {
return tgbotapi.NewInlineKeyboardMarkup(
tgbotapi.NewInlineKeyboardRow(
tgbotapi.NewInlineKeyboardButtonData(ConfirmYesButton, ConfirmCloseData+":"+key),
tgbotapi.NewInlineKeyboardButtonData(ConfirmNoButton, CancelCloseData),
),
)
}

func startQueueButton() tgbotapi.InlineKeyboardButton {
return tgbotapi.NewInlineKeyboardButtonData(StartQueueButton, StartQueueData)
}
func kickFirstButton() tgbotapi.InlineKeyboardButton {
return tgbotapi.NewInlineKeyboardButtonData(KickFirstButton, KickFirstData)
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