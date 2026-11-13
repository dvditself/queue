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
ErrNoQueue          = errors.New("no active queue in this chat")
ErrNotInQueue       = errors.New("user is not in queue")
ErrPositionOccupied = storage.ErrPositionOccupied
ErrPositionTooBig   = storage.ErrPositionTooBig
ErrAlreadyInQueue   = storage.ErrAlreadyInQueue
)

const HelloMessage = `Привет! Я бот для создания очередей.

Как пользоваться:
• /create <название> — создать очередь (например, /create ЛР3)
• /join <Фамилия Имя> — встать в конец
• /join <Фамилия Имя> <N> — на позицию N
• /join <Фамилия Имя> #N — то же, с решёткой
• /leave — выйти из очереди
• /positions — какие позиции свободны
• /kickfirst — убрать первого (только админ)
• /close — закрыть очередь (только админ)
• /debug — реальные позиции (только админ)
• /help — эта справка`

const CreateUsageMessage = `Использование: /create <название>
Пример: /create ЛР3`

const JoinUsageMessage = `Использование:
/join Фамилия Имя — встать в конец
/join Фамилия Имя 3 — на позицию 3
/join Фамилия Имя #3 — то же, с решёткой

Пример: /join Иванов Иван 3`

const HelpMessage = `Команды бота:
/create <название> — создать очередь
/join <Фамилия Имя> [N|#N] — записаться (N — позиция)
/leave — выйти
/positions — свободные позиции
/kickfirst — убрать первого (только админ)
/close — закрыть очередь (только админ)
/debug — реальные order_number (только админ)
/start — приветствие
/help — эта справка

Позицию можно указать числом (3) или с решёткой (#3).
Позиция должна быть 1..999 и свободна.`

type TelegramBot struct {
TgBot *tgbotapi.BotAPI
u     usecase.Bot
}

func NewTelegramBot(tgBot *tgbotapi.BotAPI, u usecase.Bot) *TelegramBot {
return &TelegramBot{TgBot: tgBot, u: u}
}

func queueKey(chatID int64, messageID int) string {
return fmt.Sprintf("%d:%d", chatID, messageID)
}

func parseQueueKey(key string) (int64, int, error) {
parts := strings.SplitN(key, ":", 2)
if len(parts) != 2 {
return 0, 0, fmt.Errorf("invalid key: %s", key)
}
cid, err := strconv.ParseInt(parts[0], 10, 64)
if err != nil {
return 0, 0, err
}
mid, err := strconv.Atoi(parts[1])
if err != nil {
return 0, 0, err
}
return cid, mid, nil
}

func keyFromCallback(cb *tgbotapi.CallbackQuery) string {
return queueKey(cb.Message.Chat.ID, cb.Message.MessageID)
}

func (b TelegramBot) SendHelloMessage(m *tgbotapi.Message) error {
_, err := b.TgBot.Send(tgbotapi.NewMessage(m.Chat.ID, HelloMessage))
return err
}
func (b TelegramBot) SendCreateUsage(chatID int64) error {
_, err := b.TgBot.Send(tgbotapi.NewMessage(chatID, CreateUsageMessage))
return err
}
func (b TelegramBot) SendHelpMessage(chatID int64) error {
_, err := b.TgBot.Send(tgbotapi.NewMessage(chatID, HelpMessage))
return err
}
func (b TelegramBot) SendJoinUsage(chatID int64) error {
_, err := b.TgBot.Send(tgbotapi.NewMessage(chatID, JoinUsageMessage))
return err
}

func (b TelegramBot) SendFreePositions(chatID int64) error {
ctx := context.Background()
key, err := b.u.GetLastQueueKeyInChat(ctx, chatID)
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
text = "Свободных позиций нет. Новая запись встанет в конец."
} else {
parts := make([]string, len(free))
for i, p := range free {
parts[i] = strconv.Itoa(p)
}
text = "Свободные позиции: " + strings.Join(parts, ", ") +
"\nЗанять: /join Иванов Иван <номер> или /join Иванов Иван #<номер>"
}
_, err = b.TgBot.Send(tgbotapi.NewMessage(chatID, text))
return err
}

func (b TelegramBot) SendDebug(chatID, userID int64) error {
ctx := context.Background()
isAdmin, err := b.IsChatAdmin(ctx, chatID, userID)
if err != nil {
return err
}
if !isAdmin {
return ErrNotAdmin
}
key, err := b.u.GetLastQueueKeyInChat(ctx, chatID)
if err != nil {
return err
}
if key == "" {
return ErrNoQueue
}
rows, err := b.u.DebugPositions(ctx, key)
if err != nil {
return err
}
sb := strings.Builder{}
sb.WriteString("Отладка order_number (включая удалённых):\n")
if len(rows) == 0 {
sb.WriteString("(пусто)")
}
for _, r := range rows {
del := ""
if r.Deleted == 1 {
del = " [удалён]"
}
sb.WriteString(fmt.Sprintf("#%d %s (id=%d)%s\n", r.Order, r.UserName, r.UserID, del))
}
_, err = b.TgBot.Send(tgbotapi.NewMessage(chatID, sb.String()))
return err
}

func (b TelegramBot) CreateQueueFromCommand(chatID int64, description string) error {
msg := tgbotapi.NewMessage(chatID, getMessageContentBeforeStart(description, nil))
msg.ParseMode = tgbotapi.ModeMarkdown
kb := GetBeforeStartKeyboard()
msg.ReplyMarkup = kb
sent, err := b.TgBot.Send(msg)
if err != nil {
return fmt.Errorf("couldn't send queue message: %w", err)
}
key := queueKey(chatID, sent.MessageID)
if err := b.u.CreateQueue(context.Background(), key, description); err != nil {
return fmt.Errorf("couldn't create queue: %w", err)
}
slog.Info("Queue created", "key", key, "description", description)
return nil
}

func (b TelegramBot) JoinQueueFromCommand(chatID, userID int64, fullName string, position int) error {
ctx := context.Background()
key, err := b.u.GetLastQueueKeyInChat(ctx, chatID)
if err != nil {
return fmt.Errorf("couldn't find queue: %w", err)
}
if key == "" {
return ErrNoQueue
}
if err := b.u.SaveUserName(ctx, userID, fullName); err != nil {
return fmt.Errorf("couldn't save name: %w", err)
}
if err := b.u.JoinQueue(ctx, key, entity.NewFullName(userID, fullName), position); err != nil {
return err
}
if err := b.refreshQueueMessage(ctx, key); err != nil {
return err
}
// подтверждение
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
_, _ = b.TgBot.Send(tgbotapi.NewMessage(chatID,
fmt.Sprintf("%s записан на позицию %d", fullName, myPos)))
}
}
return nil
}

func (b TelegramBot) LeaveQueueFromCommand(chatID, userID int64) error {
ctx := context.Background()
key, err := b.u.GetLastQueueKeyInChat(ctx, chatID)
if err != nil {
return fmt.Errorf("couldn't find queue: %w", err)
}
if key == "" {
return ErrNoQueue
}
queue, err := b.u.GetQueue(ctx, key)
if err != nil {
return fmt.Errorf("couldn't get queue: %w", err)
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
return fmt.Errorf("couldn't leave: %w", err)
}
return b.refreshQueueMessage(ctx, key)
}

func (b TelegramBot) KickFirstFromCommand(chatID, userID int64) error {
ctx := context.Background()
isAdmin, err := b.IsChatAdmin(ctx, chatID, userID)
if err != nil {
return fmt.Errorf("couldn't check admin: %w", err)
}
if !isAdmin {
return ErrNotAdmin
}
key, err := b.u.GetLastQueueKeyInChat(ctx, chatID)
if err != nil {
return fmt.Errorf("couldn't find queue: %w", err)
}
if key == "" {
return ErrNoQueue
}
if err := b.u.KickFirstUser(ctx, key); err != nil {
return fmt.Errorf("couldn't kick first: %w", err)
}
return b.refreshQueueMessage(ctx, key)
}

func (b TelegramBot) CloseQueueFromCommand(chatID, userID int64) error {
ctx := context.Background()
isAdmin, err := b.IsChatAdmin(ctx, chatID, userID)
if err != nil {
return fmt.Errorf("couldn't check admin: %w", err)
}
if !isAdmin {
return ErrNotAdmin
}
key, err := b.u.GetLastQueueKeyInChat(ctx, chatID)
if err != nil {
return fmt.Errorf("couldn't find queue: %w", err)
}
if key == "" {
return ErrNoQueue
}
_, msgID, _ := parseQueueKey(key)
if _, err := b.TgBot.Request(GetFinishedMessage(chatID, msgID)); err != nil {
return fmt.Errorf("couldn't edit message: %w", err)
}
return b.u.FinishQueue(ctx, key)
}

func (b TelegramBot) refreshQueueMessage(ctx context.Context, key string) error {
chatID, msgID, err := parseQueueKey(key)
if err != nil {
return err
}
queue, err := b.u.GetQueue(ctx, key)
if err != nil {
return fmt.Errorf("couldn't get queue: %w", err)
}
_, err = b.TgBot.Request(GetUpdatedQueueMessage(chatID, msgID, queue.Description, queue.Users))
return err
}

func (b TelegramBot) IsChatAdmin(ctx context.Context, chatID, userID int64) (bool, error) {
member, err := b.TgBot.GetChatMember(tgbotapi.GetChatMemberConfig{
ChatConfigWithUser: tgbotapi.ChatConfigWithUser{ChatID: chatID, UserID: userID},
})
if err != nil {
return false, fmt.Errorf("couldn't get chat member: %w", err)
}
return member.Status == "administrator" || member.Status == "creator", nil
}

func (b TelegramBot) Start(ctx context.Context, cb *tgbotapi.CallbackQuery, isShuffled bool) error {
key := keyFromCallback(cb)
if err := b.u.StartQueue(ctx, key, isShuffled); err != nil {
return fmt.Errorf("couldn't start: %w", err)
}
return b.sendStatus(ctx, cb)
}

func (b TelegramBot) Next(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
key := keyFromCallback(cb)
if err := b.u.SetNextPersonToQueue(ctx, key); err != nil {
return fmt.Errorf("couldn't next: %w", err)
}
return b.sendStatus(ctx, cb)
}

func (b TelegramBot) KickFirst(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
key := keyFromCallback(cb)
if err := b.u.KickFirstUser(ctx, key); err != nil {
return fmt.Errorf("couldn't kick first: %w", err)
}
return b.refreshQueueMessage(ctx, key)
}

func (b TelegramBot) GoToMenu(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
key := keyFromCallback(cb)
queue, err := b.u.GetQueue(ctx, key)
if err != nil {
return fmt.Errorf("couldn't get queue: %w", err)
}
_, err = b.TgBot.Request(GetQueueMessage(cb.Message.Chat.ID, cb.Message.MessageID, queue.Users, queue.Description))
return err
}

func (b TelegramBot) FinishQueue(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
key := keyFromCallback(cb)
if _, err := b.TgBot.Request(GetFinishedMessage(cb.Message.Chat.ID, cb.Message.MessageID)); err != nil {
return fmt.Errorf("couldn't edit: %w", err)
}
return b.u.FinishQueue(ctx, key)
}

func (b TelegramBot) sendStatus(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
key := keyFromCallback(cb)
queue, err := b.u.GetQueue(ctx, key)
if err != nil {
return fmt.Errorf("couldn't get queue: %w", err)
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