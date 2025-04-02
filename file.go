package crud

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/polaris0915/go-crud/cError"
	"github.com/polaris0915/go-crud/file_storage"
	"github.com/polaris0915/go-crud/model"
	"github.com/spf13/cast"
	"net/http"
	"path/filepath"
	"strings"
)

// 初始化文件存储
// storagePath := filepath.Join(".", "storage")
// fileStorage := file_storage.NewLocalFileStorage(storagePath)
// RegisterFileApi

func RegisterFileApi(r *gin.RouterGroup, storage file_storage.FileStorage, middlewares ...gin.HandlerFunc) {
	// 迁移模型到mysql
	if !model.Use().Migrator().HasTable(&model.File{}) {
		err := model.Use().Migrator().AutoMigrate(&model.File{})
		if err != nil {
			panic("crud中的文件模型迁移失败")
			return
		}
	}
	fileController := NewFileController(storage)

	fileGroup := r.Group("/file", middlewares...)
	{
		fileGroup.POST("/upload", fileController.UploadHandler)
		fileGroup.POST("/batch_upload", fileController.BatchUploadHandler)
		fileGroup.DELETE("/:id", fileController.DeleteHandler)
		fileGroup.POST("/batch_delete", fileController.BatchDeleteHandler)
		fileGroup.GET("/download/:id", fileController.DownloadHandler)
	}

}

// FileController 文件控制器
type FileController struct {
	storage file_storage.FileStorage
}

// NewFileController 创建文件控制器
func NewFileController(storage file_storage.FileStorage) *FileController {
	return &FileController{
		storage: storage,
	}
}

// UploadHandler 处理文件上传
func (fc *FileController) UploadHandler(c *gin.Context) {
	// 从上下文中获取当前用户ID

	//userID := c.GetInt64("user_id")
	t, _ := c.Get("user_id")
	userID := cast.ToUint64(t)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    cError.ErrUnauthorized,
			"message": "用户未登录",
		})
		return
	}

	// 获取相关实体信息（可选）
	relatedID := cast.ToUint64(c.PostForm("related_id"))
	relatedType := c.PostForm("related_type")

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    cError.ErrInvalidRequest,
			"message": "文件上传失败",
			"detail":  err.Error(),
		})
		return
	}

	// 检查文件类型和大小（可根据需求调整）
	if file.Size > 50*1024*1024 { // 例如：限制50MB
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    cError.ErrInvalidRequest,
			"message": "文件大小超过限制",
		})
		return
	}

	// 保存文件
	filePath, err := fc.storage.Save(file, relatedID, relatedType, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    cError.ErrInternal,
			"message": "文件保存失败",
			"detail":  err.Error(),
		})
		return
	}

	// 提取文件信息
	ext := filepath.Ext(file.Filename)
	fileType := getFileType(ext)

	// 创建文件记录
	fileModel := model.File{
		FileName:    filepath.Base(filePath),
		DisplayName: file.Filename,
		FileSize:    file.Size,
		FileType:    fileType,
		FilePath:    filePath,
		Uploader:    userID,
		RelatedID:   relatedID,
		RelatedType: relatedType,
	}

	db := model.Use()
	if err := db.Create(&fileModel).Error; err != nil {
		// 如果数据库创建失败，尝试删除已上传的文件
		_ = fc.storage.Delete(filePath)

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    cError.ErrCreateGeneral,
			"message": "文件记录创建失败",
			"detail":  err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "文件上传成功",
		"data":    fileModel,
	})
}

// BatchUploadHandler 处理批量文件上传
func (fc *FileController) BatchUploadHandler(c *gin.Context) {
	// 从上下文中获取当前用户ID
	userID := cast.ToUint64(c.GetInt64("user_id"))
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    cError.ErrUnauthorized,
			"message": "用户未登录",
		})
		return
	}

	// 获取相关实体信息（可选）
	relatedID := cast.ToUint64(c.PostForm("related_id"))
	relatedType := c.PostForm("related_type")

	// 获取上传的多个文件
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    cError.ErrInvalidRequest,
			"message": "获取表单数据失败",
			"detail":  err.Error(),
		})
		return
	}

	files := form.File["files[]"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    cError.ErrInvalidRequest,
			"message": "未上传任何文件",
		})
		return
	}

	// 用于存储上传成功的文件信息
	var uploadedFiles []model.File

	// 开启事务
	db := model.Use().Begin()
	if db.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    cError.ErrDBTransaction,
			"message": "开启事务失败",
		})
		return
	}

	// 处理每个文件
	for _, file := range files {
		// 检查文件大小
		if file.Size > 50*1024*1024 {
			db.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    cError.ErrInvalidRequest,
				"message": fmt.Sprintf("文件 %s 大小超过限制", file.Filename),
			})
			return
		}

		// 保存文件
		filePath, err := fc.storage.Save(file, relatedID, relatedType, userID)
		if err != nil {
			db.Rollback()
			// 清理已上传的文件
			for _, f := range uploadedFiles {
				_ = fc.storage.Delete(f.FilePath)
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    cError.ErrInternal,
				"message": fmt.Sprintf("文件 %s 保存失败", file.Filename),
				"detail":  err.Error(),
			})
			return
		}

		// 提取文件信息
		ext := filepath.Ext(file.Filename)
		fileType := getFileType(ext)

		// 创建文件记录
		fileModel := model.File{
			FileName:    filepath.Base(filePath),
			DisplayName: file.Filename,
			FileSize:    file.Size,
			FileType:    fileType,
			FilePath:    filePath,
			Uploader:    userID,
			RelatedID:   relatedID,
			RelatedType: relatedType,
		}

		// 保存到数据库
		if err := db.Create(&fileModel).Error; err != nil {
			db.Rollback()
			// 清理已上传的文件
			_ = fc.storage.Delete(filePath)
			for _, f := range uploadedFiles {
				_ = fc.storage.Delete(f.FilePath)
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    cError.ErrCreateGeneral,
				"message": fmt.Sprintf("文件 %s 记录创建失败", file.Filename),
				"detail":  err.Error(),
			})
			return
		}

		uploadedFiles = append(uploadedFiles, fileModel)
	}

	// 提交事务
	if err := db.Commit().Error; err != nil {
		db.Rollback()
		// 清理已上传的文件
		for _, f := range uploadedFiles {
			_ = fc.storage.Delete(f.FilePath)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    cError.ErrDBTransaction,
			"message": "提交事务失败",
			"detail":  err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": fmt.Sprintf("成功上传 %d 个文件", len(uploadedFiles)),
		"data":    uploadedFiles,
	})
}

// DeleteHandler 处理文件删除
func (fc *FileController) DeleteHandler(c *gin.Context) {
	// 获取文件ID
	fileID := cast.ToUint64(c.Param("id"))
	if fileID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    cError.ErrDeleteMissingField,
			"message": "无效的文件ID",
		})
		return
	}

	// 检查用户权限
	t, _ := c.Get("user_id")
	userID := cast.ToUint64(t)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    cError.ErrUnauthorized,
			"message": "用户未登录",
		})
		return
	}

	// 查询文件信息
	var fileModel model.File
	db := model.Use()
	if err := db.First(&fileModel, fileID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    cError.ErrDeleteNotFound,
			"message": "文件不存在",
		})
		return
	}

	// 可选：检查是否有删除权限(例如只能删除自己上传的文件)
	if fileModel.Uploader != userID {
		// 如果需要更严格的权限控制，可以在这里添加
		// 例如，检查用户是否为管理员
		t, _ := c.Get("user_role")
		userRole := t.(string)
		if userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    cError.ErrDeletePermission,
				"message": "没有权限删除此文件",
			})
			return
		}
	}

	// 开启事务
	tx := db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    cError.ErrDBTransaction,
			"message": "开启事务失败",
		})
		return
	}

	// 删除数据库记录
	if err := tx.Delete(&fileModel).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    cError.ErrDeleteGeneral,
			"message": "删除文件记录失败",
			"detail":  err.Error(),
		})
		return
	}

	// 删除物理文件
	if err := fc.storage.Delete(fileModel.FilePath); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    cError.ErrDeleteGeneral,
			"message": "删除物理文件失败",
			"detail":  err.Error(),
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    cError.ErrDBTransaction,
			"message": "提交事务失败",
			"detail":  err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "文件删除成功",
	})
}

// BatchDeleteHandler 处理批量文件删除
func (fc *FileController) BatchDeleteHandler(c *gin.Context) {
	// 获取要删除的文件ID列表
	var requestBody struct {
		FileIDs []uint64 `json:"file_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    cError.ErrInvalidRequest,
			"message": "无效的请求数据",
			"detail":  err.Error(),
		})
		return
	}

	if len(requestBody.FileIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    cError.ErrInvalidRequest,
			"message": "文件ID列表不能为空",
		})
		return
	}

	// 检查用户权限
	t, _ := c.Get("user_id")
	userID := cast.ToUint64(t)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    cError.ErrUnauthorized,
			"message": "用户未登录",
		})
		return
	}

	// 查询所有相关文件信息
	var files []model.File
	db := model.Use()
	if err := db.Where("id IN ?", requestBody.FileIDs).Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    cError.ErrDBQuery,
			"message": "查询文件信息失败",
			"detail":  err.Error(),
		})
		return
	}

	// 如果找不到任何文件，返回错误
	if len(files) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    cError.ErrDeleteNotFound,
			"message": "未找到指定的文件",
		})
		return
	}

	// 开启事务
	tx := db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    cError.ErrDBTransaction,
			"message": "开启事务失败",
		})
		return
	}

	// 记录成功和失败的文件
	var successFiles []uint64
	var failedFiles []struct {
		ID    uint64 `json:"id"`
		Error string `json:"error"`
	}

	// 处理每个文件
	for _, file := range files {
		// 可选：权限检查
		if file.Uploader != userID {

			t, _ := c.Get("user_role")
			userRole := t.(string)

			if userRole != "admin" {
				failedFiles = append(failedFiles, struct {
					ID    uint64 `json:"id"`
					Error string `json:"error"`
				}{
					ID:    file.ID,
					Error: "没有权限删除此文件",
				})
				continue
			}
		}

		// 删除数据库记录
		if err := tx.Delete(&file).Error; err != nil {
			failedFiles = append(failedFiles, struct {
				ID    uint64 `json:"id"`
				Error string `json:"error"`
			}{
				ID:    file.ID,
				Error: "删除数据库记录失败: " + err.Error(),
			})
			continue
		}

		// 删除物理文件
		if err := fc.storage.Delete(file.FilePath); err != nil {
			failedFiles = append(failedFiles, struct {
				ID    uint64 `json:"id"`
				Error string `json:"error"`
			}{
				ID:    file.ID,
				Error: "删除物理文件失败: " + err.Error(),
			})
			continue
		}

		// 记录成功
		successFiles = append(successFiles, file.ID)
	}

	// 如果所有操作都失败，回滚事务
	if len(successFiles) == 0 && len(failedFiles) > 0 {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    cError.ErrDeleteGeneral,
			"message": "所有文件删除失败",
			"detail":  failedFiles,
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    cError.ErrDBTransaction,
			"message": "提交事务失败",
			"detail":  err.Error(),
		})
		return
	}

	// 返回部分成功和部分失败的情况
	if len(failedFiles) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": fmt.Sprintf("成功删除 %d 个文件，%d 个文件删除失败", len(successFiles), len(failedFiles)),
			"data": gin.H{
				"success": successFiles,
				"failed":  failedFiles,
			},
		})
		return
	}

	// 全部成功
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": fmt.Sprintf("成功删除 %d 个文件", len(successFiles)),
		"data":    successFiles,
	})
}

// DownloadHandler 处理文件下载
func (fc *FileController) DownloadHandler(c *gin.Context) {
	// 获取文件ID
	fileID := cast.ToUint64(c.Param("id"))
	if fileID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    cError.ErrReadInvalidID,
			"message": "无效的文件ID",
		})
		return
	}

	// 查询文件信息
	var fileModel model.File
	db := model.Use()
	if err := db.First(&fileModel, fileID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    cError.ErrReadNotFound,
			"message": "文件不存在",
		})
		return
	}

	// 获取文件
	file, err := fc.storage.Get(fileModel.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    cError.ErrInternal,
			"message": "文件获取失败",
			"detail":  err.Error(),
		})
		return
	}
	defer file.Close()

	// 设置文件名
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileModel.DisplayName))

	// 设置内容类型
	contentType := getContentType(fileModel.FileType, filepath.Ext(fileModel.DisplayName))
	c.Header("Content-Type", contentType)

	// 发送文件
	c.File(file.Name())
}

// 辅助函数：根据扩展名确定文件类型
func getFileType(ext string) string {
	ext = strings.ToLower(ext)

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return "image"
	case ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".pdf":
		return "document"
	case ".mp3", ".wav", ".ogg", ".flac":
		return "audio"
	case ".mp4", ".avi", ".mkv", ".mov", ".wmv":
		return "video"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "archive"
	default:
		return "other"
	}
}

// 辅助函数：获取内容类型
func getContentType(fileType, ext string) string {
	ext = strings.ToLower(ext)

	// 常见MIME类型映射
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".mp3":  "audio/mpeg",
		".mp4":  "video/mp4",
		".zip":  "application/zip",
		".rar":  "application/x-rar-compressed",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}

	// 根据文件类型返回通用MIME类型
	switch fileType {
	case "image":
		return "image/jpeg"
	case "document":
		return "application/octet-stream"
	case "audio":
		return "audio/mpeg"
	case "video":
		return "video/mp4"
	case "archive":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}
