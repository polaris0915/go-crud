package crud

import (
	"github.com/gin-gonic/gin"
	"github.com/polaris0915/go-crud/cError"
)

type Core[T CModel] struct {
	// gin的Context上下文
	ginCtx *gin.Context
	// 具体模型工厂函数
	getModel func() T
	// 当前请求的错误，如果最终错误不为空，就会返回错误
	err *cError.Error
	// 是否开启事务
	enableTransaction bool

	// 增删改查操作执行之前的钩子函数
	beforeHook HookFunc
	// 增删改查操作执行之后的钩子函数
	afterHook HookFunc

	// 请求参数
	payload map[string]interface{}
	// 校验标签以及规则
	rules map[string]interface{}
}

// NewCore 实例化最终操作对象
func NewCore[T CModel](
	ginCtx *gin.Context, getModel func() T, enableTransaction bool,
	beforeHook HookFunc, afterHook HookFunc,
	rules map[string]interface{},
) (c *Core[T]) {

	c = &Core[T]{
		ginCtx:            ginCtx,
		getModel:          getModel,
		enableTransaction: enableTransaction,

		beforeHook: beforeHook,
		afterHook:  afterHook,
		payload:    make(map[string]interface{}),
		rules:      rules,
	}
	return
}

// HandleRes 全局响应处理函数
func HandleRes(c *gin.Context, code int, data interface{}, message string) {
	// TODO 根据data以及message数据的有无，来指定具体的响应形式，减少传输数据的成本
	response := gin.H{
		"code": code,
	}

	if message != "" {
		response["message"] = message
	}

	if data != nil {
		response["data"] = data
	}

	c.JSON(code, response)
}
