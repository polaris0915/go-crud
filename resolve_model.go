package crud

import (
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
)

// empty 仅做一个占位，表示这个字段在这个要求中需要
var empty struct{}

// 存储所有需要注册的模型
var collection []CModel

var registeredModels = make(map[string]*RegisteredModel)

// Fields 存储注册模型的字段信息
type Fields struct {
	Name          string
	GormFieldName string // 在数据库中的字段名，在gorm标签中的column的值
	Type          reflect.Type
	Unique        bool   // 是否唯一
	Default       string // 字段默认值

	GormTag string
	JsonTag string
	CrudTag string
}

// RegisteredModel 存储已经注册的模型信息
type RegisteredModel struct {
	ModelName             string
	Fields                []*Fields
	Rules                 map[string]map[string]interface{}
	RequireOnCreateFields map[string]struct{}
	PartialUpdateFields   map[string]struct{}
	AllowGetFields        map[string]struct{}

	// Associations 存储关联关系的所有信息
	// 例如 User表关联Role表
	// 数据形式为: map["role"] = "RoleID"
	Associations map[string]string
}

func register(model ...CModel) {
	collection = append(collection, model...)
}

func deepResolve(r *RegisteredModel, m reflect.Type) {
	r.Rules["create"] = make(map[string]interface{})
	r.Rules["update"] = make(map[string]interface{})
	r.Rules["get"] = make(map[string]interface{})

	// 解析模型的所有字段
	for i := 0; i < m.NumField(); i++ {
		// 获取字段
		field := m.Field(i)
		//// 获取字段类型
		//fieldType := field.Type

		modelFields := &Fields{
			Name:    field.Name,
			Type:    field.Type,
			JsonTag: field.Tag.Get("json"),
			GormTag: field.Tag.Get("gorm"),
			CrudTag: field.Tag.Get("crud"),
		}

		//if modelFields.CrudTag != "" {
		//	r.Rules[modelFields.JsonTag] = modelFields.CrudTag
		//}

		// 根据binding:"partial_update"标签，解析出部分更新时所涉及到的字段
		if modelFields.CrudTag != "" {
			crudTags := strings.Split(modelFields.CrudTag, ",")

			for _, tag := range crudTags {
				if tag == "required_on_create" {
					r.RequireOnCreateFields[modelFields.JsonTag] = empty
					r.Rules["create"][modelFields.JsonTag] = "required_on_create"
				}
				if tag == "partial_update" {
					r.PartialUpdateFields[modelFields.JsonTag] = empty
					r.Rules["update"][modelFields.JsonTag] = "partial_update"
				}
				if tag == "allow_get" {
					r.AllowGetFields[modelFields.JsonTag] = empty
					r.Rules["get"][modelFields.JsonTag] = "allow_get"
				}
			}
		}

		// 获取gorm标签，检查当前字段是否是唯一或者有默认值等
		if modelFields.GormTag != "" {
			tags := strings.Split(modelFields.GormTag, ";")
			for _, tag := range tags {
				// 获取所有的列名
				if strings.Contains(tag, "column") {
					columnName := strings.Split(tag, ":")
					if len(columnName) == 2 {
						modelFields.GormFieldName = columnName[1]
					}
				}
				if strings.Contains(tag, "unique") {
					modelFields.Unique = true
				}
				if strings.Contains(tag, "default") {
					defaultValue := strings.Split(tag, ":")
					if len(defaultValue) == 2 {
						modelFields.Default = defaultValue[1]
					}
				}
				if strings.Contains(tag, "foreignKey") {
					if r.Associations == nil {
						r.Associations = make(map[string]string)
					}
					foreignKey := strings.Split(tag, ":")
					if len(foreignKey) == 2 {
						if tableName, ok := strings.CutSuffix(foreignKey[1], "ID"); ok {
							r.Associations[strcase.ToSnake(tableName)] = foreignKey[1]
						} else {
							panic("解析模型关联关系出错")
						}
					}
				}
			}
		}
		r.Fields = append(r.Fields, modelFields)
	}

}

func resolveModels() {
	for _, model := range collection {
		// 解析模型元数据
		m := reflect.TypeOf(model).Elem()
		r := &RegisteredModel{
			ModelName:             model.TableName(),
			Fields:                make([]*Fields, 0, 10),
			Rules:                 make(map[string]map[string]interface{}),
			RequireOnCreateFields: make(map[string]struct{}),
			PartialUpdateFields:   make(map[string]struct{}),
			AllowGetFields:        make(map[string]struct{}),
		}
		// 深度解析
		deepResolve(r, m)
		registeredModels[model.TableName()] = r
	}
}

func getModelMeta(modelName string) *RegisteredModel {
	return registeredModels[modelName]
}
