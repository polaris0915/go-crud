package crud

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/polaris0915/go-crud/model"
	"gorm.io/gorm"
)

var (
	errReflectUniField = errors.New("反射解析唯一字段出错")
	errDataDuplicated  = errors.New("数据重复")
	errUnknown         = errors.New("未知错误")
)

// checkUniqueness 检查当前请求数据中的唯一属性的字段是否已经在数据库中存在了
func checkUniqueness(getModel func() CModel, payload map[string]interface{}) error {
	// 获取需要检查的模型
	// TODO 这边得验证这个jsonModel是值传递还是引用传递，如果是引用传递下面的First就会出问题
	jsonModel := getModel()
	// 获取模型元数据
	modelMeta := getModelMeta(jsonModel.TableName())

	for _, field := range modelMeta.Fields {
		if field.Unique {
			var err error
			val := payload[field.Name]
			if err =
				model.Use().Model(jsonModel).Where(fmt.Sprintf("%s = ?", field.GormFieldName), val).
					Limit(1).First(jsonModel).Error; err == nil {
				// 数据已经被创建过
				// TODO 输出日志
				return errDataDuplicated
			}
			// 数据没有被创建过
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}

			//// TODO 输出日志
			return errUnknown
		}
	}
	return nil
}

// weakDecode decodes the input data to the output data with weakly typed input
func weakDecode(input, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
		DecodeHook:       mapstructure.ComposeDecodeHookFunc(),
		TagName:          "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}
