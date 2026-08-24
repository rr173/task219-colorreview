// Package service 编排各业务包与持久化层，实现完整业务闭环。
package service

import (
	"context"

	"task219-colorreview/internal/store"
)

// Service 聚合所有用例，是 HTTP 层的唯一依赖。
type Service struct {
	store *store.Store
}

// New 构造服务。
func New(st *store.Store) *Service {
	return &Service{store: st}
}

// Store 暴露底层仓储（供自检与事务编排使用）。
func (s *Service) Store() *store.Store { return s.store }

// Ping 检查数据库连通性。
func (s *Service) Ping(ctx context.Context) error {
	return s.store.DB().PingContext(ctx)
}
