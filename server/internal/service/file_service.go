package service

import (
	"fmt"
	"html"
	"path/filepath"
	"strings"

	"fujiantong/internal/model"
	"fujiantong/pkg/utils"
)

// CreateFile 创建文件记录 + 触发内容审核（V2.8 §15.2.2）
func (c *Container) CreateFile(authorUserID uint64, appID, originalName, storagePath, mimeType string, size int64) (*model.File, error) {
	ext := strings.TrimPrefix(filepath.Ext(originalName), ".")
	code := utils.GenerateShortCode(8)
	name := strings.TrimSuffix(originalName, filepath.Ext(originalName))

	// 按文件类型判定初始审核状态
	// 图片/音频可走 mediaCheckAsync 机审；文档类需主理人人工复审
	initialStatus := mediaCheckInitialStatus(mimeType)

	f := &model.File{
		AuthorUserID:     authorUserID,
		AppID:            appID,
		Name:             name,
		OriginalName:     originalName,
		Ext:              ext,
		Size:             size,
		MimeType:         mimeType,
		StorageType:      "local",
		StoragePath:      storagePath,
		DownloadCode:     code,
		MediaCheckStatus: initialStatus,
		Status:           "active",
	}
	if err := c.FileRepo.Create(f); err != nil {
		return nil, err
	}

	// 异步触发机审（仅图片/音频）— 失败不阻塞上传流程
	if initialStatus == "reviewing" && appID != "" {
		go c.triggerMediaCheck(f, appID)
	}

	return f, nil
}

// mediaCheckInitialStatus 根据 MIME 类型决定初始审核状态
func mediaCheckInitialStatus(mimeType string) string {
	mt := strings.ToLower(mimeType)
	switch {
	case strings.HasPrefix(mt, "image/"), strings.HasPrefix(mt, "audio/"):
		return "reviewing" // 走 mediaCheckAsync
	default:
		// 文档类（pdf/word/excel/ppt 等）走主理人人工复审通道
		return "manual_review_pending"
	}
}

// triggerMediaCheck 调用微信 mediaCheckAsync 异步机审
func (c *Container) triggerMediaCheck(f *model.File, appID string) {
	// 占位：实际接入需要先把文件上传到一个可公网访问的 URL（COS）才能传给微信
	// MVP 阶段本地存储无 URL 直接可用，此处仅写日志；V2.0 接 COS 后启用
	_ = appID
	// 留 traceID 占位以便后续 callback 关联
}

// ReplaceFile 替换附件（不换链接，铁律：链接不变）
func (c *Container) ReplaceFile(fileID, authorUserID uint64, newPath, newOrigName, mimeType string, size int64) error {
	f, err := c.FileRepo.GetByID(fileID)
	if err != nil {
		return err
	}
	if f.AuthorUserID != authorUserID {
		return fmt.Errorf("无权操作此文件")
	}
	ext := strings.TrimPrefix(filepath.Ext(newOrigName), ".")
	return c.FileRepo.UpdateFields(fileID, map[string]any{
		"original_name":      newOrigName,
		"ext":                ext,
		"size":               size,
		"mime_type":          mimeType,
		"storage_path":       newPath,
		"media_check_status": "pending",
	})
}

// GenerateShareCards 生成四种分享卡片 HTML（核心复制功能）
func (c *Container) GenerateShareCards(fileID uint64) ([]model.FileShareCard, error) {
	f, err := c.FileRepo.GetByID(fileID)
	if err != nil {
		return nil, err
	}
	u, err := c.UserRepo.GetByID(f.AuthorUserID)
	if err != nil {
		return nil, err
	}
	appID := u.BoundAppID
	if appID == "" {
		return nil, fmt.Errorf("作者尚未绑定小程序 appid")
	}

	displayName := f.Name
	if f.Ext != "" {
		displayName = f.Name + "." + f.Ext
	}
	mpPath := fmt.Sprintf("pages/file-detail/file-detail?id=%d", f.ID)

	cards := []model.FileShareCard{
		{
			FileID:   fileID,
			CardType: "weapp_text_link",
			AppID:    appID,
			HTML: fmt.Sprintf(
				`<a class="weapp_text_link js_weapp_entry" data-miniprogram-type="text" data-miniprogram-appid="%s" data-miniprogram-path="%s" data-miniprogram-nickname="附件通" data-miniprogram-servicetype="">%s</a>`,
				html.EscapeString(appID),
				html.EscapeString(mpPath),
				html.EscapeString(displayName),
			),
		},
		{
			FileID:   fileID,
			CardType: "mp_miniprogram",
			AppID:    appID,
			HTML: fmt.Sprintf(
				`<mp-miniprogram data-miniprogram-appid="%s" data-miniprogram-path="%s" data-miniprogram-title="%s" data-miniprogram-imageurl="" data-miniprogram-type="card"></mp-miniprogram>`,
				html.EscapeString(appID),
				html.EscapeString(mpPath),
				html.EscapeString(displayName),
			),
		},
		{
			FileID:   fileID,
			CardType: "h5_backup",
			AppID:    appID,
			HTML:     fmt.Sprintf("https://fujian.5g6g.top/f/%s", f.DownloadCode),
		},
		{
			FileID:   fileID,
			CardType: "miniprogram_path",
			AppID:    appID,
			HTML:     mpPath,
		},
	}

	for i := range cards {
		if err := c.FileRepo.UpsertShareCard(&cards[i]); err != nil {
			return nil, err
		}
	}
	return cards, nil
}

// GetFileByCode C 端通过下载码获取文件
func (c *Container) GetFileByCode(code string) (*model.File, error) {
	f, err := c.FileRepo.GetByDownloadCode(code)
	if err != nil {
		return nil, err
	}
	_ = c.FileRepo.IncrViewCount(f.ID)
	return f, nil
}

// UpdateMediaCheckStatus 异步 mediaCheck 回调更新
func (c *Container) UpdateMediaCheckStatus(traceID, status string) error {
	var f model.File
	if err := c.DB.Where("media_check_trace_id = ?", traceID).First(&f).Error; err != nil {
		return err
	}
	fields := map[string]any{"media_check_status": status}
	if status == "reject" {
		fields["status"] = "banned"
	}
	return c.FileRepo.UpdateFields(f.ID, fields)
}
