package api

import (
	"net/http"
	"net/url"
	"path/filepath"

	"fujiantong/internal/service"

	"github.com/gin-gonic/gin"
)

type CHandler struct {
	Svc *service.Container
}

func NewCHandler(svc *service.Container) *CHandler {
	return &CHandler{Svc: svc}
}

// GetFileByCode GET /api/v1/c/file/:code
// C 端 H5 下载页加载文件元信息（含作者昵称）
func (h *CHandler) GetFileByCode(c *gin.Context) {
	code := c.Param("code")
	f, err := h.Svc.GetFileByCode(code)
	if err != nil {
		Fail(c, 404, "文件不存在或已下架")
		return
	}
	if f.MediaCheckStatus == "reject" {
		Fail(c, 403, "该文件未通过内容审核")
		return
	}

	// 计 view（异步）
	go h.Svc.FileRepo.IncrViewCount(f.ID)

	// 取作者昵称
	author, _ := h.Svc.UserRepo.GetByID(f.AuthorUserID)
	authorName := ""
	if author != nil {
		authorName = author.Nickname
	}

	OK(c, gin.H{
		"id":             f.ID,
		"name":           f.OriginalName,
		"ext":            f.Ext,
		"size":           f.Size,
		"download_code":  f.DownloadCode,
		"view_count":     f.ViewCount + 1,
		"download_count": f.DownloadCount,
		"author_name":    authorName,
		"created_at":     f.CreatedAt,
	})
}

// DownloadFile GET /api/v1/c/download/:code
// 直接 streaming 下载文件，自增 download_count
func (h *CHandler) DownloadFile(c *gin.Context) {
	code := c.Param("code")
	f, err := h.Svc.GetFileByCode(code)
	if err != nil {
		c.String(http.StatusNotFound, "文件不存在或已下架")
		return
	}
	if f.MediaCheckStatus == "reject" {
		c.String(http.StatusForbidden, "该文件未通过内容审核")
		return
	}

	// 自增下载计数（异步）
	go h.Svc.FileRepo.IncrDownloadCount(f.ID)

	// MVP 阶段本地存储，直接返回文件
	if f.StorageType == "local" {
		filename := f.OriginalName
		// RFC 5987 编码处理中文名
		c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(filename))
		c.Header("Content-Type", f.MimeType)
		c.File(filepath.Clean(f.StoragePath))
		return
	}

	// V2.0 后接入 COS：302 重定向到 COS 预签名 URL
	c.String(http.StatusNotImplemented, "COS 存储尚未启用")
}

// GetLaunchInfo GET /api/v1/c/launch/:code
// 返回拉起作者小程序所需的全套链接（appid/path/scheme/url_link）
// 用于 H5 兜底页主动跳小程序，确保平台依然能获得广告变现
func (h *CHandler) GetLaunchInfo(c *gin.Context) {
	code := c.Param("code")
	info, err := h.Svc.GenerateLaunchInfo(code)
	if err != nil {
		Fail(c, 200, err.Error()) // 200 + 业务码：前端拿不到则降级为 H5 直接下载
		return
	}
	OK(c, info)
}

// ReportUV POST /api/v1/c/uv
// UV 上报（小程序端调用），用于推广员申请门槛检测
func (h *CHandler) ReportUV(c *gin.Context) {
	// MVP 阶段简单记录，后续可扩展为按 day+file 去重
	OKMsg(c, "ok")
}
