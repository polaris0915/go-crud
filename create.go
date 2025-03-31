package crud

import (
	"errors"
	"github.com/polaris0915/go-crud/cError"
	"github.com/polaris0915/go-crud/model"
	"net/http"
)

func (c *Core[T]) Create() {
	// 1. 获取新的模型T的对象
	jsonModel := c.getModel()

	// 3. 绑定请求数据
	if err := c.ginCtx.ShouldBindJSON(&c.payload); err != nil {
		// TODO detail字段可以更加详细
		// TODO 内部错误需要优化
		c.err = cError.New(cError.ErrCreateMissingField, nil, errors.New("请求参数解析错误"))
		return
	}

	// TODO 根据tag检查字段
	if errs := UseValidator().ValidateMap(c.payload, c.rules); len(errs) > 0 {
		c.err = cError.New(cError.ErrCreateMissingField, nil, errors.New("请求参数解析错误"))
		return
	}

	// 4. 检查唯一性约束
	if err := checkUniqueness(func() CModel { return jsonModel }, c.payload); err != nil {
		// 数据重复
		if errors.Is(err, errDataDuplicated) {
			c.err = cError.New(cError.ErrCreateDuplicate, nil, errDataDuplicated)
			return
		}
		// 不是数据重复的错误
		c.err = cError.New(cError.ErrCreateGeneral, nil, err)
		return
	}

	err := weakDecode(c.payload, &jsonModel)
	if err != nil {
		c.err = cError.New(cError.ErrCreateGeneral, nil, err)
		return
	}

	// 创建前置钩子
	if c.beforeHook != nil {
		if err := c.beforeHook(jsonModel); err != nil {
			c.err = cError.New(cError.ErrCreateHookFailure, nil, errors.New("创建前置钩子函数执行失败"))
			return
		}
	}

	// 开启事务（如果启用）
	var tx = model.Use()
	if c.enableTransaction {
		tx = model.Use().Begin()
		if tx.Error != nil {
			c.err = cError.New(cError.ErrDBTransaction, nil, tx.Error)
			return
		}
		defer func() {
			if c.err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()
	}

	// 执行创建操作
	result := tx.Create(jsonModel)
	if result.Error != nil {
		c.err = cError.New(cError.ErrCreateGeneral, nil, result.Error)
		return
	}

	// 后置钩子
	if c.afterHook != nil {
		if err := c.afterHook(jsonModel); err != nil {
			c.err = cError.New(cError.ErrCreateHookFailure, nil, errors.New("更新后置钩子函数执行失败"))
			return
		}
	}

	// 返回成功响应
	HandleRes(c.ginCtx, http.StatusCreated, true, "")
}
