package crud

import (
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/polaris0915/go-crud/cError"
	"github.com/polaris0915/go-crud/model"
	"github.com/spf13/cast"
	"net/http"
	"strings"
)

// TODO 需要添加添加获取字段信息的接口给用户

func (c *Core[T]) Get() {
	ctx := c.ginCtx

	// 解析请求参数
	id := cast.ToUint64(ctx.Param("id"))
	if id == 0 {
		c.err = cError.New(cError.ErrReadInvalidID, nil, errors.New("资源ID不能为空"))
		return
	}

	// 获取查询参数
	fields := ctx.Query("fields")
	expand := ctx.Query("expand")

	// 解析字段选择参数
	var requestedFields []string
	if fields != "" {
		requestedFields = strings.Split(fields, ",")
		for i := range requestedFields {
			requestedFields[i] = strings.TrimSpace(requestedFields[i])
		}
	}

	// 解析关联数据展开参数
	var expandRelations []string
	if expand != "" {
		expandRelations = strings.Split(expand, ",")
		for i := range expandRelations {
			expandRelations[i] = strings.TrimSpace(expandRelations[i])
		}
	}

	// 获取模型元数据
	modelMeta := getModelMeta(c.getModel().TableName())
	if modelMeta == nil {
		c.err = cError.New(cError.ErrReadGeneral, nil, errors.New("未找到模型元数据"))
		return
	}

	// 执行前置钩子
	if c.beforeHook != nil {
		// TODO
		if err := c.beforeHook(nil); err != nil {
			c.err = cError.New(cError.ErrReadHookFailure, nil, errors.New("查询前置钩子执行失败"))
			return
		}
	}

	// 检查读取字段的合法性
	for _, field := range requestedFields {
		_, ok := modelMeta.AllowGetFields[field]
		if !ok {
			c.err = cError.New(cError.ErrReadInvalidField, nil, fmt.Errorf("用户读取没有allow_get的字段%s", field))
			return
		}
	}

	// 构建查询
	db := model.Use()
	query := db.Table(c.getModel().TableName()).Where("id = ?", id).Limit(1)

	// 选择字段
	if len(requestedFields) == 0 { // 如果用户没有传入选择字段，那么默认返回所有allow_get的字段信息
		for field := range modelMeta.AllowGetFields {
			requestedFields = append(requestedFields, field)
		}
	}

	// 如果需要查询关联表的信息，则需要将外键信息查询出来
	associations := getModelMeta(c.getModel().TableName()).Associations
	foreignKeys := make([]string, 0)
	if len(expandRelations) > 0 {
		for _, foreignTableName := range expandRelations {
			_, ok := associations[foreignTableName]
			if !ok {
				c.err = cError.New(cError.ErrReadRelation, nil, fmt.Errorf("关联表%s不存在", foreignTableName))
				return
			}

			requestedFields = append(requestedFields, fmt.Sprintf("%s_id", foreignTableName))
			foreignKeys = append(foreignKeys, fmt.Sprintf("%s_id", strcase.ToSnake(foreignTableName)))
		}
	}

	query = query.Select(requestedFields)

	// 执行查询
	var result map[string]interface{}
	if err := query.Scan(&result).Error; err != nil {
		c.err = cError.New(cError.ErrDBQuery, nil, err)
		return
	}

	if result == nil {
		c.err = cError.New(cError.ErrReadNotFound, nil, fmt.Errorf("ID为%d的资源不存在", id))
		return
	}

	// 处理关联数据
	if len(expandRelations) > 0 {
		for _, foreignTableName := range expandRelations {
			foreignKey, ok := getModelMeta(c.getModel().TableName()).Associations[foreignTableName]
			if !ok {
				c.err = cError.New(cError.ErrReadRelation, nil, fmt.Errorf("关联表%s不存在", foreignTableName))
				return
			}
			if res, err := getForeignTableData(foreignTableName, cast.ToUint64(result[strcase.ToSnake(foreignKey)])); err != nil {
				// TODO 这里错误没有处理，因为这个表关联数据没有查询并不是一个非常致命的错误，因为前面主要的数据都查询到了
				result[foreignTableName] = nil
			} else {
				result[foreignTableName] = res
			}

		}
	}

	// 执行后置钩子
	if c.afterHook != nil {
		// TODO
		if err := c.afterHook(nil); err != nil {
			c.err = cError.New(cError.ErrReadHookFailure, "查询后置钩子执行失败", err)
			return
		}
	}

	for _, key := range foreignKeys {
		delete(result, key)
	}

	// 返回成功结果
	HandleRes(ctx, http.StatusOK, result, "")
}

func getForeignTableData(tableName string, id uint64) (data map[string]interface{}, err error) {
	// 构建查询
	db := model.Use()
	query := db.Table(tableName).Where("id = ?", id)

	modelMeta := getModelMeta(tableName)

	// 找出关联表中允许查询的字段
	fields := make([]string, 0)
	for field := range modelMeta.AllowGetFields {
		fields = append(fields, field)
	}
	if len(fields) == 0 {
		err = errors.New("关联表中没有可查询字段")
		return
	}

	// TODO 如果关联表还有关联别的表，就应该递归处理关联的关联的表的数据，现在暂时只处理了一层
	query = query.Select(fields)

	// 执行查询
	if err = query.Scan(&data).Error; err != nil {

		return
	}

	return
}
