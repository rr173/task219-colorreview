package store

import (
	"database/sql"
	"errors"

	"task219-colorreview/internal/model"
)

// mapErr 把 SQLite 错误映射为领域哨兵错误。
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return model.ErrNotFound
	}
	return err
}

// isUniqueViolation 判断是否为唯一约束冲突（modernc sqlite 错误文本）。
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// modernc.org/sqlite 对 UNIQUE 冲突返回包含 "constraint failed" 的错误。
	msg := err.Error()
	return containsAny(msg, "UNIQUE constraint failed", "constraint failed")
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) > 0 && contains(s, sub) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	return len(sub) <= len(s) && indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
