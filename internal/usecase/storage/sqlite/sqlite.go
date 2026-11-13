package sqlite

import (
"context"
"database/sql"
"errors"
"fmt"

_ "modernc.org/sqlite"

"QueueBot/internal/entity"
)

type Database struct {
db *sql.DB
}

func (s Database) Close() error { return s.db.Close() }

func (s Database) CreateQueue(ctx context.Context, messageID string, description string) error {
stmt, err := s.db.PrepareContext(ctx, "INSERT INTO queues (message_id, description) VALUES (?, ?)")
if err != nil {
return fmt.Errorf("couldn't prepare create queue: %w", err)
}
defer stmt.Close()
_, err = stmt.ExecContext(ctx, messageID, description)
return err
}

func (s Database) JoinQueue(ctx context.Context, messageID string, user entity.User) error {
stmt, err := s.db.PrepareContext(ctx, `INSERT INTO participants(message_id, user_id, user_name)
VALUES (?, ?, ?)
ON CONFLICT(message_id, user_id) DO UPDATE SET
user_name=excluded.user_name,
isDeleted=0,
joined_at=CURRENT_TIMESTAMP,
order_number=NULL`)
if err != nil {
return fmt.Errorf("couldn't prepare join queue: %w", err)
}
defer stmt.Close()
_, err = stmt.ExecContext(ctx, messageID, user.ID, user.Name)
return err
}

func (s Database) LeaveQueue(ctx context.Context, messageID string, userID int64) error {
stmt, err := s.db.PrepareContext(ctx, "UPDATE participants SET isDeleted=1 WHERE message_id = ? AND user_id = ? AND isDeleted = 0")
if err != nil {
return fmt.Errorf("couldn't prepare leave queue: %w", err)
}
defer stmt.Close()
_, err = stmt.ExecContext(ctx, messageID, userID)
return err
}

func (s Database) GetQueue(ctx context.Context, messageID string) (entity.Queue, error) {
descStmt, err := s.db.PrepareContext(ctx, "SELECT description, current_user_index FROM queues WHERE message_id = ?")
if err != nil {
return entity.Queue{}, fmt.Errorf("couldn't prepare desc: %w", err)
}
defer descStmt.Close()

usersStmt, err := s.db.PrepareContext(ctx, "SELECT user_id, user_name FROM participants WHERE message_id = ? AND isDeleted = 0 ORDER BY order_number")
if err != nil {
return entity.Queue{}, fmt.Errorf("couldn't prepare users: %w", err)
}
defer usersStmt.Close()

var description string
var currentUserIndex int
if err := descStmt.QueryRowContext(ctx, messageID).Scan(&description, &currentUserIndex); err != nil {
return entity.Queue{}, fmt.Errorf("couldn't scan desc in %s: %w", messageID, err)
}

rows, err := usersStmt.QueryContext(ctx, messageID)
if err != nil {
return entity.Queue{}, fmt.Errorf("couldn't query users in %s: %w", messageID, err)
}
defer rows.Close()

var users []entity.User
for rows.Next() {
var u entity.User
if err := rows.Scan(&u.ID, &u.Name); err != nil {
return entity.Queue{}, fmt.Errorf("couldn't scan user in %s: %w", messageID, err)
}
users = append(users, u)
}
return entity.Queue{MessageID: messageID, Description: description, Users: users, CurrentPersonIdx: currentUserIndex}, nil
}

func setParticipantsOrder(ctx context.Context, tx *sql.Tx, messageID string, isShuffle bool) error {
var stmt *sql.Stmt
var err error
if isShuffle {
stmt, err = tx.PrepareContext(ctx, `UPDATE participants SET order_number = random() WHERE message_id = ?`)
} else {
stmt, err = tx.PrepareContext(ctx, `UPDATE participants SET order_number = dense_rank FROM
(SELECT dense_rank() OVER (ORDER BY joined_at) AS dense_rank, user_id FROM participants WHERE message_id = $1) AS sub
WHERE participants.user_id = sub.user_id AND message_id = $1`)
}
if err != nil {
return fmt.Errorf("couldn't prepare order: %w", err)
}
defer stmt.Close()
_, err = stmt.ExecContext(ctx, messageID)
return err
}

func (s Database) StartQueue(ctx context.Context, messageID string, isShuffle bool) error {
tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
if err != nil {
return fmt.Errorf("couldn't begin tx: %w", err)
}
defer func() { _ = tx.Rollback() }()

stmt, err := tx.PrepareContext(ctx, "UPDATE queues SET current_user_index = 0 WHERE message_id = ?")
if err != nil {
return fmt.Errorf("couldn't prepare current idx: %w", err)
}
defer stmt.Close()
if _, err := stmt.ExecContext(ctx, messageID); err != nil {
return fmt.Errorf("couldn't set current idx: %w", err)
}
if err := setParticipantsOrder(ctx, tx, messageID, isShuffle); err != nil {
return err
}
return tx.Commit()
}

func (s Database) IncrementCurrentPerson(ctx context.Context, messageID string) error {
stmt, err := s.db.PrepareContext(ctx, "UPDATE queues SET current_user_index = current_user_index + 1 WHERE message_id = ?")
if err != nil {
return fmt.Errorf("couldn't prepare increment: %w", err)
}
defer stmt.Close()
_, err = stmt.ExecContext(ctx, messageID)
return err
}

func (s Database) DeleteQueue(ctx context.Context, messageID string) error {
tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
if err != nil {
return fmt.Errorf("couldn't begin tx: %w", err)
}
defer func() { _ = tx.Rollback() }()

if _, err := tx.ExecContext(ctx, "DELETE FROM participants WHERE message_id = ?", messageID); err != nil {
return fmt.Errorf("couldn't delete participants: %w", err)
}
if _, err := tx.ExecContext(ctx, "DELETE FROM queues WHERE message_id = ?", messageID); err != nil {
return fmt.Errorf("couldn't delete queue: %w", err)
}
return tx.Commit()
}

func (s Database) SaveUserName(ctx context.Context, userID int64, fullName string) error {
stmt, err := s.db.PrepareContext(ctx, `INSERT INTO user_names(user_id, full_name) VALUES (?, ?)
ON CONFLICT(user_id) DO UPDATE SET full_name=excluded.full_name, updated_at=CURRENT_TIMESTAMP`)
if err != nil {
return fmt.Errorf("couldn't prepare save name: %w", err)
}
defer stmt.Close()
_, err = stmt.ExecContext(ctx, userID, fullName)
return err
}

func (s Database) GetUserName(ctx context.Context, userID int64) (string, error) {
var name string
err := s.db.QueryRowContext(ctx, "SELECT full_name FROM user_names WHERE user_id = ?", userID).Scan(&name)
if errors.Is(err, sql.ErrNoRows) {
return "", nil
}
if err != nil {
return "", fmt.Errorf("couldn't get name: %w", err)
}
return name, nil
}

func (s Database) GetLastQueueKeyInChat(ctx context.Context, chatID int64) (string, error) {
prefix := fmt.Sprintf("%d:%%", chatID)
var key string
err := s.db.QueryRowContext(ctx,
`SELECT message_id FROM queues WHERE message_id LIKE ?
 ORDER BY CAST(SUBSTR(message_id, INSTR(message_id, ':')+1) AS INTEGER) DESC LIMIT 1`,
prefix).Scan(&key)
if errors.Is(err, sql.ErrNoRows) {
return "", nil
}
if err != nil {
return "", fmt.Errorf("couldn't find last queue: %w", err)
}
return key, nil
}

func NewDatabase(databasePath string) (*Database, error) {
db, err := sql.Open("sqlite", fmt.Sprintf("%s?cache=shared", databasePath))
if err != nil {
return nil, fmt.Errorf("couldn't open database: %w", err)
}
if _, err := db.Exec(CreateTables); err != nil {
return nil, fmt.Errorf("couldn't create tables: %w", err)
}
return &Database{db: db}, nil
}

func NewDatabaseFromDB(db *sql.DB) *Database { return &Database{db: db} }