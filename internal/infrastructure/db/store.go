package db

import (
	"database/sql"
	"fmt"

	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	_ "github.com/mattn/go-sqlite3"
)

type UserStatesStore interface {
    Get(chatID int64) (*domain.UserState, error)
    Set(chatID int64, state *domain.UserState) error
    Delete(chatID int64) error
}


type SQLiteUserStateStore struct {
	db *sql.DB
}

func NewSQLiteUserStateStore(dbPath string) (*SQLiteUserStateStore, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    createTable := `
    CREATE TABLE IF NOT EXISTS user_states (
        chat_id   INTEGER PRIMARY KEY,
        min_diff  REAL NOT NULL,
        max_sum   REAL NOT NULL,
        step      TEXT NOT NULL
    );
    `
    if _, err := db.Exec(createTable); err != nil {
        return nil, fmt.Errorf("failed to create user_states table: %w", err)
    }

    return &SQLiteUserStateStore{db: db}, nil
}

func (s *SQLiteUserStateStore) Set(chatID int64, state *domain.UserState) error {
    query := `INSERT OR REPLACE INTO user_states (chat_id, min_diff, max_sum, step) VALUES (?, ?, ?, ?)`
    if _, err := s.db.Exec(query, chatID, state.MinDiff, state.MaxSum, state.Step); err != nil {
    	logger.Log.Errorf("failed to exec DB: %v", err)
		return err
	}

	return nil
}

func (s *SQLiteUserStateStore) Get(chatID int64) (*domain.UserState, error) {
    query := `SELECT min_diff, max_sum, step FROM user_states WHERE chat_id = ?`
    row:= s.db.QueryRow(query, chatID)
	var state domain.UserState
	if err := row.Scan(&state.MinDiff, &state.MaxSum, &state.Step); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logger.Log.Errorf("failed to scan DB row: %v", err)
		return nil, err
	}

	return &state, nil
}

func (s *SQLiteUserStateStore) Delete(chatID int64) error {
	query := `DELETE FROM user_states WHERE chat_id = ?`
	_, err := s.db.Exec(query, chatID)
	if err != nil {
		logger.Log.Errorf("failed to delete from DB: %v", err)
	}
	return err
}
