package api

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"fujiantong/internal/middleware"
	"fujiantong/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthorHandler struct {
	Svc *service.Container
}

func NewAuthorHandler(svc *service.Container) *AuthorHandler {
	return &AuthorHandler{Svc: svc}
}

// GetProfile GET /api/v1/author/profile
func (h *AuthorHandler) GetProfile(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	user, err := h.Svc.GetUserProfile(uint64(uid))
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, user)
}

// ListFiles GET /api/v1/author/files
func (h *AuthorHandler) ListFiles(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	page, size := pageParams(c)
	list, total, err := h.Svc.FileRepo.ListByAuthor(uint64(uid), page, size)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, gin.H{"list": list, "total": total, "page": page, "size": size})
}

// UploadFile POST /api/v1/author/files/upload
func (h *AuthorHandler) UploadFile(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	user, err := h.Svc.GetUserProfile(uint64(uid))
	if err != nil {
		FailErr(c, err)
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Fail(c, 400, "请选择文件上传")
		return
	}
	defer file.Close()

	// 本地存储
	dir := fmt.Sprintf("./uploads/%d/%s", uid, time.Now().Format("2006-01"))
	_ = os.MkdirAll(dir, 0755)
	dst := filepath.Join(dir, header.Filename)
	out, err := os.Create(dst)
	if err != nil {
		FailErr(c, err)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		FailErr(c, err)
		return
	}

	f, err := h.Svc.CreateFile(
		uint64(uid),
		user.BoundAppID,
		header.Filename,
		dst,
		header.Header.Get("Content-Type"),
		header.Size,
	)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, f)
}

// ReplaceFile PUT /api/v1/author/files/:id
func (h *AuthorHandler) ReplaceFile(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	fileID := atoui64(c.Param("id"))

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Fail(c, 400, "请选择文件")
		return
	}
	defer file.Close()

	dir := fmt.Sprintf("./uploads/%d/%s", uid, time.Now().Format("2006-01"))
	_ = os.MkdirAll(dir, 0755)
	dst := filepath.Join(dir, header.Filename)
	out, err := os.Create(dst)
	if err != nil {
		FailErr(c, err)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		FailErr(c, err)
		return
	}

	if err := h.Svc.ReplaceFile(fileID, uint64(uid), dst, header.Filename, header.Header.Get("Content-Type"), header.Size); err != nil {
		FailErr(c, err)
		return
	}
	OKMsg(c, "替换成功")
}

// GenerateShareCard POST /api/v1/author/files/:id/share-card
func (h *AuthorHandler) GenerateShareCard(c *gin.Context) {
	fileID := atoui64(c.Param("id"))
	cards, err := h.Svc.GenerateShareCards(fileID)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, cards)
}

// SubmitAppID POST /api/v1/author/appid
func (h *AuthorHandler) SubmitAppID(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var req struct {
		AppID string `json:"appid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误")
		return
	}
	if err := h.Svc.SubmitAppID(uint64(uid), req.AppID); err != nil {
		FailErr(c, err)
		return
	}
	OKMsg(c, "appid 绑定成功")
}

// GetRevenue GET /api/v1/author/revenue
func (h *AuthorHandler) GetRevenue(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	user, err := h.Svc.GetUserProfile(uint64(uid))
	if err != nil {
		FailErr(c, err)
		return
	}
	if user.BoundAppID == "" {
		Fail(c, 400, "尚未绑定小程序")
		return
	}
	page, size := pageParams(c)
	list, total, err := h.Svc.FinanceRepo.GetSettlements(user.BoundAppID, page, size)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, gin.H{"list": list, "total": total})
}
