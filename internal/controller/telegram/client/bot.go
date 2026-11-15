package client

import (
"context"
"errors"
"fmt"
"log/slog"
"strconv"
"strings"

tgbotapi "github.com/OvyFlash/telegram-bot-api"

"QueueBot/internal/entity"
"QueueBot/internal/usecase"
storage "QueueBot/internal/usecase/storage/sqlite"
)

var (
ErrNotAdmin         = errors.New("only admin can do this")
ErrNoQueue          = errors.New("no active queue in this thread")
ErrNotInQueue       = storage.ErrNotInQueue
ErrQueueAlreadyOpen = errors.New("queue already open in this thread")
ErrPositionOccupied = storage.ErrPositionOccupied
ErrPositionTooBig   = storage.ErrPositionTooBig
ErrAlreadyInQueue   = storage.ErrAlreadyInQueue
ErrPositionEmpty    = storage.ErrPositionEmpty
ErrSamePosition     = storage.ErrSamePosition
)

const HelloMessage = `Привет! Я бот для создания очередей.

Как пользоваться:
• /create <название> — создать очередь в текущей теме
• /join <Фамилия Имя> — встать в начало
• /join <Фамилия Имя> #N — встать на позицию N
• /leave — покинуть очередь
• /positions — свободные позиции
• /switch N — поменяться с позицией N (админ)
• /kick N — выгнать с позиции N (админ)
• /kickfirst — убрать первого (админ)
• /close — закрыть очередь (админ, с подтверждением)
• /help — эта справка

В каждой теме — своя очередь, только одна открытая одновременно.`

const CreateUsageMessage = `Использование: /create <название>
Пример: /create ЛР3`

const JoinUsageMessage = `Использование:
/join Фамилия Имя — встать в начало
/join Фамилия Имя #N — встать на позицию N

Пример: /join Иванов Иван #3`

const SwitchUsageMessage = `Использование: /switch N — поменяться местами с человеком на позиции N
Пример: /switch 2`

const KickUsageMessage = `Использование: /kick N — выгнать человека с позиции N
Пример: /kick 3`

const HelpMessage = `Команды:
/create <название> — создать очередь в текущей теме
/join <Фамилия Имя> — встать в начало
/join <Фамилия Имя> #N — встать на позицию N
/leave — покинуть очередь
/positions — свободные позиции
/switch N — поменяться местами с позицией N (админ)
/kick N — выгнать с позиции N (админ)
/kickfirst — убрать первого (админ)
/close — закрыть очередь (админ, потребует подтверждения)
/debug — диагностика (админ)

/join работает ТОЛЬКО в той теме, где уже открыта очередь.`

const CloseConfirmMessage = "Вы уверены, что хотите закрыть очередь? Это действие нельзя отменить. Нажмите «Да, закрыть» для подтверждения."
const CloseDoneMessage = "'Очередь' закрыта 🎉"
const CloseCanceledMessage = "Закрытие отменено."
const CloseAlreadyGoneMessage = "Очередь уже закрыта."

type TelegramBot struct {
TgBot *tgbotapi.BotAPI
u     usecase.Bot
}

func NewTelegramBot(tgBot *tgbotapi.BotAPI, u usecase.Bot) *TelegramBot {
return &TelegramBot{TgBot: tgBot, u: u}
}

func threadIDForSend(threadID int) int {
if threadID == 1 {
return 0
}
return threadID
}

func queueKey(chatID int64, threadID int, messageID int) string {
return fmt.Sprintf("%d:%d:%d", chatID, threadID, messageID)
}

func parseQueueKey(key string) (int64, int, int, error) {
parts := strings.SplitN(key, ":", 3)
if len(parts) != 3 {
return 0, 0, 0, fmt.Errorf("invalid key: %s", key)
}
cid, err := strconv.ParseInt(parts[0], 10, 64)
if err != nil {
return 0, 0, 0, err
}
tid, err := strconv.Atoi(parts[1])
if err != nil {
return 0, 0, 0, err
}
mid, err := strconv.Atoi(parts[2])
if err != nil {
return 0, 0, 0, err
}
return cid, tid, mid, nil
}

func keyFromCallback(cb *tgbotapi.CallbackQuery) string {
tid := threadIDForSend(cb.Message.MessageThreadID)
return queueKey(cb.Message.Chat.ID, tid, cb.Message.MessageID)
}

func (b TelegramBot) SendHelloMessage(m *tgbotapi.Message) error {
msg := tgbotapi.NewMessage(m.Chat.ID, HelloMessage)
msg.MessageThreadID = threadIDForSend(m.MessageThreadID)
_, err := b.TgBot.Send(msg)
return err
}
func (b TelegramBot) SendCreateUsage(chatID int64, threadID int) error {
msg := tgbotapi.NewMessage(chatID, CreateUsageMessage)
msg.MessageThreadID = threadIDForSend(threadID)
_, err := b.TgBot.Send(msg)
return err
}
func (b TelegramBot) SendHelpMessage(chatID int64, threadID int) error {
msg := tgbotapi.NewMessage(chatID, HelpMessage)
msg.MessageThreadID = threadIDForSend(threadID)
_, err := b.TgBot.Send(msg)
return err
}
func (b TelegramBot) SendJoinUsage(chatID int64, threadID int) error {
msg := tgbotapi.NewMessage(chatID, JoinUsageMessage)
msg.MessageThreadID = threadIDForSend(threadID)
_, err := b.TgBot.Send(msg)
return err
}
func (b TelegramBot) SendSwitchUsage(chatID int64, threadID int) error {
msg := tgbotapi.NewMessage(chatID, SwitchUsageMessage)
msg.MessageThreadID = threadIDForSend(threadID)
_, err := b.TgBot.Send(msg)
return err
}
func (b TelegramBot) SendKickUsage(chatID int64, threadID int) error {
msg := tgbotapi.NewMessage(chatID, KickUsageMessage)
msg.MessageThreadID = threadIDForSend(threadID)
_, err := b.TgBot.Send(msg)
return err
}

// sendCloseConfirm — отправляет сообщение с кнопками Да/Отмена для конкретной очереди key.
func (b TelegramBot) sendCloseConfirm(chatID int64, threadID int, key string) error {
msg := tgbotapi.NewMessage(chatID, CloseConfirmMessage)
msg.MessageThreadID = threadIDForSend(threadID)
kb := GetCloseConfirmKeyboard(key)
msg.ReplyMarkup = kb
_, err := b.TgBot.Send(msg)
return err
}

func (b TelegramBot) SendFreePositions(chatID int64, threadID int) error {
ctx := context.Background()
key, err := b.u.GetLastQueueKeyInChatThread(ctx, chatID, threadID)
if err != nil {
return err
}
if key == "" {
return ErrNoQueue
}
free, err := b.u.GetFreePositions(ctx, key)
if err != nil {
return err
}
var text string
if len(free) == 0 {
text = "Свободных позиций нет."
} else {
parts := make([]string, len(free))
for i, p := range free {
parts[i] = strconv.Itoa(p)
}
text = "Свободные позиции: " + strings.Join(parts, ", ")
}
msg := tgbotapi.NewMessage(chatID, text)
msg.MessageThreadID = threadIDForSend(threadID)
_, err = b.TgBot.Send(msg)
return err
}

func (b TelegramBot) SendDebug(chatID, userID int64, threadID int) error {
ctx := context.Background()
isAdmin, err := b.IsChatAdmin(ctx, chatID, userID)
if err != nil {
return err
}
if !isAdmin {
return ErrNotAdmin
}
sb := strings.Builder{}
sb.WriteString(fmt.Sprintf("chatID=%d threadID=%d\n", chatID, threadID))
key, err := b.u.GetLastQueueKeyInChatThread(ctx, chatID, threadID)
if err != nil {
return err
}
if key == "" {
sb.WriteString("В этой теме активной очереди нет.\n")
} else {
sb.WriteString(fmt.Sprintf("Ключ текущей темы: %s\n", key))
rows, err := b.u.DebugPositions(ctx, key)
if err == nil {
sb.WriteString("Участники (order_number):\n")
if len(rows) == 0 {
sb.WriteString("(пусто)\n")
}
for _, r := range rows {
del := ""
if r.Deleted == 1 {
del = " [удалён]"
}
sb.WriteString(fmt.Sprintf("  #%d %s (id=%d)%s\n", r.Order, r.UserName, r.UserID, del))
}
}
}
allKeys, err := b.u.DebugChatQueues(ctx, chatID)
if err == nil {
sb.WriteString("Все ключи очередей в чате:\n")
if len(allKeys) == 0 {
sb.WriteString("(нет)\n")
}
for _, k := range allKeys {
sb.WriteString(fmt.Sprintf("  %s\n", k))
}
}
msg := tgbotapi.NewMessage(chatID, sb.String())
msg.MessageThreadID = threadIDForSend(threadID)
_, err = b.TgBot.Send(msg)
return err
}

func (b TelegramBot) CreateQueueFromCommand(chatID int64, threadID int, description string) error {
ctx := context.Background()
existing, err := b.u.GetLastQueueKeyInChatThread(ctx, chatID, threadID)
if err != nil {
return fmt.Errorf("couldn't check existing: %w", err)
}
if existing != "" {
return ErrQueueAlreadyOpen
}
msg := tgbotapi.NewMessage(chatID, getMessageContentBeforeStart(description, nil))
msg.ParseMode = tgbotapi.ModeMarkdown
msg.MessageThreadID = threadIDForSend(threadID)
kb := GetBeforeStartKeyboard()
msg.ReplyMarkup = kb
sent, err := b.TgBot.Send(msg)
if err != nil {
return fmt.Errorf("couldn't send queue message: %w", err)
}
key := queueKey(chatID, threadID, sent.MessageID)
if err := b.u.CreateQueue(ctx, key, description); err != nil {
return fmt.Errorf("couldn't create queue: %w", err)
}
slog.Info("Queue created", "key", key, "description", description, "threadID", threadID)
return nil
}

func (b TelegramBot) JoinQueueFromCommand(chatID, userID int64, threadID int, fullName string, position int) error {
ctx := context.Background()
key, err := b.u.GetLastQueueKeyInChatThread(ctx, chatID, threadID)
if err != nil {
return fmt.Errorf("couldn't find queue: %w", err)
}
if key == "" {
return ErrNoQueue
}
if err := b.u.SaveUserName(ctx, userID, fullName); err != nil {
return err
}
if err := b.u.JoinQueue(ctx, key, entity.NewFullName(userID, fullName), position); err != nil {
return err
}
if err := b.refreshQueueMessage(ctx, key); err != nil {
return err
}
queue, err := b.u.GetQueue(ctx, key)
if err == nil {
myPos := 0
for i, u := range queue.Users {
if u.ID == userID {
myPos = u.Position
if myPos <= 0 {
myPos = i + 1
}
break
}
}
if myPos > 0 {
cm := tgbotapi.NewMessage(chatID, fmt.Sprintf("%s записан на позицию %d", fullName, myPos))
cm.MessageThreadID = threadIDForSend(threadID)
_, _ = b.TgBot.Send(cm)
}
}
return nil
}

func (b TelegramBot) LeaveQueueFromCommand(chatID, userID int64, threadID int) error {
ctx := context.Background()
key, err := b.u.GetLastQueueKeyInChatThread(ctx, chatID, threadID)
if err != nil {
return err
}
if key == "" {
return ErrNoQueue
}
queue, err := b.u.GetQueue(ctx, key)
if err != nil {
return err
}
found := false
for _, u := range queue.Users {
if u.ID == userID {
found = true
break
}
}
if !found {
return ErrNotInQueue
}
if err := b.u.LeaveQueue(ctx, key, userID); err != nil {
return err
}
return b.refreshQueueMessage(ctx, key)
}

func (b TelegramBot) KickFirstFromCommand(chatID, userID int64, threadID int) error {
ctx := context.Background()
isAdmin, err := b.IsChatAdmin(ctx, chatID, userID)
if err != nil {
return err
}
if !isAdmin {
return ErrNotAdmin
}
key, err := b.u.GetLastQueueKeyInChatThread(ctx, chatID, threadID)
if err != nil {
return err
}
if key == "" {
return ErrNoQueue
}
if err := b.u.KickFirstUser(ctx, key); err != nil {
return err
}
return b.refreshQueueMessage(ctx, key)
}

func (b TelegramBot) KickAtFromCommand(chatID, userID int64, threadID int, position int) error {
ctx := context.Background()
isAdmin, err := b.IsChatAdmin(ctx, chatID, userID)
if err != nil {
return err
}
if !isAdmin {
return ErrNotAdmin
}
key, err := b.u.GetLastQueueKeyInChatThread(ctx, chatID, threadID)
if err != nil {
return err
}
if key == "" {
return ErrNoQueue
}
if err := b.u.KickAtPosition(ctx, key, position); err != nil {
return err
}
return b.refreshQueueMessage(ctx, key)
}

func (b TelegramBot) SwitchFromCommand(chatID, userID int64, threadID int, position int) error {
ctx := context.Background()
isAdmin, err := b.IsChatAdmin(ctx, chatID, userID)
if err != nil {
return err
}
if !isAdmin {
return ErrNotAdmin
}
key, err := b.u.GetLastQueueKeyInChatThread(ctx, chatID, threadID)
if err != nil {
return err
}
if key == "" {
return ErrNoQueue
}
if err := b.u.SwitchPositions(ctx, key, userID, position); err != nil {
return err
}
return b.refreshQueueMessage(ctx, key)
}

// CloseQueueFromCommand — /close: права, наличие очереди, отправка подтверждения.
func (b TelegramBot) CloseQueueFromCommand(chatID, userID int64, threadID int) error {
ctx := context.Background()
isAdmin, err := b.IsChatAdmin(ctx, chatID, userID)
if err != nil {
return err
}
if !isAdmin {
return ErrNotAdmin
}
key, err := b.u.GetLastQueueKeyInChatThread(ctx, chatID, threadID)
if err != nil {
return err
}
if key == "" {
return ErrNoQueue
}
return b.sendCloseConfirm(chatID, threadID, key)
}

// RequestCloseConfirm — нажатие "Закрыть" под сообщением: права + отправка подтверждения с key.
func (b TelegramBot) RequestCloseConfirm(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
chatID := cb.Message.Chat.ID
userID := cb.From.ID
tid := threadIDForSend(cb.Message.MessageThreadID)

isAdmin, err := b.IsChatAdmin(ctx, chatID, userID)
if err != nil {
return err
}
if !isAdmin {
return ErrNotAdmin
}
key, err := b.u.GetLastQueueKeyInChatThread(ctx, chatID, tid)
if err != nil {
return err
}
if key == "" {
return ErrNoQueue
}
return b.sendCloseConfirm(chatID, tid, key)
}

// ConfirmClose — "Да, закрыть" для конкретной очереди key (из callback data).
// Защита от race: даже если в чате уже создали новую очередь, закроется именно эта.
func (b TelegramBot) ConfirmClose(ctx context.Context, cb *tgbotapi.CallbackQuery, key string) error {
chatID := cb.Message.Chat.ID
userID := cb.From.ID

isAdmin, err := b.IsChatAdmin(ctx, chatID, userID)
if err != nil {
return err
}
if !isAdmin {
return ErrNotAdmin
}

keyChatID, _, msgID, err := parseQueueKey(key)
if err != nil {
return fmt.Errorf("invalid key in callback: %w", err)
}

// проверяем, что очередь ещё существует
if _, err := b.u.GetQueue(ctx, key); err != nil {
// уже закрыта или удалена
edit := tgbotapi.NewEditMessageText(chatID, cb.Message.MessageID, CloseAlreadyGoneMessage)
_, _ = b.TgBot.Request(edit)
return nil
}

// редактируем исходное сообщение очереди
if _, err := b.TgBot.Request(GetFinishedMessage(keyChatID, msgID)); err != nil {
slog.Warn("Couldn't edit original queue message", "error", err, "key", key)
}
// удаляем из БД
if err := b.u.FinishQueue(ctx, key); err != nil {
return err
}
// редактируем сообщение-подтверждение
edit := tgbotapi.NewEditMessageText(chatID, cb.Message.MessageID, CloseDoneMessage)
_, _ = b.TgBot.Request(edit)
slog.Info("Queue closed via confirm", "key", key, "by", userID)
return nil
}

// CancelClose — "Отмена": помечаем сообщение, ничего не закрываем.
func (b TelegramBot) CancelClose(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
edit := tgbotapi.NewEditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, CloseCanceledMessage)
_, err := b.TgBot.Request(edit)
return err
}

func (b TelegramBot) refreshQueueMessage(ctx context.Context, key string) error {
chatID, _, msgID, err := parseQueueKey(key)
if err != nil {
return err
}
queue, err := b.u.GetQueue(ctx, key)
if err != nil {
return err
}
_, err = b.TgBot.Request(GetUpdatedQueueMessage(chatID, msgID, queue.Description, queue.Users))
return err
}

func (b TelegramBot) IsChatAdmin(ctx context.Context, chatID, userID int64) (bool, error) {
member, err := b.TgBot.GetChatMember(tgbotapi.GetChatMemberConfig{
ChatConfigWithUser: tgbotapi.ChatConfigWithUser{ChatID: chatID, UserID: userID},
})
if err != nil {
return false, err
}
return member.Status == "administrator" || member.Status == "creator", nil
}

func (b TelegramBot) Start(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
key := keyFromCallback(cb)
if err := b.u.StartQueue(ctx, key); err != nil {
return err
}
return b.sendStatus(ctx, cb)
}

func (b TelegramBot) Next(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
key := keyFromCallback(cb)
if err := b.u.SetNextPersonToQueue(ctx, key); err != nil {
return err
}
return b.sendStatus(ctx, cb)
}

func (b TelegramBot) KickFirst(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
key := keyFromCallback(cb)
if err := b.u.KickFirstUser(ctx, key); err != nil {
return err
}
return b.refreshQueueMessage(ctx, key)
}

func (b TelegramBot) GoToMenu(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
key := keyFromCallback(cb)
queue, err := b.u.GetQueue(ctx, key)
if err != nil {
return err
}
_, err = b.TgBot.Request(GetQueueMessage(cb.Message.Chat.ID, cb.Message.MessageID, queue.Users, queue.Description))
return err
}

func (b TelegramBot) sendStatus(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
key := keyFromCallback(cb)
queue, err := b.u.GetQueue(ctx, key)
if err != nil {
return err
}
var msg tgbotapi.EditMessageTextConfig
if queue.CurrentPersonIdx >= len(queue.Users) {
msg = GetEndQueueMessage(cb.Message.Chat.ID, cb.Message.MessageID)
} else {
msg = GetQueueAfterStartMessage(cb.Message.Chat.ID, cb.Message.MessageID, queue.Description, queue.Users, queue.CurrentPersonIdx)
}
_, err = b.TgBot.Request(msg)
return err
}