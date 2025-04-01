GetList 接口现在支持以下请求规则：

1. 分页参数：

- `page`：指定页码，默认为1
- `per_page`：每页记录数，默认为10，最大不超过100

1. 字段选择：

- `fields`：逗号分隔的字段列表
- 只能选择在模型中标记为 `allow_get` 的字段
- 未指定时返回所有 `allow_get` 字段

1. 排序：

- `sort_by`：指定排序字段
- `sort_order`：排序方向（`asc` 或 `desc`），默认 `desc`
- 只能按 `allow_get` 字段排序

1. 过滤查询：支持多种匹配方式：

- 精确匹配：直接使用 

  ```
  字段=值
  ```

  ```
  GET /api/user?status=1
  ```

- 模糊匹配（包含）：使用 

  ```
  like:
  ```

   前缀

  ```
  GET /api/user?address=like:江西
  ```

  查询 address 中包含 "江西" 的所有用户

- 开头匹配：使用 

  ```
  start:
  ```

   前缀

  ```
  GET /api/user?address=start:中国
  ```

  查询 address 以 "中国" 开头的所有用户

- 结尾匹配：使用 

  ```
  end:
  ```

   前缀

  ```
  GET /api/user?address=end:江西
  ```

  查询 address 以 "江西" 结尾的所有用户

1. 关联数据展开：

- `expand`：逗号分隔的关联表名
- 仅支持在模型中定义的关联关系
- 展开的关联数据会作为字段插入到每条记录中

完整示例：

```
GET /api/user?
    page=1&                 # 第1页
    per_page=10&            # 每页10条
    fields=username,email&  # 只返回username和email字段
    sort_by=age&            # 按年龄排序
    sort_order=desc&        # 降序排序
    status=1&               # 过滤状态为1的用户
    address=like:江西&      # 地址包含"江西"
    expand=role             # 展开角色关联数据
```

响应示例：

```
{
    "code": 200,
    "data": {
        "data": [
            {
                "username": "user1",
                "email": "user1@example.com",
                "role": {
                    "name": "普通用户",
                    "description": "系统普通用户角色"
                }
            }
        ],
        "pagination": {
            "total": 3,
            "per_page": 10,
            "current_page": 1,
            "total_pages": 1
        }
    }
}
```

限制和注意事项：

- 所有字段过滤和选择都基于模型的 `crud` 标签
- 过滤仅支持单个字段的匹配
- 不支持复杂的多条件组合查询
- 关联数据展开仅支持一层关联