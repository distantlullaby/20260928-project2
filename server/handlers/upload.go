package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"taste-server/config"
	"taste-server/middleware"

	"github.com/gin-gonic/gin"
)

var allowedExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

// Upload 上传现场实拍图片，返回可访问 URL
func Upload(c *gin.Context) {
	_ = middleware.GetUserID(c) // 需要登录（路由上已挂 Auth）

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未收到上传文件"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExt[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 jpg/jpeg/png/gif/webp 图片"})
		return
	}
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片不能超过 10MB"})
		return
	}

	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	full := filepath.Join(config.UploadDir, name)
	if err := c.SaveUploadedFile(file, full); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + name})
}
