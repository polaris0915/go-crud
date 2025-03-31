package crud

import (
	"github.com/gin-gonic/gin"
	"github.com/polaris0915/go-crud/model"
	"gorm.io/gorm"
	"reflect"
)

func InitCrud(db *gorm.DB, models ...CModel) {
	// 注册所有需要创建crud基本接口的模型
	register(models...)
	// 解析所有模型的元数据
	resolveModels()

	// 初始化自定义验证器
	initValidator()

	model.InitDB(db)
}

func RegisterModelApi[T CModel](r *gin.RouterGroup, preSuffix string, opts ...Option) {
	// 创建用户CRUD处理器
	crud := newCrud(
		// 模型工厂函数
		func() T {
			var m T
			// 如果 T 是指针类型，确保它被初始化
			modelType := reflect.TypeOf(m)
			if modelType.Kind() == reflect.Ptr {
				// 创建一个新的实例并返回其指针
				modelValue := reflect.New(modelType.Elem())
				return modelValue.Interface().(T)
			}
			return m
		},
		opts...,
	)

	registerRoutes[T](r, preSuffix, crud)
}

// registerRoutes 注册CRUD路由
func registerRoutes[T CModel](group *gin.RouterGroup, preSuffix string, crud *Crud[T]) {
	group.POST("/"+preSuffix, crud.Create()...)
	group.DELETE("/"+preSuffix+"/:id", crud.Delete()...)
	group.PATCH("/"+preSuffix+"/:id", crud.Update()...)
	group.GET("/"+preSuffix+"/:id", crud.Get()...)
}
