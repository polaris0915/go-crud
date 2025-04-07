package log

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"os"
	"sync"
	"testing"
	"time"
)

// TODO 测试使用
// 全局时间管理器，用于测试
var globalTimeManager = struct {
	mu       sync.RWMutex
	mockTime *time.Time
}{
	mockTime: nil,
}

// 替换 time.Now() 的函数
func mockNow() time.Time {
	globalTimeManager.mu.RLock()
	defer globalTimeManager.mu.RUnlock()

	if globalTimeManager.mockTime != nil {
		return *globalTimeManager.mockTime
	}
	return time.Now()
}

// 设置模拟时间
func setMockTime(t time.Time) {
	globalTimeManager.mu.Lock()
	defer globalTimeManager.mu.Unlock()
	globalTimeManager.mockTime = &t
}

// 重置为实际时间
func resetMockTime() {
	globalTimeManager.mu.Lock()
	defer globalTimeManager.mu.Unlock()
	globalTimeManager.mockTime = nil
}

func TestENV(t *testing.T) {
	value := os.Getenv("CRUD_LOG_PATH")
	fmt.Println("CRUD_LOG_PATH:", value) // 使用 fmt.Println 直接打印
	if value == "" {
		t.Error("CRUD_LOG_PATH is empty")
	}
}

func TestLog(t *testing.T) {
	// 使用默认配置初始化
	//InitLogger()

	// 或者使用自定义配置
	InitLogger(Config{
		Level:  "info", // 日志级别: debug, info, warn, error
		Format: "json", // 日志格式: json, console
		//Path:           "/Users/alin-youlinlin/Desktop/polaris-all_projects/polaris-backend-go/crud/log", // 日志路径
		MaxSize:        1,    // 单个日志文件最大大小(MB)
		MaxAge:         30,   // 日志保留天数
		Compress:       true, // 是否压缩旧日志
		ConsoleLogging: true, // 是否同时输出到控制台 TODO 没有实现
	})

	SetCallerSkip(1)

	// 基本用法
	Info("系统启动成功", zap.Int("port", 8080))
	Debug("调试信息", zap.String("config_file", "config.yaml"))
	Warn("警告信息", zap.String("resource", "database"), zap.Int("retry", 3))
	Error("错误信息", zap.Error(errors.New("连接数据库失败")))

	// 带上下文的日志
	ginCtx := &gin.Context{} // 实际使用中这通常来自中间件
	ginCtx.Set("user_id", "12345")
	ginCtx.Set("user_role", "admin")
	ginCtx.Set("user_ip", "192.168.1.1")

	InfoWithContext(ginCtx, "用户登录成功")
	ErrorWithContext(ginCtx, "操作失败", zap.String("action", "delete_user"))

	for i := 0; i < 2000; i++ {
		InfoWithContext(ginCtx, "用户登录成功")
		ErrorWithContext(ginCtx, "操作失败", zap.String("action", "delete_user"))
	}

	//// 创建具有特定字段的 logger
	//orderLogger := log.WithFields(zap.String("module", "order"), zap.String("version", "v1.0"))
	//orderLogger.Info("订单处理开始")
	//orderLogger.Error("订单处理失败", zap.String("order_id", "ORD-12345"))

	// 动态修改日志级别
	//log.SetLogLevel("warn") // 运行时修改为只输出警告及以上级别的日志

	// 确保退出前刷新缓冲区
	defer Sync()
}
