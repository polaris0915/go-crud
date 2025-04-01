package crud

import (
	"github.com/gin-gonic/gin"
)

// HookFunc 定义db操作前的hook行为
type HookFunc func(core ICore) error

// Option 配置选项函数类型
type Option func(*Config)

type Config struct {
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

	GetListMiddlewares []gin.HandlerFunc
	BeforeGetList      HookFunc
	AfterGetList       HookFunc
}

// CreateMiddlewares 添加进入创建路由前的钩子，例如权限验证等
func CreateMiddlewares(handlers ...gin.HandlerFunc) Option {
	return func(c *Config) {
		c.CreateMiddlewares = append(c.CreateMiddlewares, handlers...)
	}
}

// DeleteMiddlewares 添加进入创建路由前的钩子，例如权限验证等
func DeleteMiddlewares(handlers ...gin.HandlerFunc) Option {
	return func(c *Config) {
		c.DeleteMiddlewares = append(c.DeleteMiddlewares, handlers...)
	}
}

// UpdateMiddlewares 添加进入创建路由前的钩子，例如权限验证等
func UpdateMiddlewares(handlers ...gin.HandlerFunc) Option {
	return func(c *Config) {
		c.UpdateMiddlewares = append(c.UpdateMiddlewares, handlers...)
	}
}

// GetMiddlewares 添加进入创建路由前的钩子，例如权限验证等
func GetMiddlewares(handlers ...gin.HandlerFunc) Option {
	return func(c *Config) {
		c.GetMiddlewares = append(c.GetMiddlewares, handlers...)
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

// BeforeDelete 添加创建数据前的钩子
func BeforeDelete(hook HookFunc) Option {
	return func(c *Config) {
		c.BeforeDelete = hook
	}
}

// AfterDelete 添加创建数据后的钩子
func AfterDelete(hook HookFunc) Option {
	return func(c *Config) {
		c.AfterDelete = hook
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

// BeforeGet 添加创建数据前的钩子
func BeforeGet(hook HookFunc) Option {
	return func(c *Config) {
		c.BeforeGet = hook
	}
}

// AfterGet 添加创建数据后的钩子
func AfterGet(hook HookFunc) Option {
	return func(c *Config) {
		c.AfterGet = hook
	}
}
