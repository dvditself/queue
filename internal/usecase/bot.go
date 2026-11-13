package usecase

import (
"context"

"QueueBot/internal/entity"
"QueueBot/internal/usecase/storage"
"QueueBot/internal/usecase/storage/sqlite"
)

type Bot interface {
CreateQueue(ctx context.Context, messageID string, description string) error
JoinQueue(ctx context.Context, messageID string, user entity.User, position int) error
LeaveQueue(ctx context.Context, messageID string, userID int64) error
KickFirstUser(ctx context.Context, messageID string) error
StartQueue(ctx context.Context, messageID string, shuffle bool) error
FinishQueue(ctx context.Context, messageID string) error
SetNextPersonToQueue(ctx context.Context, messageID string) error
GetQueue(ctx context.Context, messageID string) (entity.Queue, error)
GetFreePositions(ctx context.Context, messageID string) ([]int, error)
DebugPositions(ctx context.Context, messageID string) ([]sqlite.DebugRow, error)

SaveUserName(ctx context.Context, userID int64, fullName string) error
GetUserName(ctx context.Context, userID int64) (string, error)
GetLastQueueKeyInChat(ctx context.Context, chatID int64) (string, error)
}

type BotUseCase struct{ Storage storage.Storage }

func NewBotUseCase(s storage.Storage) *BotUseCase { return &BotUseCase{Storage: s} }

func (b BotUseCase) CreateQueue(ctx context.Context, m, d string) error {
return b.Storage.CreateQueue(ctx, m, d)
}
func (b BotUseCase) JoinQueue(ctx context.Context, m string, u entity.User, p int) error {
return b.Storage.JoinQueue(ctx, m, u, p)
}
func (b BotUseCase) LeaveQueue(ctx context.Context, m string, uid int64) error {
return b.Storage.LeaveQueue(ctx, m, uid)
}
func (b BotUseCase) KickFirstUser(ctx context.Context, m string) error {
return b.Storage.KickFirstUser(ctx, m)
}
func (b BotUseCase) StartQueue(ctx context.Context, m string, s bool) error {
return b.Storage.StartQueue(ctx, m, s)
}
func (b BotUseCase) FinishQueue(ctx context.Context, m string) error {
return b.Storage.DeleteQueue(ctx, m)
}
func (b BotUseCase) SetNextPersonToQueue(ctx context.Context, m string) error {
return b.Storage.IncrementCurrentPerson(ctx, m)
}
func (b BotUseCase) GetQueue(ctx context.Context, m string) (entity.Queue, error) {
return b.Storage.GetQueue(ctx, m)
}
func (b BotUseCase) GetFreePositions(ctx context.Context, m string) ([]int, error) {
return b.Storage.FreePositions(ctx, m)
}
func (b BotUseCase) DebugPositions(ctx context.Context, m string) ([]sqlite.DebugRow, error) {
return b.Storage.DebugPositions(ctx, m)
}
func (b BotUseCase) SaveUserName(ctx context.Context, uid int64, n string) error {
return b.Storage.SaveUserName(ctx, uid, n)
}
func (b BotUseCase) GetUserName(ctx context.Context, uid int64) (string, error) {
return b.Storage.GetUserName(ctx, uid)
}
func (b BotUseCase) GetLastQueueKeyInChat(ctx context.Context, c int64) (string, error) {
return b.Storage.GetLastQueueKeyInChat(ctx, c)
}