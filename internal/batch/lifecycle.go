package batch

import (
	"fmt"

	"task219-colorreview/internal/model"
)

// Lifecycle 描述一次批次状态推进操作及其合法性判定。
type Lifecycle struct {
	From model.BatchStatus
	To   model.BatchStatus
}

// Advance 根据当前状态计算推进目标。
// 业务规则：配方中 -> 生产中 -> 待复核 -> 已确认 -> 封存；待复核也可直接封存。
func Advance(cur model.BatchStatus) (model.BatchStatus, error) {
	switch cur {
	case model.BatchRecipe:
		return model.BatchInProduction, nil
	case model.BatchInProduction:
		return model.BatchPendingReview, nil
	case model.BatchPendingReview:
		return model.BatchConfirmed, nil
	case model.BatchConfirmed:
		return model.BatchSealed, nil
	case model.BatchSealed:
		return model.BatchSealed, nil
	default:
		return "", fmt.Errorf("%w: unknown status %q", model.ErrInvalidTransition, cur)
	}
}

// Sealable 判断当前状态是否允许直接封存。
func Sealable(cur model.BatchStatus) bool {
	return cur == model.BatchPendingReview || cur == model.BatchConfirmed
}
