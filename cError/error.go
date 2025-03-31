package cError

import (
	"errors"
	"fmt"
	"net/http"
)

// Error 错误基础结构
type Error struct {
	// 错误码
	Code int `json:"code"`
	// 用户友好的错误消息
	Message string `json:"message"`
	// http响应状态码
	HttpStatus int `json:"-"`
	// 可选的错误详情
	Detail interface{} `json:"detail,omitempty"`
	// 内部错误，不对外暴露
	Internal error `json:"-"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("错误码: %d, 消息: %s", e.Code, e.Message)
}

// New 创建一个新的应用错误
func New(code int, details interface{}, internalErr error) *Error {
	info, exists := errorMap[code]

	// 如果错误码没有被定义，使用通用内部错误
	if !exists {
		info = errorMap[ErrInternal]
		code = ErrInternal
	}

	// TODO
	// 那如果这里传进来的错误就是在errorMap中不存在的
	// 但是我们还是依旧将传入进来的details跟internalErr写入到这个错误里面，是不是有点不合理呢？
	return &Error{
		Code:       code,
		Message:    info.Message,
		HttpStatus: info.HTTPStatus,
		Detail:     details,
		Internal:   internalErr,
	}
}

// NewWithMessage 创建一个带自定义消息的应用错误
func NewWithMessage(code int, message string, details interface{}, internalErr error) *Error {
	info, exists := errorMap[code]
	httpStatus := http.StatusInternalServerError
	// 如果传进来的code能成功检索到对应的http状态码，那就用这个，否则就是500
	if exists {
		httpStatus = info.HTTPStatus
	}
	return &Error{
		Code:       code,
		Message:    message, // 重点是自定义这部分数据
		HttpStatus: httpStatus,
		Detail:     details,
		Internal:   internalErr,
	}
}

// IsError 检查错误是否为特定错误码
// TODO 有什么用？
func IsError(err error, code int) bool {
	// TODO 能否换成这样子的写法
	var e *Error
	if errors.As(err, &e) {
		return e.Code == code
	}
	//if e, ok := err.(*Error); ok{
	//	return e.Code == code
	//}
	// 压根也不能转换成这里的Error就肯定对应不上错误码
	return false
}

// GetHttpStatus 获取错误对应的http响应码
func GetHttpStatus(err error) int {
	var e *Error
	if errors.As(err, &e) {
		return e.HttpStatus
	}
	return http.StatusInternalServerError
}

// Wrap 包装标准错误为应用错误
func Wrap(err error, code int, details interface{}) *Error {
	if err == nil {
		return nil
	}
	// 如果已经是Error, 则保留原始错误码
	var e *Error
	if errors.As(err, &e) {
		return err.(*Error)
	}
	// 包装
	return New(code, details, err)
}

// UnWrap 实现errors.Unwrap接口，支持errors.Is和errors.As
func (e *Error) UnWrap() error {
	return e.Internal
}

// Is 检查错误是否为特定错误或其包装的错误
// 实现errors.Is接口的自定义行为
func (e *Error) Is(target error) bool {
	var t *Error
	if !errors.As(target, &t) {
		return false
	}
	return e.Code == t.Code
}
