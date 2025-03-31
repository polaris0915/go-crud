package crud

import (
	"errors"
	"fmt"
	"github.com/polaris0915/go-crud/cError"
	"github.com/polaris0915/go-crud/model"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"net/http"
)

func (c *Core[T]) Delete() {
	// 1. 解析路径参数，获取资源ID
	id := cast.ToUint64(c.ginCtx.Param("id"))
	if id == 0 {
		c.err = cError.New(cError.ErrDeleteMissingField, nil, errors.New("缺少资源ID字段信息"))
		return
	}

	// 2. 检查资源是否存在
	jsonModel := c.getModel()
	db := model.Use()

	// 查询记录是否存在
	result := db.First(&jsonModel, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.err = cError.New(cError.ErrDeleteNotFound, nil, fmt.Errorf("ID: %s的资源不存在", id))
		} else {
			c.err = cError.New(cError.ErrDBQuery, nil, result.Error)
		}
		return
	}

	// 3. 执行前置钩子（可用于权限检查和业务规则验证）
	if c.beforeHook != nil {
		if err := c.beforeHook(jsonModel); err != nil {
			c.err = cError.New(cError.ErrDeleteHookFailure, nil, errors.New("删除前置钩子函数执行失败"))
			return
		}
	}

	// 4. 处理事务
	// 开启事务（如果启用）
	if c.enableTransaction {
		db = model.Use().Begin()
		if db.Error != nil {
			c.err = cError.New(cError.ErrDBTransaction, nil, db.Error)
			return
		}
		defer func() {
			if c.err != nil {
				db.Rollback()
			} else {
				db.Commit()
			}
		}()
	}

	// 5. 执行删除操作（软删除）
	// 假设模型已经实现了gorm.Model或包含DeletedAt字段
	result = db.Delete(&jsonModel)
	if result.Error != nil {
		c.err = cError.New(cError.ErrDeleteGeneral, nil, result.Error)
		return
	}

	if result.RowsAffected == 0 {
		// 这种情况通常不会发生，因为我们已经检查了记录是否存在
		// 但为了健壮性，仍然处理这种情况
		c.err = cError.New(cError.ErrDeleteGeneral, nil, errors.New("删除操作未影响任何记录"))
		return
	}

	// 6. 执行后置钩子（可用于清理相关资源、发送通知等）
	if c.afterHook != nil {
		if err := c.afterHook(jsonModel); err != nil {
			c.err = cError.New(cError.ErrDeleteHookFailure, nil, errors.New("删除后置钩子函数执行失败"))
			return
		}
	}

	// 7. 返回结果
	HandleRes(c.ginCtx, http.StatusNoContent, true, "")
}
