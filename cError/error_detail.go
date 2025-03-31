package cError

import "net/http"

// 通用错误
const (
	ErrInternal        = 1000 // 内部服务器错误
	ErrInvalidRequest  = 1001 // 无效请求
	ErrUnauthorized    = 1002 // 未授权
	ErrForbidden       = 1003 // 禁止访问
	ErrTimeout         = 1004 // 操作超时
	ErrTooManyRequests = 1005 // 请求过多
	ErrInvalidConfig   = 1006 // 无效配置
)

// 数据库错误
const (
	ErrDBConnection  = 2000 // 数据库连接错误
	ErrDBQuery       = 2001 // 数据库查询错误
	ErrDBExecution   = 2002 // 数据库执行错误
	ErrDBTransaction = 2003 // 数据库事务错误
	ErrDBLock        = 2004 // 数据库锁错误
	ErrDBTimeout     = 2005 // 数据库操作超时
	ErrDBConstraint  = 2006 // 数据库约束错误
)

// 创建操作错误
const (
	ErrCreateGeneral      = 3000 // 创建通用错误
	ErrCreateDuplicate    = 3001 // 创建重复记录
	ErrCreateValidation   = 3002 // 创建验证失败
	ErrCreateMissingField = 3003 // 创建缺少必填字段
	ErrCreateInvalidField = 3004 // 创建字段无效
	ErrCreateRelation     = 3005 // 创建关联错误
	ErrCreateHookFailure  = 3006 // 创建钩子函数失败
)

// 读取操作错误
const (
	ErrReadGeneral      = 4000 // 读取通用错误
	ErrReadNotFound     = 4001 // 记录不存在
	ErrReadPermission   = 4002 // 读取权限不足
	ErrReadInvalidID    = 4003 // 无效ID
	ErrReadFilter       = 4004 // 过滤条件错误
	ErrReadPagination   = 4005 // 分页参数错误
	ErrReadSort         = 4006 // 排序参数错误
	ErrReadRelation     = 4007 // 关联读取错误
	ErrReadHookFailure  = 4008 // 读取钩子函数失败
	ErrReadMissingField = 4009 // 读取缺少必填字段
	ErrReadInvalidField = 4010 // 读取无效字段
)

// 更新操作错误
const (
	ErrUpdateGeneral      = 5000 // 更新通用错误
	ErrUpdateNotFound     = 5001 // 更新目标不存在
	ErrUpdateValidation   = 5002 // 更新验证失败
	ErrUpdateConflict     = 5003 // 更新冲突
	ErrUpdateInvalidField = 5004 // 更新字段无效
	ErrUpdateRelation     = 5005 // 更新关联错误
	ErrUpdateConcurrency  = 5006 // 并发更新错误
	ErrUpdateHookFailure  = 5007 // 更新钩子函数失败
	ErrUpdateMissingField = 5008 // 更新缺少必填字段
)

// 删除操作错误
const (
	ErrDeleteGeneral      = 6000 // 删除通用错误
	ErrDeleteNotFound     = 6001 // 删除目标不存在
	ErrDeleteConstraint   = 6002 // 删除约束错误(如外键约束)
	ErrDeletePermission   = 6003 // 删除权限不足
	ErrDeleteRelation     = 6004 // 删除关联错误
	ErrDeleteProtected    = 6005 // 受保护记录不能删除
	ErrDeleteHookFailure  = 6006 // 删除钩子函数失败
	ErrDeleteMissingField = 6007 // 删除缺少必填字段
)

// 验证错误
const (
	ErrValidationGeneral   = 7000 // 验证通用错误
	ErrValidationRequired  = 7001 // 必填字段缺失
	ErrValidationFormat    = 7002 // 格式错误
	ErrValidationRange     = 7003 // 范围错误
	ErrValidationUnique    = 7004 // 唯一性错误
	ErrValidationReference = 7005 // 引用错误
	ErrValidationCustom    = 7006 // 自定义验证错误
)

// 业务规则错误
const (
	ErrBusinessGeneral    = 8000 // 业务规则通用错误
	ErrBusinessState      = 8001 // 状态错误
	ErrBusinessFlow       = 8002 // 流程错误
	ErrBusinessLimit      = 8003 // 限制错误
	ErrBusinessDependency = 8004 // 依赖错误
	ErrBusinessLogic      = 8005 // 逻辑错误
)

// 错误信息映射表
var errorMap = map[int]struct {
	Message    string
	HTTPStatus int
}{
	// 通用错误
	ErrInternal:        {"内部服务器错误", http.StatusInternalServerError},
	ErrInvalidRequest:  {"无效请求", http.StatusBadRequest},
	ErrUnauthorized:    {"未授权", http.StatusUnauthorized},
	ErrForbidden:       {"禁止访问", http.StatusForbidden},
	ErrTimeout:         {"操作超时", http.StatusGatewayTimeout},
	ErrTooManyRequests: {"请求过多", http.StatusTooManyRequests},
	ErrInvalidConfig:   {"无效配置", http.StatusInternalServerError},

	// 数据库错误
	ErrDBConnection:  {"数据库连接错误", http.StatusInternalServerError},
	ErrDBQuery:       {"数据库查询错误", http.StatusInternalServerError},
	ErrDBExecution:   {"数据库执行错误", http.StatusInternalServerError},
	ErrDBTransaction: {"数据库事务错误", http.StatusInternalServerError},
	ErrDBLock:        {"数据库锁错误", http.StatusInternalServerError},
	ErrDBTimeout:     {"数据库操作超时", http.StatusInternalServerError},
	ErrDBConstraint:  {"数据库约束错误", http.StatusBadRequest},

	// 创建操作错误
	ErrCreateGeneral:      {"创建资源失败", http.StatusInternalServerError},
	ErrCreateDuplicate:    {"资源已存在", http.StatusConflict},
	ErrCreateValidation:   {"创建验证失败", http.StatusBadRequest},
	ErrCreateMissingField: {"缺少必填字段", http.StatusBadRequest},
	ErrCreateInvalidField: {"字段值无效", http.StatusBadRequest},
	ErrCreateRelation:     {"关联创建失败", http.StatusBadRequest},
	ErrCreateHookFailure:  {"创建钩子执行失败", http.StatusInternalServerError},

	// 读取操作错误
	ErrReadGeneral:      {"读取资源失败", http.StatusInternalServerError},
	ErrReadNotFound:     {"资源不存在", http.StatusNotFound},
	ErrReadPermission:   {"无权读取资源", http.StatusForbidden},
	ErrReadInvalidID:    {"无效的资源ID", http.StatusBadRequest},
	ErrReadFilter:       {"无效的过滤条件", http.StatusBadRequest},
	ErrReadPagination:   {"无效的分页参数", http.StatusBadRequest},
	ErrReadSort:         {"无效的排序参数", http.StatusBadRequest},
	ErrReadRelation:     {"关联读取失败", http.StatusInternalServerError},
	ErrReadHookFailure:  {"读取钩子执行失败", http.StatusInternalServerError},
	ErrReadMissingField: {"读取缺少必填字段", http.StatusBadRequest},
	ErrReadInvalidField: {"读取无效字段", http.StatusBadRequest},

	// 更新操作错误
	ErrUpdateGeneral:      {"更新资源失败", http.StatusInternalServerError},
	ErrUpdateNotFound:     {"更新目标不存在", http.StatusNotFound},
	ErrUpdateValidation:   {"更新验证失败", http.StatusBadRequest},
	ErrUpdateConflict:     {"更新冲突", http.StatusConflict},
	ErrUpdateInvalidField: {"更新字段无效", http.StatusBadRequest},
	ErrUpdateRelation:     {"关联更新失败", http.StatusBadRequest},
	ErrUpdateConcurrency:  {"并发更新冲突", http.StatusConflict},
	ErrUpdateHookFailure:  {"更新钩子执行失败", http.StatusInternalServerError},
	ErrUpdateMissingField: {"缺少必填字段", http.StatusBadRequest},

	// 删除操作错误
	ErrDeleteGeneral:      {"删除资源失败", http.StatusInternalServerError},
	ErrDeleteNotFound:     {"删除目标不存在", http.StatusNotFound},
	ErrDeleteConstraint:   {"删除约束错误", http.StatusBadRequest},
	ErrDeletePermission:   {"无权删除资源", http.StatusForbidden},
	ErrDeleteRelation:     {"关联删除失败", http.StatusBadRequest},
	ErrDeleteProtected:    {"资源受保护，不能删除", http.StatusForbidden},
	ErrDeleteHookFailure:  {"删除钩子执行失败", http.StatusInternalServerError},
	ErrDeleteMissingField: {"缺少必填字段", http.StatusBadRequest},

	// 验证错误
	ErrValidationGeneral:   {"验证失败", http.StatusBadRequest},
	ErrValidationRequired:  {"必填字段缺失", http.StatusBadRequest},
	ErrValidationFormat:    {"字段格式错误", http.StatusBadRequest},
	ErrValidationRange:     {"字段值超出范围", http.StatusBadRequest},
	ErrValidationUnique:    {"字段值必须唯一", http.StatusBadRequest},
	ErrValidationReference: {"引用验证失败", http.StatusBadRequest},
	ErrValidationCustom:    {"自定义验证失败", http.StatusBadRequest},

	// 业务规则错误
	ErrBusinessGeneral:    {"业务规则错误", http.StatusBadRequest},
	ErrBusinessState:      {"状态错误", http.StatusBadRequest},
	ErrBusinessFlow:       {"流程错误", http.StatusBadRequest},
	ErrBusinessLimit:      {"超出限制", http.StatusBadRequest},
	ErrBusinessDependency: {"依赖错误", http.StatusBadRequest},
	ErrBusinessLogic:      {"业务逻辑错误", http.StatusBadRequest},
}
