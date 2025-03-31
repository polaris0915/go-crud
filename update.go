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

// Update 执行部分更新操作（PATCH）
// TODO 注意事项 在编写更新操作的钩子函数的时候，传入进去的是map[string]interface{}
func (c *Core[T]) Update() {
	// 1. 解析路径参数（获取资源 ID）
	id := cast.ToUint64(c.ginCtx.Param("id"))
	if id == 0 {
		c.err = cError.New(cError.ErrUpdateMissingField, nil, errors.New("缺少资源ID字段信息"))
		return
	}

	// 2. 检查资源是否存在
	existingModel := c.getModel()
	result := model.Use().Where("id = ?", id).First(&existingModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.err = cError.New(cError.ErrUpdateNotFound, nil, fmt.Errorf("ID: %d的资源不存在", id))
		} else {
			c.err = cError.New(cError.ErrDBQuery, nil, result.Error)
		}
		return
	}

	// 3. 绑定请求数据
	jsonMap := map[string]interface{}{}
	if err := c.ginCtx.ShouldBindJSON(&jsonMap); err != nil {
		c.err = cError.New(cError.ErrUpdateInvalidField, nil, errors.New("无效的请求数据格式"))
		return
	}

	// 如果请求体为空，返回错误
	if len(jsonMap) == 0 {
		c.err = cError.New(cError.ErrUpdateMissingField, nil, errors.New("请求体不能为空"))
		return
	}

	// 4. 检查请求数据中的字段是否都支持正常的部分更新操作
	// 获取支持部分更新的字段map
	modelMeta := getModelMeta(existingModel.TableName())
	if modelMeta == nil {
		c.err = cError.New(cError.ErrUpdateGeneral, nil, fmt.Errorf("无法获取模型为%s的元数据", existingModel.TableName()))
		return
	}

	partialUpdateFields := modelMeta.PartialUpdateFields
	if len(partialUpdateFields) == 0 { // 该模型不支持字段更新
		c.err = cError.New(cError.ErrUpdateInvalidField, nil, errors.New("该模型不支持部分更新"))
		return
	}

	//validFieldsMap := make(map[string]interface{})
	for field := range jsonMap {
		_, ok := modelMeta.PartialUpdateFields[field]
		if !ok {
			c.err = cError.New(cError.ErrUpdateInvalidField, nil, fmt.Errorf("字段 '%s' 不支持更新操作", field))
			return
		}
	}

	// 5. 权限检查会在中间件中进行处理

	//// 6. 将合法的所有字段信息映射到模型结构体中，以后用户去执行他们编写的更新前置钩子函数以及后置钩子函数
	//jsonModel := c.getModel()
	//if err := weakDecode(validFieldsMap, jsonModel); err != nil {
	//	c.err = cError.New(cError.ErrUpdateInvalidField, nil, errors.New("Update中执行weakDecode函数失败"))
	//	return
	//}

	// 7. 执行更新操作
	// 调用前置钩子（如果有）
	if c.beforeHook != nil {
		if err := c.beforeHook(jsonMap); err != nil {
			c.err = cError.New(cError.ErrUpdateHookFailure, nil, errors.New("更新前置钩子函数执行失败"))
			return
		}
	}

	// TODO 添加审计信息（暂时不用考虑，后续完成）
	// 添加审计字段（如果需要）
	// 这部分可以根据需要取消注释
	/*
		if _, hasUpdatedAt := validFieldsMap["updated_at"]; !hasUpdatedAt && modelMeta.HasField("updated_at") {
			validFieldsMap["updated_at"] = time.Now()
		}

		// 如果系统支持记录更新者，可以添加
		if _, hasUpdatedBy := validFieldsMap["updated_by"]; !hasUpdatedBy && modelMeta.HasField("updated_by") {
			// 从上下文获取当前用户ID（如果有）
			if userID, exists := c.ginCtx.Get("user_id"); exists {
				validFieldsMap["updated_by"] = userID
			}
		}
	*/

	// 开启事务（如果启用）
	tx := model.Use()
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

	// 执行更新操作
	// 将用户在钩子函数中操作完之后的jsonModel拿过去更新
	result = tx.Model(existingModel).Where("id = ?", id).Updates(jsonMap)
	if result.Error != nil {
		c.err = cError.New(cError.ErrUpdateGeneral, nil, result.Error)
		return
	}

	if result.RowsAffected == 0 {
		// TODO 如果这里需要告诉用户字段没有发生变化怎么编写响应信息合适？
	}

	// 获取更新后的资源
	updatedModel := c.getModel()
	if err := tx.Where("id = ?", id).First(updatedModel).Error; err != nil {
		c.err = cError.New(cError.ErrReadGeneral, nil, errors.New("无法获取更新后的资源"))
		return
	}

	// 调用后置钩子（如果有）
	if c.afterHook != nil {
		if err := c.afterHook(jsonMap); err != nil {
			c.err = cError.New(cError.ErrUpdateHookFailure, nil, errors.New("更新后置钩子函数执行失败"))
			return
		}
	}

	// 8. 返回结果
	HandleRes(c.ginCtx, http.StatusOK, updatedModel, "")
}
