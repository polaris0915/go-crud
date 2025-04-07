// Package log
// 有关zap的基本使用可以参考以下链接:
// https://juejin.cn/post/7215025208410734648
// Debug、Info、Warn、Error、DPanic、Panic、Fatal
// 相对于zap的标准日志输出，目前只实现了 Debug, Info, Warn, Error 四种（也包括含有用户上下文的）。
// 日志的输出位置，如果在初始化时，没有指定Config中的Path参数，就会去环境变量中找CRUD_LOG_PATH的值作为输出路径，
// 如果这个路径不存在，那么就会使用当前程序运行的目录作为日志的输出目录。
// 默认日志级别是Info。
// 目前支持从gin.Context中获取用户信息，提取Context中的"user_id","user_role","user_ip"。
// 当前日志系统还存在一下问题:
// 1. getLogPath 中没有处理MkdirAll的错误
// 2. Write 现在实现日志的时间轮转是在每次写入日志前，检查当前时间是否是新的一天，可以使用goroutine去监控当前时间，提升性能
// 3. createNewLogFile 在创建新的日志文件的时候没有检查是否到达了磁盘的空间限制
// 4. cleanOldLogs 处理旧文件发生错误时使用fmt.Printf进行输出的
// 5. SetLogLevel 并发情况下不安全，无法确认其他goroutine中是否还在访问logger全局对象，因此现在的版本中尽量不要去使用
// 6. Sync 目前时直接忽略了错误
// 7. Write 如果单次系统运行写入的日志信息不足MaxSize，当下次系统运行时，不能正确的将此时文件的大小记录到dailyLogWriter中的currentSize
package log

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const logPathEnv = "CRUD_LOG_PATH"
const logsDir = "log/logs"

var (
	logger *zap.Logger
	cfg    Config
	once   sync.Once
)

type Config struct {
	Level          string `json:"level" yaml:"level"`                     // 日志级别: debug, info, warn, error
	Format         string `json:"format" yaml:"format"`                   // 日志格式: json, console
	Path           string `json:"path" yaml:"path"`                       // 日志路径
	MaxSize        int    `json:"max_size" yaml:"max_size"`               // 单个日志文件最大大小(MB)
	MaxAge         int    `json:"max_age" yaml:"max_age"`                 // 日志保留天数
	Compress       bool   `json:"compress" yaml:"compress"`               // 是否压缩
	ConsoleLogging bool   `json:"console_logging" yaml:"console_logging"` // 是否同时输出到控制台
}

var defaultConfig = Config{
	Level:          "info",
	Format:         "json",
	Path:           "",
	MaxSize:        100,
	MaxAge:         30,
	Compress:       false,
	ConsoleLogging: true,
}

// getLogPath 获取系统日志路径
// TODO 没有处理MkdirAll的错误
func getLogPath() (logPath string) {
	if cfg.Path != "" {
		logPath = cfg.Path
	} else {
		logPath = os.Getenv(logPathEnv)
		if logPath == "" {
			execPath, _ := os.Executable()
			// 当前执行程序的目录路径 + "/logs"
			logPath = filepath.Join(filepath.Dir(execPath), logsDir)
		}
	}

	os.MkdirAll(logPath, os.ModePerm)
	return logPath
}

// parseLevel 将字符串转换为 zapcore.Level
// TODO 默认日志处理级别是Info
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// dailyLogWriter 实现zapcore.WriteSyncer
type dailyLogWriter struct {
	mu           sync.Mutex
	logPath      string
	file         *os.File
	lastDate     string
	maxSizeBytes int64
	currentSize  int64
	maxAge       int
	// TODO 没有实现日志压缩功能
	compress bool
	// 当文件大小超出限制，来表示当前文件是当天的第几个文件
	fileCount int
}

func newDailyLogWriter(logPath string, maxSizeMB int, maxAge int, compress bool) *dailyLogWriter {
	if maxSizeMB <= 0 {
		maxSizeMB = defaultConfig.MaxSize
	}
	return &dailyLogWriter{
		logPath:      logPath,
		maxSizeBytes: int64(maxSizeMB) * 1024 * 1024,
		maxAge:       maxAge,
		compress:     compress,
	}
}

// Sync 实现zapcore.WriteSyncer
func (w *dailyLogWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}

// Write 实现zapcore.WriteSyncer
func (w *dailyLogWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// TODO 测试使用
	//testDate1, _ := time.Parse("2006-01-02", "2025-04-08")
	//setMockTime(testDate1)
	//now := mockNow()

	// 获取当前日期
	now := time.Now()
	currentDay := now.Format("2006-01-02")

	// 日期变更或文件未打开，创建新文件
	// TODO 每次写入日志数据的时候都会检查这个日期是否是新日期，可能会影响性能，可以考虑使用一个goroutine来检查日期变更
	if currentDay != w.lastDate || w.file == nil {
		if w.file != nil { // 日期变更，先关闭之前的日志文件
			oldFile := w.file
			w.file = nil // 防止重复关闭
			oldFile.Close()
		}

		// 如果配置中有过期删除配置，就要检查过期文件，执行删除
		if w.maxAge > 0 {
			w.cleanOldLogs(now)
		}

		w.lastDate = currentDay
		w.fileCount = 0
		w.currentSize = 0
		return w.createNewLogFile(p, currentDay)
	}

	// 不需要新建，直接写入当前日志文件
	w.currentSize += int64(len(p))
	if w.maxSizeBytes > 0 && w.currentSize >= w.maxSizeBytes {
		oldFile := w.file
		w.file = nil
		oldFile.Close()

		w.fileCount++
		w.currentSize = int64(len(p))
		return w.createNewLogFile(p, currentDay)
	}
	// 直接写入，没有发生日期的变更，获取当前文件没超过限制
	return w.file.Write(p)
}

// createNewLogFile 创建新的日志文件
// 两种情况会触发，一个是日志单个文件穿出配置大小，另一个是新的日期
// TODO 在创建时，没有检查是否到达了磁盘的空间限制
// TODO 待测试
func (w *dailyLogWriter) createNewLogFile(p []byte, currentDay string) (n int, err error) {
	var fileName string
	if w.fileCount > 0 {
		fileName = fmt.Sprintf("%s.%d.log", currentDay, w.fileCount)
	} else {
		fileName = fmt.Sprintf("%s.log", currentDay)
	}
	logPath := filepath.Join(w.logPath, fileName)
	w.file, err = os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return 0, err
	}

	// 写入此时的数据，计算大小
	n, err = w.file.Write(p)
	if err == nil {
		w.currentSize += int64(n)
	}

	return n, err
}

// cleanOldLogs 清理过期日志 TODO 待测试
// TODO 目前版本是以日志文件名来清理过期日志的，处理就文件发生错误时是使用fmt.Printf进行输出的
func (w *dailyLogWriter) cleanOldLogs(now time.Time) {
	cutoffDate := now.AddDate(0, 0, -w.maxAge)
	cutoffDateStr := cutoffDate.Format("2006-01-02")

	files, err := os.ReadDir(w.logPath)
	if err != nil {
		fmt.Printf("读取日志文件目录错误: %v\n", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		// YYYY-MM-DD 正确的日期格式大小为10个字符
		if matched, _ := filepath.Match("[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]*.log", name); !matched {
			continue
		}

		fileDate := name[:10]
		if fileDate < cutoffDateStr { // 如果小于最早的过期时间的文件就要移除
			if err := os.Remove(filepath.Join(w.logPath, name)); err != nil {
				fmt.Printf("移除过期日志%s失败: %v\n", name, err)
			}
		}
	}

}

// 初始化基本配置
func newLogger(config Config) (*zap.Logger, error) {
	//logPath := getLogPath()
	//// TODO 测试使用
	//logPath := "/Users/alin-youlinlin/Desktop/polaris-all_projects/polaris-backend-go/crud/log"
	cfg = config
	logPath := getLogPath()

	dailyLogWriter := newDailyLogWriter(logPath, cfg.MaxSize, cfg.MaxAge, cfg.Compress)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		StacktraceKey: "stacktrace",
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "message",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		// 配置时间格式
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 根据配置选择编码器
	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 配置输出
	var writerSyncers []zapcore.WriteSyncer
	writerSyncers = append(writerSyncers, dailyLogWriter)

	// 根据配置查看是否需要控制台输出
	if cfg.ConsoleLogging {
		writerSyncers = append(writerSyncers, zapcore.Lock(os.Stdout))
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(
			writerSyncers...,
		),
		zap.NewAtomicLevelAt(parseLevel(cfg.Level)),
	)

	// 添加Error及以上级别日志的堆栈信息，已经溯源到日志生成的地方
	//  zap.AddCallerSkip(1),
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel)), nil
}

// InitLogger 初始化日志系统
func InitLogger(config ...Config) {
	once.Do(func() {
		var err error

		// 使用默认配置或者传入配置
		c := defaultConfig
		if len(config) > 0 {
			c = config[0]
		}

		logger, err = newLogger(c)
		if err != nil {
			panic(err)
		}
	})
}

// GetLogger 获取logger实例
func GetLogger() *zap.Logger {
	if logger == nil {
		InitLogger()
	}
	return logger
}

// loggerMutex 防止并发时更新全局日志对象的级别出错
var loggerMutex sync.Mutex

// SetLogLevel 设置日志处理等级
// TODO 注意：并发情况下还是不要使用这个函数去重新设置日志的级别
// 虽然目前已经在替换全局logger的时候做了锁的处理，避免了竞态问题，但是我们没有等待别的并发任务中完成对旧的logger的访问
// 因此如果贸然替换掉，会出现意外的错误
// 考虑可以用原子操作来处理这个情况
func SetLogLevel(level string) {
	zapLevel := parseLevel(level)

	// 加锁，更新配置
	loggerMutex.Lock()
	cfg.Level = zapLevel.String()

	// 重新创建logger
	newLogger, err := newLogger(cfg)
	if err == nil {
		// 执行替换，先要刷新
		oldLogger := logger
		logger = newLogger
		loggerMutex.Unlock()

		if oldLogger != nil {
			_ = oldLogger.Sync()
		}
	} else {
		loggerMutex.Unlock()
		fmt.Printf("设置日志Level失败: %v\n", err)
	}
}

// SetCallerSkip 设置跳过调用函数的层数，0表示不跳过
func SetCallerSkip(skip int) {
	logger = logger.WithOptions(zap.AddCallerSkip(skip))
}

// Sync 刷新日志缓冲区
// TODO 目前版本是直接忽略了错误
func Sync() {
	_ = logger.Sync()
}

// handleGinContext 从gin框架中的上下文Context中提取用户相关信息
// TODO 目前支持从gin.Context中获取用户信息，提取Context中的"user_id","user_role","user_ip"
func handleGinContext(c context.Context) (context map[string]interface{}) {
	ginCtx, ok := c.(*gin.Context)
	if !ok {
		return nil
	}

	context = make(map[string]interface{})
	if userID, ok := ginCtx.Get("user_id"); ok {
		context["user_id"] = userID
	}

	if userRole, ok := ginCtx.Get("user_role"); ok {
		context["user_role"] = userRole
	}

	if userIp, ok := ginCtx.Get("user_ip"); ok {
		context["user_ip"] = userIp
	}

	return
}

func Debug(msg string, fields ...zap.Field) {
	if logger == nil {
		InitLogger()
	}
	logger.Debug(msg, fields...)
}

func DebugWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	if logger == nil {
		InitLogger()
	}

	userInfo := handleGinContext(ctx)
	if userInfo != nil {
		logger.With(zap.Any("user", userInfo)).Debug(msg, fields...)
		return
	}
	logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	if logger == nil {
		InitLogger()
	}
	logger.Info(msg, fields...)
}

func InfoWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	if logger == nil {
		InitLogger()
	}
	userInfo := handleGinContext(ctx)
	if userInfo != nil {
		logger.With(zap.Any("user", userInfo)).Info(msg, fields...)
		return
	}
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	if logger == nil {
		InitLogger()
	}
	logger.Warn(msg, fields...)
}

func WarnWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	if logger == nil {
		InitLogger()
	}
	userInfo := handleGinContext(ctx)
	if userInfo != nil {
		logger.With(zap.Any("user", userInfo)).Warn(msg, fields...)
		return
	}
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	if logger == nil {
		InitLogger()
	}
	logger.Error(msg, fields...)
}

func ErrorWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	if logger == nil {
		InitLogger()
	}
	userInfo := handleGinContext(ctx)
	if userInfo != nil {
		logger.With(zap.Any("user", userInfo)).Error(msg, fields...)
		return
	}
	logger.Error(msg, fields...)
}
