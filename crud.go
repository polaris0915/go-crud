package crud

import (
	"github.com/gin-gonic/gin"
)

// CModel 定义所有表模型的行为
type CModel interface {
	TableName() string
}

type ICrud interface {
	Create() []gin.HandlerFunc
	Delete() []gin.HandlerFunc
	Update() []gin.HandlerFunc
	Get() []gin.HandlerFunc
	GetList() []gin.HandlerFunc
}

// Crud
// Crud[T CModel]这里这样子写，只有实现了 CModel 接口的模型才能调用 Create() 等方法
type Crud[T CModel] struct {
	// GetModel 获取模型实例的工厂函数
	GetModel func() T
	// config 保存当前模型所有的执行钩子
	// 例如 创建前 创建后等等
	config Config
}

func newCrud[T CModel](getModel func() T, opts ...Option) *Crud[T] {
	// 默认配置
	config := Config{}

	// 配置创建 删除 修改 查询的钩子函数以及额外的配置等
	for _, opt := range opts {
		opt(&config)
	}

	return &Crud[T]{
		config:   config,
		GetModel: getModel,
	}
}

// Create 实例化单个创建函数
func (c *Crud[T]) Create() (ginHandlers []gin.HandlerFunc) {
	//var ginHandlers []gin.HandlerFunc
	// 添加路由中间件
	ginHandlers = append(ginHandlers, c.config.CreateMiddlewares...)
	// 添加实际路由执行函数
	ginHandlers = append(
		ginHandlers,
		func(ginCtx *gin.Context) {
			// 实例化核心对象
			core := NewCore[T](
				ginCtx, c.GetModel, // 根据c.config来配置路由
				c.config.BeforeCreate, c.config.AfterCreate, // 创建前置钩子，猴子钩子
				getModelMeta(c.GetModel().TableName()).Rules["create"], // 校验规则
			)
			// 执行创建函数
			core.Create()
			// 如果有错误，组织错误响应
			if core.err != nil {
				ginCtx.JSON(core.err.HttpStatus, gin.H{
					"code":    core.err.Code,
					"message": core.err.Message,
				})
				return
			}
		})
	return ginHandlers
}

// Delete 实例化单个软删除函数
func (c *Crud[T]) Delete() (ginHandlers []gin.HandlerFunc) {
	//var ginHandlers []gin.HandlerFunc
	// 添加路由中间件
	ginHandlers = append(ginHandlers, c.config.DeleteMiddlewares...)
	// 添加实际路由执行函数
	ginHandlers = append(
		ginHandlers,
		func(ginCtx *gin.Context) {
			// 实例化核心对象
			core := NewCore[T](
				ginCtx, c.GetModel,
				c.config.BeforeDelete, c.config.AfterDelete,
				getModelMeta(c.GetModel().TableName()).Rules["delete"],
			)
			// 执行创建函数
			core.Delete()
			// 如果有错误，组织错误响应
			if core.err != nil {
				ginCtx.JSON(core.err.HttpStatus, gin.H{
					"code":    core.err.Code,
					"message": core.err.Message,
				})
				return
			}
		})
	return ginHandlers
}

// Update 实例化单个更新函数
func (c *Crud[T]) Update() (ginHandlers []gin.HandlerFunc) {
	//var ginHandlers []gin.HandlerFunc
	// 添加路由中间件
	ginHandlers = append(ginHandlers, c.config.UpdateMiddlewares...)
	// 添加实际路由执行函数
	ginHandlers = append(
		ginHandlers,
		func(ginCtx *gin.Context) {
			// 实例化核心对象
			core := NewCore[T](
				ginCtx, c.GetModel,
				c.config.BeforeUpdate, c.config.AfterUpdate,
				getModelMeta(c.GetModel().TableName()).Rules["update"],
			)
			// 执行创建函数
			core.Update()
			// 如果有错误，组织错误响应
			if core.err != nil {
				ginCtx.JSON(core.err.HttpStatus, gin.H{
					"code":    core.err.Code,
					"message": core.err.Message,
				})
				return
			}
		})
	return ginHandlers
}

func (c *Crud[T]) Get() (ginHandlers []gin.HandlerFunc) {
	//var ginHandlers []gin.HandlerFunc
	// 添加路由中间件
	ginHandlers = append(ginHandlers, c.config.GetMiddlewares...)
	// 添加实际路由执行函数
	ginHandlers = append(
		ginHandlers,
		func(ginCtx *gin.Context) {
			// 实例化核心对象
			core := NewCore[T](
				ginCtx, c.GetModel,
				c.config.BeforeGet, c.config.AfterGet,
				getModelMeta(c.GetModel().TableName()).Rules["get"],
			)
			// 执行创建函数
			core.Get()
			// 如果有错误，组织错误响应
			if core.err != nil {
				ginCtx.JSON(core.err.HttpStatus, gin.H{
					"code":    core.err.Code,
					"message": core.err.Message,
				})
				return
			}
		})
	return ginHandlers
}

func (c *Crud[T]) GetList() (ginHandlers []gin.HandlerFunc) {
	//var ginHandlers []gin.HandlerFunc
	// 添加路由中间件
	ginHandlers = append(ginHandlers, c.config.GetListMiddlewares...)
	// 添加实际路由执行函数
	ginHandlers = append(
		ginHandlers,
		func(ginCtx *gin.Context) {
			// 实例化核心对象
			core := NewCore[T](
				ginCtx, c.GetModel,
				c.config.BeforeGetList, c.config.AfterGetList,
				getModelMeta(c.GetModel().TableName()).Rules["get"],
			)
			// 执行创建函数
			core.GetList()
			// 如果有错误，组织错误响应
			if core.err != nil {
				ginCtx.JSON(core.err.HttpStatus, gin.H{
					"code":    core.err.Code,
					"message": core.err.Message,
				})
				return
			}
		})
	return ginHandlers
}
