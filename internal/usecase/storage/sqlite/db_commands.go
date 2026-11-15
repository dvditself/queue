package sqlite

const CreateTables string = `
CREATE TABLE IF NOT EXISTS queues
(
    message_id         TEXT NOT NULL PRIMARY KEY,
    description        TEXT    DEFAULT NULL,
    current_user_index INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_queues_message_id ON queues (message_id, current_user_index);

CREATE TABLE IF NOT EXISTS participants
(
    message_id   TEXT NOT NULL REFERENCES queues (message_id),
    user_id      BIGINT  NOT NULL,
    user_name    VARCHAR NOT NULL,
    joined_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    order_number INTEGER,
    isDeleted    INTEGER NOT NULL DEFAULT 0,
    primary key (message_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_prt_message_id ON participants (message_id, user_id, isDeleted);

CREATE TABLE IF NOT EXISTS user_names
(
    user_id    BIGINT NOT NULL PRIMARY KEY,
    full_name  VARCHAR NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

// CreateUniquePosIndex — создаётся ПОСЛЕ миграции, чтобы не упасть на дубликатах.
const CreateUniquePosIndex string = `
CREATE UNIQUE INDEX IF NOT EXISTS idx_participants_msg_pos
ON participants (message_id, order_number)
WHERE isDeleted = 0;
`