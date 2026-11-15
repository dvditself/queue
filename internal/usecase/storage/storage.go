package storage

import (
"context"

"QueueBot/internal/entity"
"QueueBot/internal/usecase/storage/sqlite"
)

type Storage interface {
CreateQueue(ctx context.Context, messageID string, description string) error
JoinQueue(ctx context.Context, messageID string, user entity.User, position int) error
LeaveQueue(ctx context.Context, messageID string, userID int64) error
KickFirstUser(ctx context.Context, messageID string) error
KickAtPosition(ctx context.Context, messageID string, position int) error
SwitchPositions(ctx context.Context, messageID string, userID int64, position int) error
GetQueue(ctx context.Context, messageID string) (entity.Queue, error)
FreePositions(ctx context.Context, messageID string) ([]int, error)
DebugPositions(ctx context.Context, messageID string) ([]sqlite.DebugRow, error)
DebugChatQueues(ctx context.Context, chatID int64) ([]string, error)

StartQueue(ctx context.Context, messageID string) error
IncrementCurrentPerson(ctx context.Context, messageID string) error
DeleteQueue(ctx context.Context, messageID string) error

SaveUserName(ctx context.Context, userID int64, fullName string) error
GetUserName(ctx context.Context, userID int64) (string, error)
GetLastQueueKeyInChatThread(ctx context.Context, chatID int64, threadID int) (string, error)

Close() error
}