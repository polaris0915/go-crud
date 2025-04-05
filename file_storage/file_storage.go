package file_storage

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var localFileStorage *LocalFileStorage

func GetFileStorage() *LocalFileStorage {
	return localFileStorage
}

// FileStorage 文件存储接口
type FileStorage interface {
	Save(file *multipart.FileHeader, relateTypeID uint64, uploader uint64) (string, error)
	Delete(path string) error
	Get(path string) (*os.File, error)
}

// LocalFileStorage 本地文件存储实现
type LocalFileStorage struct {
	BasePath string // 基础存储路径
}

// NewLocalFileStorage 创建本地文件存储
func NewLocalFileStorage(basePath string) *LocalFileStorage {
	localFileStorage = &LocalFileStorage{
		BasePath: basePath,
	}
	return localFileStorage
}

// Save 保存文件到本地存储
func (s *LocalFileStorage) Save(file *multipart.FileHeader, relateTypeID uint64, uploader uint64) (string, error) {
	// 创建日期目录
	now := time.Now()
	datePath := now.Format("2006-01-02")
	dirPath := filepath.Join(s.BasePath, "uploads", datePath)

	// 确保目录存在
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成唯一文件名 (时间戳-原始文件名)
	uniqueFileName := fmt.Sprintf("%d-%s", now.UnixNano(), file.Filename)
	filePath := filepath.Join(dirPath, uniqueFileName)

	// 打开源文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开上传文件失败: %v", err)
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("保存文件失败: %v", err)
	}

	// 返回相对路径
	return filepath.Join("uploads", datePath, uniqueFileName), nil
}

// ValidatePath 在 LocalFileStorage 中添加一个验证路径的方法
func (s *LocalFileStorage) ValidatePath(path string) error {
	// 获取绝对路径
	fullPath := filepath.Join(s.BasePath, path)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("无法获取绝对路径: %v", err)
	}

	// 获取基础目录的绝对路径
	absBasePath, err := filepath.Abs(s.BasePath)
	if err != nil {
		return fmt.Errorf("无法获取基础目录的绝对路径: %v", err)
	}

	// 确保文件路径在基础目录内部（防止目录遍历攻击）
	if !strings.HasPrefix(absPath, absBasePath) {
		return errors.New("无效的文件路径")
	}

	return nil
}

// Delete 删除文件
func (s *LocalFileStorage) Delete(path string) error {
	// 验证路径
	if err := s.ValidatePath(path); err != nil {
		return err
	}

	// 继续删除操作...
	fullPath := filepath.Join(s.BasePath, path)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(fullPath)
}

// Get 获取文件
func (s *LocalFileStorage) Get(path string) (*os.File, error) {
	fullPath := filepath.Join(s.BasePath, path)
	return os.Open(fullPath)
}
