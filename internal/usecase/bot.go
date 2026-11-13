package usecase

import (
"context"
"fmt"

"QueueBot/internal/entity"
"QueueBot/internal/usecase/storage"
)

type Bot interface {
CreateQueue(ctx context.Context, messageID string, description string) error
JoinQueue(ctx context.Context, messageID string, user entity.User) error
LeaveQueue(ctx context.Context, messageID string, userID int64) error
KickFirstUser(ctx context.Context, messageID string) error
StartQueue(ctx context.Context, messageID string, shuffle bool) error
FinishQueue(ctx context.Context, messageID string) error
SetNextPersonToQueue(ctx context.Context, messageID string) error
GetQueue(ctx context.Context, messageID string) (entity.Queue, error)

SaveUserName(ctx context.Context, userID int64, fullName string) error
GetUserName(ctx context.Context, userID int64) (string, error)
GetLastQueueKeyInChat(ctx context.Context, chatID int64) (string, error)
}

type BotUseCase struct {
Storage storage.Storage
}

func NewBotUseCase(storage storage.Storage) *BotUseCase { return &BotUseCase{Storage: storage} }

func (b BotUseCase) CreateQueue(ctx context.Context, messageID string, description string) error {
if err := b.Storage.CreateQueue(ctx, messageID, description); err != nil {
return fmt.Errorf("couldn't create queue: %w", err)
}
return nil
}
func (b BotUseCase) JoinQueue(ctx context.Context, messageID string, user entity.User) error {
if err := b.Storage.JoinQueue(ctx, messageID, user); err != nil {
return fmt.Errorf("couldn't join queue: %w", err)
}
return nil
}
func (b BotUseCase) LeaveQueue(ctx context.Context, messageID string, userID int64) error {
if err := b.Storage.LeaveQueue(ctx, messageID, userID); err != nil {
return fmt.Errorf("couldn't leave queue: %w", err)
}
return nil
}
func (b BotUseCase) KickFirstUser(ctx context.Context, messageID string) error {
if err := b.Storage.KickFirstUser(ctx, messageID); err != nil {
return fmt.Errorf("couldn't kick first user: %w", err)
}
return nil
}
func (b BotUseCase) StartQueue(ctx context.Context, messageID string, shuffle bool) error {
if err := b.Storage.StartQueue(ctx, messageID, shuffle); err != nil {
return fmt.Errorf("couldn't start queue: %w", err)
}
return nil
}
func (b BotUseCase) FinishQueue(ctx context.Context, messageID string) error {
if err := b.Storage.DeleteQueue(ctx, messageID); err != nil {
return fmt.Errorf("couldn't delete queue: %w", err)
}
return nil
}
func (b BotUseCase) SetNextPersonToQueue(ctx context.Context, messageID string) error {
if err := b.Storage.IncrementCurrentPerson(ctx, messageID); err != nil {
return fmt.Errorf("couldn't increment current: %w", err)
}
return nil
}
func (b BotUseCase) GetQueue(ctx context.Context, messageID string) (entity.Queue, error) {
q, err := b.Storage.GetQueue(ctx, messageID)
if err != nil {
return entity.Queue{}, fmt.Errorf("couldn't get queue: %w", err)
}
return q, nil
}
func (b BotUseCase) SaveUserName(ctx context.Context, userID int64, fullName string) error {
return b.Storage.SaveUserName(ctx, userID, fullName)
}
func (b BotUseCase) GetUserName(ctx context.Context, userID int64) (string, error) {
return b.Storage.GetUserName(ctx, userID)
}
func (b BotUseCase) GetLastQueueKeyInChat(ctx context.Context, chatID int64) (string, error) {
return b.Storage.GetLastQueueKeyInChat(ctx, chatID)
}