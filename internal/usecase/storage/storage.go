package storage

import (
"context"

"QueueBot/internal/entity"
)

type Storage interface {
CreateQueue(ctx context.Context, messageID string, description string) error
JoinQueue(ctx context.Context, messageID string, user entity.User) error
LeaveQueue(ctx context.Context, messageID string, userID int64) error
GetQueue(ctx context.Context, messageID string) (entity.Queue, error)

StartQueue(ctx context.Context, messageID string, isShuffle bool) error
IncrementCurrentPerson(ctx context.Context, messageID string) error
DeleteQueue(ctx context.Context, messageID string) error

SaveUserName(ctx context.Context, userID int64, fullName string) error
GetUserName(ctx context.Context, userID int64) (string, error)
GetLastQueueKeyInChat(ctx context.Context, chatID int64) (string, error)

Close() error
}