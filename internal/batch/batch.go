// Package batch 负责染色批次的生命周期与状态机规则。
package batch

import (
	"fmt"

	"task219-colorreview/internal/model"
)

// CanTransition 判断批次能否从 from 流转到 to。
func CanTransition(from, to model.BatchStatus) bool {
	if from == to {
		return false
	}
	targets, ok := model.ValidBatchTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// Validate 校验批次字段是否合法。
func Validate(b *model.DyeBatch) error {
	if b.Code == "" {
		return fmt.Errorf("%w: code required", model.ErrInvalidArgument)
	}
	if b.Name == "" {
		return fmt.Errorf("%w: name required", model.ErrInvalidArgument)
	}
	if b.Recipe == "" {
		return fmt.Errorf("%w: recipe required", model.ErrInvalidArgument)
	}
	switch b.Status {
	case model.BatchRecipe, model.BatchInProduction, model.BatchPendingReview, model.BatchConfirmed, model.BatchSealed, "":
	default:
		return fmt.Errorf("%w: unknown status %q", model.ErrInvalidArgument, b.Status)
	}
	return nil
}

// IsSealed 判断批次是否已封存（封存后不可变）。
func IsSealed(b *model.DyeBatch) bool { return b != nil && b.Status == model.BatchSealed }

// Next 返回允许的下一状态集合。
func Next(from model.BatchStatus) []model.BatchStatus {
	return model.ValidBatchTransitions[from]
}
