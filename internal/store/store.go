// Package store 提供基于 SQLite 的持久化层：建表迁移与各实体 CRUD。
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // 纯 Go SQLite 驱动，CGO 无关
)

// Store 封装数据库连接与各实体的仓储。
type Store struct {
	db *sql.DB
}

// schema 定义全部建表语句（幂等，IF NOT EXISTS）。
var schema = []string{
	`CREATE TABLE IF NOT EXISTS batches (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		code        TEXT NOT NULL UNIQUE,
		name        TEXT NOT NULL,
		recipe      TEXT NOT NULL,
		color_space TEXT NOT NULL DEFAULT '',
		status      TEXT NOT NULL,
		created_at  TEXT NOT NULL,
		updated_at  TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS bath_curves (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		batch_id   INTEGER NOT NULL,
		channel    TEXT NOT NULL,
		points     TEXT NOT NULL,
		created_at TEXT NOT NULL,
		UNIQUE(batch_id, channel)
	)`,
	`CREATE TABLE IF NOT EXISTS measure_points (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		batch_id      INTEGER NOT NULL,
		sample_no     INTEGER NOT NULL,
		position      TEXT NOT NULL,
		color_space   TEXT NOT NULL,
		l             REAL NOT NULL,
		a             REAL NOT NULL,
		b             REAL NOT NULL,
		measured_at   TEXT NOT NULL,
		status        TEXT NOT NULL,
		reject_reason TEXT NOT NULL DEFAULT '',
		delta_e       REAL NOT NULL DEFAULT 0,
		created_at    TEXT NOT NULL,
		UNIQUE(batch_id, sample_no)
	)`,
	`CREATE TABLE IF NOT EXISTS instrument_calibrations (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		instrument_id TEXT NOT NULL,
		calibrated_at TEXT NOT NULL,
		ref_l         REAL NOT NULL,
		ref_a         REAL NOT NULL,
		ref_b         REAL NOT NULL,
		offset_l      REAL NOT NULL,
		offset_a      REAL NOT NULL,
		offset_b      REAL NOT NULL,
		created_at    TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS evidences (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		batch_id    INTEGER NOT NULL,
		kind        TEXT NOT NULL,
		description TEXT NOT NULL,
		status      TEXT NOT NULL,
		attached_at TEXT NOT NULL,
		created_at  TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS conclusions (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		batch_id     INTEGER NOT NULL,
		verdict      TEXT NOT NULL,
		summary      TEXT NOT NULL,
		status       TEXT NOT NULL,
		version      INTEGER NOT NULL,
		published_at TEXT,
		created_at   TEXT NOT NULL,
		updated_at   TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS conclusion_versions (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		conclusion_id INTEGER NOT NULL,
		batch_id    INTEGER NOT NULL,
		verdict     TEXT NOT NULL,
		summary     TEXT NOT NULL,
		version     INTEGER NOT NULL,
		snapshot_at TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_measure_batch ON measure_points(batch_id)`,
	`CREATE INDEX IF NOT EXISTS idx_evidence_batch ON evidences(batch_id)`,
	`CREATE INDEX IF NOT EXISTS idx_conclusion_batch ON conclusions(batch_id)`,
	`CREATE INDEX IF NOT EXISTS idx_calibration_instrument ON instrument_calibrations(instrument_id)`,
}

// Open 打开（或创建）SQLite 数据库并执行建表迁移。
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// SQLite 并发：单写多读，合理设置连接池避免 busy 锁。
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(0)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	for _, stmt := range schema {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

// Close 关闭数据库连接。
func (s *Store) Close() error { return s.db.Close() }

// DB 暴露底层连接（供事务编排使用）。
func (s *Store) DB() *sql.DB { return s.db }

// now 返回带时区的时间字符串（SQLite 存储格式）。
func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }

// WithTx 在事务内执行 fn，自动提交或回滚。
func WithTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
