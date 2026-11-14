package sqlite

import (
"context"
"database/sql"
"errors"
"fmt"
"strconv"
"strings"

_ "modernc.org/sqlite"

"QueueBot/internal/entity"
)

var (
ErrPositionOccupied = errors.New("position is occupied")
ErrPositionTooBig   = errors.New("position is too big")
ErrAlreadyInQueue   = errors.New("user is already in queue")
ErrPositionEmpty    = errors.New("position is empty")
ErrSamePosition     = errors.New("you are already at this position")
ErrNotInQueue       = errors.New("user is not in queue")
)

const maxAllowedPosition = 999

type Database struct{ db *sql.DB }

func (s Database) Close() error { return s.db.Close() }

func (s Database) migrateOrderNumbers(ctx context.Context) error {
_, err := s.db.ExecContext(ctx, `
UPDATE participants SET order_number = (
SELECT rn FROM (
SELECT user_id, message_id,
       ROW_NUMBER() OVER (PARTITION BY message_id ORDER BY joined_at) AS rn
FROM participants
WHERE isDeleted=0 AND (order_number IS NULL OR order_number=0)
) AS sub
WHERE participants.user_id=sub.user_id
  AND participants.message_id=sub.message_id
)
WHERE isDeleted=0 AND (order_number IS NULL OR order_number=0)`)
return err
}

func (s Database) migrateOldKeys(ctx context.Context) error {
rows, err := s.db.QueryContext(ctx, "SELECT message_id FROM queues")
if err != nil {
return err
}
defer rows.Close()
var toMigrate []string
for rows.Next() {
var k string
if err := rows.Scan(&k); err != nil {
return err
}
if strings.Count(k, ":") == 1 {
toMigrate = append(toMigrate, k)
}
}
for _, oldKey := range toMigrate {
parts := strings.SplitN(oldKey, ":", 2)
if len(parts) != 2 {
continue
}
newKey := fmt.Sprintf("%s:0:%s", parts[0], parts[1])
if _, err := s.db.ExecContext(ctx, "UPDATE queues SET message_id=? WHERE message_id=?", newKey, oldKey); err != nil {
return err
}
if _, err := s.db.ExecContext(ctx, "UPDATE participants SET message_id=? WHERE message_id=?", newKey, oldKey); err != nil {
return err
}
}
return nil
}

func (s Database) CreateQueue(ctx context.Context, messageID string, description string) error {
stmt, err := s.db.PrepareContext(ctx, "INSERT INTO queues (message_id, description) VALUES (?, ?)")
if err != nil {
return fmt.Errorf("prepare create queue: %w", err)
}
defer stmt.Close()
_, err = stmt.ExecContext(ctx, messageID, description)
return err
}

func (s Database) FreePositions(ctx context.Context, messageID string) ([]int, error) {
rows, err := s.db.QueryContext(ctx, "SELECT order_number FROM participants WHERE message_id=? AND isDeleted=0", messageID)
if err != nil {
return nil, err
}
defer rows.Close()
occupied := map[int]bool{}
max := 0
for rows.Next() {
var n sql.NullInt64
if err := rows.Scan(&n); err != nil {
return nil, err
}
if n.Valid && n.Int64 > 0 {
occupied[int(n.Int64)] = true
if int(n.Int64) > max {
max = int(n.Int64)
}
}
}
var free []int
for i := 1; i <= max+1; i++ {
if !occupied[i] {
free = append(free, i)
}
}
return free, nil
}

func (s Database) JoinQueue(ctx context.Context, messageID string, user entity.User, position int) error {
if err := s.migrateOrderNumbers(ctx); err != nil {
return err
}
tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
return fmt.Errorf("begin tx: %w", err)
}
defer func() { _ = tx.Rollback() }()

var cnt int
if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM participants WHERE message_id=? AND user_id=? AND isDeleted=0", messageID, user.ID).Scan(&cnt); err != nil {
return fmt.Errorf("check existing: %w", err)
}
if cnt > 0 && position <= 0 {
return ErrAlreadyInQueue
}

var targetPos int
if position <= 0 {
rows, err := tx.QueryContext(ctx, "SELECT order_number FROM participants WHERE message_id=? AND isDeleted=0", messageID)
if err != nil {
return fmt.Errorf("query occupied: %w", err)
}
occupied := map[int]bool{}
for rows.Next() {
var n sql.NullInt64
if err := rows.Scan(&n); err != nil {
rows.Close()
return err
}
if n.Valid && n.Int64 > 0 {
occupied[int(n.Int64)] = true
}
}
rows.Close()
targetPos = 1
for occupied[targetPos] {
targetPos++
}
if targetPos > maxAllowedPosition {
return ErrPositionTooBig
}
} else {
if position < 1 || position > maxAllowedPosition {
return ErrPositionTooBig
}
var occupied int
if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM participants WHERE message_id=? AND isDeleted=0 AND order_number=? AND user_id<>?", messageID, position, user.ID).Scan(&occupied); err != nil {
return fmt.Errorf("check occupied: %w", err)
}
if occupied > 0 {
return ErrPositionOccupied
}
targetPos = position
}

stmt, err := tx.PrepareContext(ctx, `INSERT INTO participants(message_id, user_id, user_name, order_number)
VALUES (?, ?, ?, ?)
ON CONFLICT(message_id, user_id) DO UPDATE SET
user_name=excluded.user_name, isDeleted=0, joined_at=CURRENT_TIMESTAMP, order_number=excluded.order_number`)
if err != nil {
return fmt.Errorf("prepare join: %w", err)
}
defer stmt.Close()
if _, err := stmt.ExecContext(ctx, messageID, user.ID, user.Name, targetPos); err != nil {
return fmt.Errorf("exec join: %w", err)
}
return tx.Commit()
}

func (s Database) SwitchPositions(ctx context.Context, messageID string, userID int64, position int) error {
if err := s.migrateOrderNumbers(ctx); err != nil {
return err
}
tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
return fmt.Errorf("begin tx: %w", err)
}
defer func() { _ = tx.Rollback() }()

var myPos sql.NullInt64
err = tx.QueryRowContext(ctx, "SELECT order_number FROM participants WHERE message_id=? AND user_id=? AND isDeleted=0", messageID, userID).Scan(&myPos)
if errors.Is(err, sql.ErrNoRows) {
return ErrNotInQueue
}
if err != nil {
return fmt.Errorf("get my pos: %w", err)
}
if !myPos.Valid || myPos.Int64 <= 0 {
return ErrNotInQueue
}
if int(myPos.Int64) == position {
return ErrSamePosition
}

var otherID int64
err = tx.QueryRowContext(ctx, "SELECT user_id FROM participants WHERE message_id=? AND order_number=? AND isDeleted=0", messageID, position).Scan(&otherID)
if errors.Is(err, sql.ErrNoRows) {
return ErrPositionEmpty
}
if err != nil {
return fmt.Errorf("get other: %w", err)
}

if _, err := tx.ExecContext(ctx, "UPDATE participants SET order_number=? WHERE message_id=? AND user_id=?", position, messageID, userID); err != nil {
return fmt.Errorf("update my pos: %w", err)
}
if _, err := tx.ExecContext(ctx, "UPDATE participants SET order_number=? WHERE message_id=? AND user_id=?", int(myPos.Int64), messageID, otherID); err != nil {
return fmt.Errorf("update other pos: %w", err)
}
return tx.Commit()
}

func (s Database) KickAtPosition(ctx context.Context, messageID string, position int) error {
res, err := s.db.ExecContext(ctx, "UPDATE participants SET isDeleted=1 WHERE message_id=? AND order_number=? AND isDeleted=0", messageID, position)
if err != nil {
return fmt.Errorf("kick at position: %w", err)
}
n, _ := res.RowsAffected()
if n == 0 {
return ErrPositionEmpty
}
return nil
}

func (s Database) LeaveQueue(ctx context.Context, messageID string, userID int64) error {
_, err := s.db.ExecContext(ctx, "UPDATE participants SET isDeleted=1 WHERE message_id=? AND user_id=? AND isDeleted=0", messageID, userID)
return err
}

func (s Database) KickFirstUser(ctx context.Context, messageID string) error {
tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
return fmt.Errorf("begin tx: %w", err)
}
defer func() { _ = tx.Rollback() }()

if _, err := tx.ExecContext(ctx, `
UPDATE participants SET isDeleted=1
WHERE message_id=? AND isDeleted=0 AND user_id = (
SELECT user_id FROM participants WHERE message_id=? AND isDeleted=0
ORDER BY COALESCE(order_number, 1000000), joined_at LIMIT 1
)`, messageID, messageID); err != nil {
return fmt.Errorf("kick first: %w", err)
}
if _, err := tx.ExecContext(ctx, `
UPDATE queues SET current_user_index =
CASE WHEN current_user_index>0 THEN current_user_index-1 ELSE 0 END
WHERE message_id=?`, messageID); err != nil {
return fmt.Errorf("adjust idx: %w", err)
}
return tx.Commit()
}

func (s Database) GetQueue(ctx context.Context, messageID string) (entity.Queue, error) {
if err := s.migrateOrderNumbers(ctx); err != nil {
return entity.Queue{}, err
}
var description string
var currentUserIndex int
if err := s.db.QueryRowContext(ctx, "SELECT description, current_user_index FROM queues WHERE message_id=?", messageID).Scan(&description, &currentUserIndex); err != nil {
return entity.Queue{}, fmt.Errorf("scan queue: %w", err)
}
rows, err := s.db.QueryContext(ctx, "SELECT user_id, user_name, COALESCE(order_number, 0) FROM participants WHERE message_id=? AND isDeleted=0 ORDER BY COALESCE(order_number, 1000000), joined_at", messageID)
if err != nil {
return entity.Queue{}, fmt.Errorf("query users: %w", err)
}
defer rows.Close()
var users []entity.User
for rows.Next() {
var u entity.User
var pos int
if err := rows.Scan(&u.ID, &u.Name, &pos); err != nil {
return entity.Queue{}, fmt.Errorf("scan user: %w", err)
}
u.Position = pos
users = append(users, u)
}
return entity.Queue{MessageID: messageID, Description: description, Users: users, CurrentPersonIdx: currentUserIndex}, nil
}

type DebugRow struct {
UserID   int64
UserName string
Order    int
Deleted  int
}

func (s Database) DebugPositions(ctx context.Context, messageID string) ([]DebugRow, error) {
rows, err := s.db.QueryContext(ctx, "SELECT user_id, user_name, COALESCE(order_number, 0), isDeleted FROM participants WHERE message_id=? ORDER BY COALESCE(order_number, 1000000), joined_at", messageID)
if err != nil {
return nil, err
}
defer rows.Close()
var result []DebugRow
for rows.Next() {
var r DebugRow
if err := rows.Scan(&r.UserID, &r.UserName, &r.Order, &r.Deleted); err != nil {
return nil, err
}
result = append(result, r)
}
return result, nil
}

func (s Database) DebugChatQueues(ctx context.Context, chatID int64) ([]string, error) {
prefix := fmt.Sprintf("%d:%%", chatID)
rows, err := s.db.QueryContext(ctx, "SELECT message_id FROM queues WHERE message_id LIKE ? ORDER BY message_id", prefix)
if err != nil {
return nil, err
}
defer rows.Close()
var keys []string
for rows.Next() {
var k string
if err := rows.Scan(&k); err != nil {
return nil, err
}
keys = append(keys, k)
}
return keys, nil
}

func (s Database) StartQueue(ctx context.Context, messageID string, isShuffle bool) error {
if err := s.migrateOrderNumbers(ctx); err != nil {
return err
}
tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
return fmt.Errorf("begin tx: %w", err)
}
defer func() { _ = tx.Rollback() }()

if _, err := tx.ExecContext(ctx, "UPDATE queues SET current_user_index=0 WHERE message_id=?", messageID); err != nil {
return fmt.Errorf("reset idx: %w", err)
}
orderExpr := "COALESCE(order_number, 1000000), joined_at"
if isShuffle {
orderExpr = "random()"
}
if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
UPDATE participants SET order_number = (
SELECT rn FROM (
SELECT user_id, ROW_NUMBER() OVER (ORDER BY %s) AS rn
FROM participants WHERE message_id=$1 AND isDeleted=0
) AS sub
WHERE participants.user_id=sub.user_id AND participants.message_id=$1
)
WHERE message_id=$1 AND isDeleted=0`, orderExpr), messageID); err != nil {
return fmt.Errorf("renumber: %w", err)
}
return tx.Commit()
}

func (s Database) IncrementCurrentPerson(ctx context.Context, messageID string) error {
_, err := s.db.ExecContext(ctx, "UPDATE queues SET current_user_index=current_user_index+1 WHERE message_id=?", messageID)
return err
}

func (s Database) DeleteQueue(ctx context.Context, messageID string) error {
tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
return fmt.Errorf("begin tx: %w", err)
}
defer func() { _ = tx.Rollback() }()
if _, err := tx.ExecContext(ctx, "DELETE FROM participants WHERE message_id=?", messageID); err != nil {
return err
}
if _, err := tx.ExecContext(ctx, "DELETE FROM queues WHERE message_id=?", messageID); err != nil {
return err
}
return tx.Commit()
}

func (s Database) SaveUserName(ctx context.Context, userID int64, fullName string) error {
_, err := s.db.ExecContext(ctx, `INSERT INTO user_names(user_id, full_name) VALUES (?, ?)
ON CONFLICT(user_id) DO UPDATE SET full_name=excluded.full_name, updated_at=CURRENT_TIMESTAMP`, userID, fullName)
return err
}

func (s Database) GetUserName(ctx context.Context, userID int64) (string, error) {
var name string
err := s.db.QueryRowContext(ctx, "SELECT full_name FROM user_names WHERE user_id=?", userID).Scan(&name)
if errors.Is(err, sql.ErrNoRows) {
return "", nil
}
return name, err
}

func (s Database) GetLastQueueKeyInChatThread(ctx context.Context, chatID int64, threadID int) (string, error) {
prefix := fmt.Sprintf("%d:%d:", chatID, threadID)
rows, err := s.db.QueryContext(ctx, "SELECT message_id FROM queues WHERE message_id LIKE ?", prefix+"%")
if err != nil {
return "", fmt.Errorf("query keys: %w", err)
}
defer rows.Close()
best := ""
bestMid := -1
for rows.Next() {
var k string
if err := rows.Scan(&k); err != nil {
return "", err
}
parts := strings.SplitN(k, ":", 3)
if len(parts) != 3 {
continue
}
mid, err := strconv.Atoi(parts[2])
if err != nil {
continue
}
if mid > bestMid {
bestMid = mid
best = k
}
}
return best, nil
}

func NewDatabase(databasePath string) (*Database, error) {
db, err := sql.Open("sqlite", fmt.Sprintf("%s?cache=shared", databasePath))
if err != nil {
return nil, fmt.Errorf("open db: %w", err)
}
if _, err := db.Exec(CreateTables); err != nil {
return nil, fmt.Errorf("create tables: %w", err)
}
d := &Database{db: db}
ctx := context.Background()
_ = d.migrateOldKeys(ctx)
_ = d.migrateOrderNumbers(ctx)
return d, nil
}

func NewDatabaseFromDB(db *sql.DB) *Database { return &Database{db: db} }