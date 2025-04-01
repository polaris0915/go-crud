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

// GetList 执行列表查询操作
func (c *Core[T]) GetList() {
	ctx := c.ginCtx

	// 1. 解析分页参数
	page := cast.ToInt(ctx.DefaultQuery("page", "1"))
	perPage := cast.ToInt(ctx.DefaultQuery("per_page", "10"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	// 2. 解析字段选择参数
	fields := ctx.Query("fields")
	expand := ctx.Query("expand")

	// 3. 解析排序参数
	sortBy := ctx.Query("sort_by")
	sortOrder := ctx.DefaultQuery("sort_order", "desc")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// 4. 解析过滤参数
	filterParams := make(map[string]interface{})
	// 直接从 Query 中获取参数
	for key, values := range ctx.Request.URL.Query() {
		if len(values) > 0 {
			filterParams[key] = values[0]
		}
	}

	// 调试：打印所有查询参数
	//log.Printf("查询参数: %+v", filterParams)

	// 调试：打印所有查询参数
	//log.Printf("查询参数: %+v", filterParams)

	// 5. 获取模型元数据
	modelMeta := getModelMeta(c.getModel().TableName())
	if modelMeta == nil {
		c.err = cError.New(cError.ErrReadGeneral, nil, errors.New("未找到模型元数据"))
		return
	}

	// 6. 执行前置钩子
	if c.beforeHook != nil {
		if err := c.beforeHook(c); err != nil {
			c.err = cError.New(cError.ErrReadHookFailure, nil, errors.New("列表查询前置钩子执行失败"))
			return
		}
	}

	// 7. 解析并验证字段选择
	var requestedFields []string
	if fields != "" {
		requestedFields = strings.Split(fields, ",")
		for i := range requestedFields {
			requestedFields[i] = strings.TrimSpace(requestedFields[i])
			if _, ok := modelMeta.AllowGetFields[requestedFields[i]]; !ok {
				c.err = cError.New(cError.ErrReadInvalidField, nil, fmt.Errorf("不允许获取字段: %s", requestedFields[i]))
				return
			}
		}
	}

	// 8. 如果没有指定字段，使用所有允许获取的字段
	if len(requestedFields) == 0 {
		for field := range modelMeta.AllowGetFields {
			requestedFields = append(requestedFields, field)
		}
	}

	// 9. 准备数据库查询
	db := model.Use().Table(c.getModel().TableName())

	// 10. 处理过滤条件
	for key, value := range filterParams {
		// 调试：打印每个过滤条件
		//log.Printf("处理过滤条件 - 字段: %s, 值: %v", key, value)

		// 检查是否是允许的字段
		if _, ok := modelMeta.AllowGetFields[key]; !ok {
			//log.Printf("跳过不允许的过滤字段: %s", key)
			continue // 跳过不允许的过滤字段
		}

		// 将值转换为字符串
		strValue, ok := value.(string)
		if !ok {
			// 如果不是字符串，使用精确匹配
			//log.Printf("非字符串值，使用精确匹配: %s = %v", key, value)
			db = db.Where(fmt.Sprintf("%s = ?", key), value)
			continue
		}

		// 忽略空字符串
		if strValue == "" {
			//log.Printf("忽略空字符串: %s", key)
			continue
		}

		// 支持特殊的查询前缀
		if strings.HasPrefix(strValue, "like:") {
			// 使用 LIKE 查询
			likeValue := strings.TrimPrefix(strValue, "like:")
			//log.Printf("模糊匹配: %s LIKE %s", key, likeValue)
			db = db.Where(fmt.Sprintf("%s LIKE ?", key), fmt.Sprintf("%%%s%%", likeValue))
		} else if strings.HasPrefix(strValue, "start:") {
			// 开头匹配
			startValue := strings.TrimPrefix(strValue, "start:")
			//log.Printf("开头匹配: %s LIKE %s", key, startValue)
			db = db.Where(fmt.Sprintf("%s LIKE ?", key), fmt.Sprintf("%s%%", startValue))
		} else if strings.HasPrefix(strValue, "end:") {
			// 结尾匹配
			endValue := strings.TrimPrefix(strValue, "end:")
			//log.Printf("结尾匹配: %s LIKE %s", key, endValue)
			db = db.Where(fmt.Sprintf("%s LIKE ?", key), fmt.Sprintf("%%%s", endValue))
		} else {
			// 默认精确匹配
			//log.Printf("精确匹配: %s = %s", key, strValue)
			db = db.Where(fmt.Sprintf("%s = ?", key), value)
		}
	}

	// 11. 处理排序
	if sortBy != "" {
		if _, ok := modelMeta.AllowGetFields[sortBy]; !ok {
			c.err = cError.New(cError.ErrReadSort, nil, fmt.Errorf("不允许按字段 %s 排序", sortBy))
			return
		}
		db = db.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
	}

	// 12. 计算总记录数
	var total int64
	countDB := db
	if err := countDB.Count(&total).Error; err != nil {
		c.err = cError.New(cError.ErrDBQuery, nil, err)
		return
	}

	// 13. 执行分页查询
	offset := (page - 1) * perPage
	db = db.Select(requestedFields).Offset(offset).Limit(perPage)

	// 14. 查询结果
	var results []map[string]interface{}
	if err := db.Find(&results).Error; err != nil {
		c.err = cError.New(cError.ErrDBQuery, nil, err)
		return
	}

	// 15. 处理关联数据展开
	var expandRelations []string
	if expand != "" {
		expandRelations = strings.Split(expand, ",")
		for _, relation := range expandRelations {
			relation = strings.TrimSpace(relation)
			foreignKey, ok := modelMeta.Associations[relation]
			if !ok {
				c.err = cError.New(cError.ErrReadRelation, nil, fmt.Errorf("关联表 %s 不存在", relation))
				return
			}

			for i, result := range results {
				if id, ok := result[fmt.Sprintf("%s_id", strcase.ToSnake(foreignKey))]; ok {
					if relationData, err := getForeignTableData(relation, cast.ToUint64(id)); err == nil {
						results[i][relation] = relationData
					}
				}
			}
		}
	}

	// 16. 执行后置钩子
	if c.afterHook != nil {
		if err := c.afterHook(c); err != nil {
			c.err = cError.New(cError.ErrReadHookFailure, nil, errors.New("列表查询后置钩子执行失败"))
			return
		}
	}

	// 17. 准备分页信息
	pagination := model.Pagination{
		Total:       total,
		PerPage:     perPage,
		CurrentPage: page,
		TotalPages:  model.TotalPage(total, perPage),
	}

	// 18. 返回结果
	HandleRes(ctx, http.StatusOK, model.DataList{
		Data:       results,
		Pagination: pagination,
	}, "")
}
