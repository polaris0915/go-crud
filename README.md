# Go-CRUD 框架使用指南

![Go Version](https://img.shields.io/badge/Go-1.18+-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)

Go-CRUD 是一个基于 Gin 和 GORM 的轻量级 CRUD 框架，专为快速开发 RESTful API 而设计。通过简单的模型定义，自动生成标准化的 API 接口，让您专注于业务逻辑，而非重复的 CRUD 代码编写。

🔗 **GitHub 地址**：[Go-CRUD](https://github.com/polaris0915/go-crud)

---

## ✨ 特性

- **模型驱动设计**：基于结构体标签自动生成 API。
- **钩子系统**：支持操作前后的自定义逻辑。
- **关联处理**：支持模型间的外键关联。
- **部分更新**：支持字段级别的 PATCH 操作。
- **灵活查询**：支持展开关联数据。
- **无缝集成**：兼容 Gin 和 GORM。

---

## 📦 安装

```sh
 go get -u github.com/polaris0915/go-crud
```

---

## 📋 快速上手

### 1️⃣ 准备数据库

在开始使用 Go-CRUD 之前，请确保您的数据库已就绪，并具有创建/修改表的权限。

### 2️⃣ 定义模型

Go-CRUD 通过结构体标签控制 CRUD 行为，示例如下：

```go
// 基础模型，包含通用字段
 type Model struct {
     ID        uint64          `gorm:"column:id;primary_key" json:"id"`
     CreatedAt time.Time       `gorm:"column:created_at" json:"created_at"`
     UpdatedAt time.Time       `gorm:"column:updated_at" json:"updated_at"`
     DeletedAt *gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
 }

// 角色模型
type Role struct {
    Model
    Role string `gorm:"column:role;type:varchar(30);unique;not null;comment:user role" json:"role" crud:"required_on_create,allow_get,partial_update"`
}

// 用户模型
type User struct {
    Model
    Name     string  `gorm:"column:name;type:varchar(50);not null;unique" json:"name" crud:"required_on_create,allow_get"`
    Email    string  `gorm:"column:email;type:varchar(255);not null;uniqueIndex" json:"email" crud:"required_on_create"`
    Password string  `gorm:"column:password;type:varchar(255);not null" json:"password" crud:"required_on_create"`
    RoleID   *uint64 `gorm:"column:role_id;type:bigint" json:"role_id"`
    Role     Role    `gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"role"`
}
```

> **📌 注意：** 每个模型都必须实现 `TableName()` 方法，以指定数据库表名。

#### 🌟 CRUD 标签说明

| 标签 | 作用 |
|------|------|
| `required_on_create` | 创建时该字段必须填写 |
| `partial_update` | 允许使用 PATCH 方法更新该字段 |
| `allow_get` | 允许通过 GET 方法获取该字段 |

---

### 3️⃣ 注册模型与路由

```go
func main() {
    // 初始化数据库
    InitMysql()
    
    // 初始化 Gin 路由
    router := gin.Default()
    r := router.Group("/api")
    
    // 注册模型
    crud.InitCrud(db, &Role{}, &User{})
    crud.RegisterModelApi[*Role](r, "/role")
    crud.RegisterModelApi[*User](r, "/user", BeforeCreate())
    
    router.Run(":8080")
}
```

---

## 🚀 API 端点

### 🌐 标准 API 结构

| 方法   | 路径               | 描述         | 查询参数说明                     |
|--------|-------------------|--------------|----------------------------------|
| GET    | /api/{path}/:id   | 获取单个资源 | `fields=字段1,字段2`（指定返回字段）<br>`expand=关联字段`（展开关联数据） |
| POST   | /api/{path}       | 创建资源     | -                                |
| PATCH  | /api/{path}/:id   | 部分更新资源 | -                                |
| DELETE | /api/{path}/:id   | 删除资源     | -                                |

### 🔍 查询参数示例

1️⃣ **查询指定字段**
```sh
GET /api/user/1?fields=name,email
```
📌 **返回仅包含** `name` 和 `email` 字段。

2️⃣ **展开关联数据**
```sh
GET /api/user/1?expand=role
```
📌 **返回用户数据时，附带其角色信息**。

---

## ⚠️ 注意事项

### ✅ 模型规范

- **必须**实现 `TableName()` 方法。
- GORM 标签应详细指定 `column` 属性，使用 **下划线命名法**。
- JSON 标签需与数据库字段保持一致。

### ✅ 路由注册

- **必须**在分组路由上注册 CRUD API。
- 避免路径末尾带 `/`，如 `/api/user`。

### ✅ 关联数据

- 使用 `gorm:"foreignKey:XXX"` 正确定义外键。
- 通过 `expand` 参数展开关联数据。

### ✅ 部分更新

- 仅支持标记了 `partial_update` 的字段。

---

## 📜 许可证

本项目采用 **MIT** 许可证。

Go-CRUD 框架正在持续优化，欢迎提交 Issue 或 Pull Request 🚀！

