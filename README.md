# Go-CRUD

![Go Version](https://img.shields.io/badge/Go-1.18+-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)

Go-CRUD是一个轻量级、高效的Go语言CRUD框架，专为快速开发RESTful API而设计。通过简单的模型定义，自动生成标准化的增删改查接口，让您专注于业务逻辑而非重复的CRUD代码编写。

## ✨ 特性

- **模型驱动设计**：基于结构体标签自动生成API
- **统一错误处理**：标准化的错误响应格式
- **灵活钩子系统**：支持操作前后的自定义逻辑
- **事务支持**：确保数据一致性
- **字段验证**：内置验证规则
- **关联处理**：支持模型间关联关系
- **与Gin和GORM无缝集成**：兼容流行的Web框架和ORM

## 📦 安装

```bash
go get github.com/polaris0915/go-crud
```

## 🚀 快速开始

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/polaris0915/go-crud"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

type User struct {
    ID    uint   `json:"id" gorm:"primaryKey"`
    Name  string `json:"name" crud:"required"`
    Email string `json:"email" crud:"required,email"`
}

func main() {
    // 初始化数据库
    db, _ := gorm.Open(mysql.Open("dsn"), &gorm.Config{})
    db.AutoMigrate(&User{})
    
    // 初始化CRUD框架
    crud.InitCrud(db)
    
    // 设置路由
    r := gin.Default()
    api := r.Group("/api")
    
    userCrud := crud.NewCrud(func() interface{} { return &User{} })
    
    api.POST("/users", userCrud.Create()...)
    api.GET("/users/:id", userCrud.Get()...)
    api.PATCH("/users/:id", userCrud.Update()...)
    api.DELETE("/users/:id", userCrud.Delete()...)
    api.GET("/users", userCrud.List()...)
    
    r.Run(":8080")
}
```

## 📋 主要功能

- 自动生成CRUD操作
- 请求参数验证
- 钩子函数支持
- 事务管理
- 统一错误处理
- 关联关系处理
- 灵活的查询选项

## 📄 许可证

本项目采用MIT许可证。

---

更多详细文档和使用示例将在后续版本中提供。如有任何问题或建议，请提交Issue或Pull Request。
