package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type AppDB struct {
	db *sql.DB
}

// If dsn is empty, it defaults to using the existing file.
func OpenAppDB(ctx context.Context, dsn string) (*AppDB, error) {
	if strings.TrimSpace(dsn) == "" {
		dsn = "file:V5.db?_foreign_keys=on&busy_timeout=5000"
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	appDB := &AppDB{db: db}
	if err := appDB.EnsureSchema(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return appDB, nil
}

func (a *AppDB) Close() error {
	if a == nil || a.db == nil {
		return nil
	}
	return a.db.Close()
}

func (a *AppDB) EnsureSchema(ctx context.Context) error {
	if a == nil || a.db == nil {
		return errors.New("db is nil")
	}
	const schema = `
		-- Table for app_aliases (already existing) --
		CREATE TABLE IF NOT EXISTS app_aliases (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chat_jid TEXT NOT NULL,
			sender_jid TEXT NOT NULL,
			alias TEXT NOT NULL,
			UNIQUE(chat_jid, sender_jid)
		);

		CREATE INDEX IF NOT EXISTS idx_app_aliases_chat_jid
			ON app_aliases(chat_jid);

		CREATE INDEX IF NOT EXISTS idx_app_aliases_sender_jid
			ON app_aliases(sender_jid);

		-- Table for group whitelist
		CREATE TABLE IF NOT EXISTS app_group_whitelist (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chat_jid TEXT NOT NULL UNIQUE
		);

		CREATE INDEX IF NOT EXISTS idx_app_group_whitelist_chat_jid
			ON app_group_whitelist(chat_jid);

		-- Table for user whitelist
		CREATE TABLE IF NOT EXISTS app_user_whitelist (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sender_jid TEXT NOT NULL UNIQUE
		);

		CREATE INDEX IF NOT EXISTS idx_app_user_whitelist_sender_jid
			ON app_user_whitelist(sender_jid);

		-- Table for image cache --
		CREATE TABLE IF NOT EXISTS app_image_cache (
			id TEXT PRIMARY KEY,
			description TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_app_image_cache 
			ON app_image_cache(id);

		-- Table for MessageContext --
		CREATE TABLE IF NOT EXISTS app_message_context (
			message_id TEXT PRIMARY KEY NOT NULL,
			media_description Text,
			chat_id TEXT NOT NULL,
			sender_name TEXT NOT NULL,
			text TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_app_message_context_chat_id
			ON app_message_context(chat_id);
	`
	_, err := a.db.ExecContext(ctx, schema)
	return err
}

// --- Alias methods ---

func (a *AppDB) SetAlias(ctx context.Context, chatJID string, senderJID string, alias string) error {
	if a == nil || a.db == nil {
		return errors.New("db is nil")
	}

	chatJID = strings.TrimSpace(chatJID)
	senderJID = strings.TrimSpace(senderJID)
	alias = strings.TrimSpace(alias)

	if chatJID == "" || senderJID == "" || alias == "" {
		return errors.New("chatJID, senderJID and alias are required")
	}

	query := `
		INSERT INTO app_aliases (
		chat_jid,
		sender_jid,
		alias
		)
		VALUES (?, ?, ?)
		ON CONFLICT(chat_jid, sender_jid) DO UPDATE SET
		alias = excluded.alias
		`
	_, err := a.db.ExecContext(ctx, query, chatJID, senderJID, alias)

	return err
}

func (a *AppDB) GetAlias(ctx context.Context, chatJID string, senderJID string) (string, bool, error) {
	if a == nil || a.db == nil {
		return "", false, errors.New("db is nil")
	}

	chatJID = strings.TrimSpace(chatJID)
	senderJID = strings.TrimSpace(senderJID)

	if chatJID == "" || senderJID == "" {
		return "", false, errors.New("chatJID and senderJID are required")
	}

	var alias string
	query := `
		SELECT alias
		FROM app_aliases
		WHERE chat_jid = ? AND sender_jid = ?
		`
	err := a.db.QueryRowContext(ctx, query, chatJID, senderJID).Scan(&alias)

	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	return alias, true, nil
}

// --- Group Whitelist methods ---

func (a *AppDB) AddGroupToWhitelist(ctx context.Context, chatJID string) error {
	if a == nil || a.db == nil {
		return errors.New("db is nil")
	}
	chatJID = strings.TrimSpace(chatJID)
	if chatJID == "" {
		return errors.New("chatJID required")
	}
	query := `
		INSERT INTO app_group_whitelist (chat_jid)
		VALUES (?)
		ON CONFLICT(chat_jid) DO NOTHING
		`
	_, err := a.db.ExecContext(ctx, query, chatJID)
	return err
}

func (a *AppDB) IsGroupWhitelisted(ctx context.Context, chatJID string) (bool, error) {
	if a == nil || a.db == nil {
		return false, errors.New("db is nil")
	}
	chatJID = strings.TrimSpace(chatJID)
	if chatJID == "" {
		return false, errors.New("chatJID required")
	}

	var id int64
	query := `
		SELECT id FROM app_group_whitelist WHERE chat_jid = ?
		`
	err := a.db.QueryRowContext(ctx, query, chatJID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *AppDB) RemoveGroupFromWhitelist(ctx context.Context, chatJID string) error {
	if a == nil || a.db == nil {
		return errors.New("db is nil")
	}
	chatJID = strings.TrimSpace(chatJID)
	if chatJID == "" {
		return errors.New("chatJID required")
	}
	query := `
		DELETE FROM app_group_whitelist WHERE chat_jid = ?
		`
	_, err := a.db.ExecContext(ctx, query, chatJID)
	return err
}

// --- User Whitelist methods ---

func (a *AppDB) AddUserToWhitelist(ctx context.Context, senderJID string) error {
	if a == nil || a.db == nil {
		return errors.New("db is nil")
	}
	senderJID = strings.TrimSpace(senderJID)
	if senderJID == "" {
		return errors.New("senderJID required")
	}
	query := `
		INSERT INTO app_user_whitelist (sender_jid)
		VALUES (?)
		ON CONFLICT(sender_jid) DO NOTHING
		`
	_, err := a.db.ExecContext(ctx, query, senderJID)
	return err
}

func (a *AppDB) IsUserWhitelisted(ctx context.Context, senderJID string) (bool, error) {
	if a == nil || a.db == nil {
		return false, errors.New("db is nil")
	}
	senderJID = strings.TrimSpace(senderJID)
	if senderJID == "" {
		return false, errors.New("senderJID required")
	}

	var id int64
	query := `
		SELECT id FROM app_user_whitelist WHERE sender_jid = ?
		`
	err := a.db.QueryRowContext(ctx, query, senderJID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *AppDB) RemoveUserFromWhitelist(ctx context.Context, senderJID string) error {
	if a == nil || a.db == nil {
		return errors.New("db is nil")
	}
	senderJID = strings.TrimSpace(senderJID)
	if senderJID == "" {
		return errors.New("senderJID required")
	}
	query := `
		DELETE FROM app_user_whitelist WHERE sender_jid = ?
		`
	_, err := a.db.ExecContext(ctx, query, senderJID)
	return err
}

// --- Image Cache methods ---

func (a *AppDB) AddImageDescription(ctx context.Context, id string, description string) error {
	if a == nil || a.db == nil {
		return errors.New("db is nil")
	}
	id = strings.TrimSpace(id)
	description = strings.TrimSpace(description)
	if id == "" || description == "" {
		return errors.New("id and description are required")
	}
	query := `
		INSERT INTO app_image_cache (id, description)
		VALUES (?, ?)
		ON CONFLICT(id) DO UPDATE SET
			description = excluded.description
	`
	_, err := a.db.ExecContext(ctx, query, id, description)
	return err
}

func (a *AppDB) GetImageDescription(ctx context.Context, id string) (string, error) {
	if a == nil || a.db == nil {
		return "", errors.New("db is nil")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return "", errors.New("id is required")
	}
	var description string
	query := `
		SELECT description FROM app_image_cache WHERE id = ?
	`
	err := a.db.QueryRowContext(ctx, query, id).Scan(&description)
	if errors.Is(err, sql.ErrNoRows) {
		return "", sql.ErrNoRows
	}
	if err != nil {
		return "", err
	}
	return description, nil
}

// --- Message Context methods ---

func (a *AppDB) InsertMessageContext(ctx context.Context, messageID string, chatID string, senderName string, mediaDescription *string, text *string) error {
	if a == nil || a.db == nil {
		return errors.New("db is nil")
	}
	messageID = strings.TrimSpace(messageID)
	chatID = strings.TrimSpace(chatID)
	senderName = strings.TrimSpace(senderName)

	if messageID == "" || chatID == "" || senderName == "" {
		return errors.New("messageID, chatID and senderName are required")
	}

	query := `
		INSERT INTO app_message_context (
			message_id,
			chat_id,
			sender_name,
			media_description,
			text
		)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(message_id) DO UPDATE SET
			chat_id = excluded.chat_id,
			sender_name = excluded.sender_name,
			media_description = excluded.media_description,
			text = excluded.text
	`
	_, err := a.db.ExecContext(ctx, query, messageID, chatID, senderName, mediaDescription, text)
	return err
}

func (a *AppDB) UpdateMessageContextMediaDescription(ctx context.Context, messageID string, mediaDescription string) error {
	if a == nil || a.db == nil {
		return errors.New("db is nil")
	}
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return errors.New("messageID is required")
	}

	query := `
		UPDATE app_message_context
		SET media_description = ?
		WHERE message_id = ?
	`
	_, err := a.db.ExecContext(ctx, query, mediaDescription, messageID)
	return err
}

func (a *AppDB) UpdateMessageContextText(ctx context.Context, messageID string, text string) error {
	if a == nil || a.db == nil {
		return errors.New("db is nil")
	}
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return errors.New("messageID is required")
	}

	query := `
		UPDATE app_message_context
		SET text = ?
		WHERE message_id = ?
	`
	_, err := a.db.ExecContext(ctx, query, text, messageID)
	return err
}
