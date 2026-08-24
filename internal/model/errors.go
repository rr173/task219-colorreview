package model

import (
	"errors"
	"fmt"
)

// 领域错误：业务规则违反时返回，由 HTTP 层映射为状态码。
var (
	ErrNotFound           = errors.New("resource not found")
	ErrAlreadyExists      = errors.New("resource already exists")
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrBatchSealed        = errors.New("batch is sealed and immutable")
	ErrInvalidTransition  = errors.New("invalid state transition")
	ErrColorSpaceMissing  = errors.New("color space not declared")
	ErrTimeInverted       = errors.New("timeline inverted")
	ErrRejectReasonMissing = errors.New("reject reason required")
	ErrConclusionFrozen   = errors.New("conclusion is frozen, create a successor version")
	ErrDuplicateSample    = errors.New("duplicate sample number")
	ErrUnknownBatch       = errors.New("unknown batch")
	ErrCalibrationMissing = errors.New("instrument not calibrated")
	ErrConflictUnresolved = errors.New("unresolved conflict")
)

// DomainError 携带业务上下文的错误。
type DomainError struct {
	Op  string
	Err error
}

func (e *DomainError) Error() string {
	if e.Op == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *DomainError) Unwrap() error { return e.Err }

// Wrap 用操作名包裹底层错误。
func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return &DomainError{Op: op, Err: err}
}

// Is 判断错误链中是否包含目标哨兵错误。
func Is(err, target error) bool { return errors.Is(err, target) }
