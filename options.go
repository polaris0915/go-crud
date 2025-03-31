package crud

import (
	"github.com/gin-gonic/gin"
)

// HookFunc 定义db操作前的hook行为
type HookFunc func(any) error

// Option 配置选项函数类型
type Option func(*Config)

type Config struct {
	// EnableTransaction 是否需要开启事务
	EnableTransaction bool // TODO 事务控制需要优化

	// CreateMiddlewares 进入创建路由前的钩子
	CreateMiddlewares []gin.HandlerFunc
	// BeforeCreate 创建数据前的钩子
	BeforeCreate HookFunc
	// AfterCreate 创建数据后的钩子
	AfterCreate HookFunc

	DeleteMiddlewares []gin.HandlerFunc
	BeforeDelete      HookFunc
	AfterDelete       HookFunc

	UpdateMiddlewares []gin.HandlerFunc
	BeforeUpdate      HookFunc
	AfterUpdate       HookFunc

	GetMiddlewares []gin.HandlerFunc
	BeforeGet      HookFunc
	AfterGet       HookFunc
}

// EnableTransaction 是否开启事务
func EnableTransaction(open bool) Option {
	return func(c *Config) {
		c.EnableTransaction = open
	}
}

// CreateMiddlewares 添加进入创建路由前的钩子，例如权限验证等
func CreateMiddlewares(handlers ...gin.HandlerFunc) Option {
	return func(c *Config) {
		c.CreateMiddlewares = append(c.CreateMiddlewares, handlers...)
	}
}

// BeforeCreate 添加创建数据前的钩子
func BeforeCreate(hook HookFunc) Option {
	return func(c *Config) {
		c.BeforeCreate = hook
	}
}

// AfterCreate 添加创建数据后的钩子
func AfterCreate(hook HookFunc) Option {
	return func(c *Config) {
		c.AfterCreate = hook
	}
}

// BeforeUpdate 添加创建数据前的钩子
func BeforeUpdate(hook HookFunc) Option {
	return func(c *Config) {
		c.BeforeUpdate = hook
	}
}

// AfterUpdate 添加创建数据后的钩子
func AfterUpdate(hook HookFunc) Option {
	return func(c *Config) {
		c.AfterUpdate = hook
	}
}
