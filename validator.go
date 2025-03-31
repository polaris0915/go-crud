package crud

import (
	"github.com/go-playground/validator/v10"
	"reflect"
)

var v *validator.Validate

// RequireOnCreate 验证字段在创建操作时是否为非零值
func RequireOnCreate(fl validator.FieldLevel) bool {
	field := fl.Field()
	// 判断字段是否为零值
	switch field.Kind() {
	case reflect.String:
		return field.String() != ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return field.Float() != 0
	case reflect.Bool:
		return field.Bool()
	case reflect.Slice, reflect.Map, reflect.Array:
		return field.Len() > 0
	case reflect.Ptr, reflect.Interface:
		return !field.IsNil()
	case reflect.Struct:
		return !field.IsZero() // Go 1.13+ 支持
	default:
		// TODO: 需要打印日志输出
		return true // 对于不支持的类型，默认返回 true
	}
}

func initValidator() {
	//v = binding.Validator.Engine().(*validator.Validate)
	v = validator.New()

	v.RegisterValidation("required_on_create", RequireOnCreate)
}

func UseValidator() *validator.Validate {
	return v
}
