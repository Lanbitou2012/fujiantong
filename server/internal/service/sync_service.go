package service

import (
	"fmt"
	"log"
	"time"

	"fujiantong/internal/model"
	"fujiantong/internal/wechat"
)

// SyncAuthorizers 调用 getAuthorizerList 同步全平台已授权小程序
// 频率：每日 02:00（V2.8 §15.2 表）
func (c *Container) SyncAuthorizers(componentAppID, componentSecret string) error {
	componentToken, err := c.EnsureComponentToken(componentAppID, componentSecret)
	if err != nil {
		return fmt.Errorf("获取 component_token 失败: %w", err)
	}

	const pageSize = 500
	offset := 0
	totalSynced := 0

	for {
		resp, err := c.WxComponent.GetAuthorizerList(componentToken, componentAppID, offset, pageSize)
		if err != nil {
			return fmt.Errorf("getAuthorizerList 失败: %w", err)
		}

		for _, item := range resp.List {
			// 检查本地是否已有该 appid 的授权记录
			auth, err := c.AuthRepo.GetByAppID(item.AuthorizerAppID)
			if err != nil || auth == nil {
				// 新增授权（可能是手动绑定但事件丢失的情况）
				log.Printf("[SyncAuthorizers] 发现新授权: %s，触发补登记", item.AuthorizerAppID)
				continue
			}
			if auth.AuthorizerRefreshToken != item.RefreshToken {
				auth.AuthorizerRefreshToken = item.RefreshToken
				_ = c.AuthRepo.Update(auth)
			}
			totalSynced++
		}

		if len(resp.List) < pageSize {
			break
		}
		offset += pageSize
	}

	log.Printf("[SyncAuthorizers] 同步完成，共 %d 条", totalSynced)
	return nil
}

// PullDailyAdStats 拉取前一天每个授权小程序的广告数据
// 频率：每日 03:00
func (c *Container) PullDailyAdStats(componentAppID, componentSecret string) error {
	componentToken, err := c.EnsureComponentToken(componentAppID, componentSecret)
	if err != nil {
		return err
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	auths, err := c.AuthRepo.ListAuthorized()
	if err != nil {
		return err
	}

	var batch []model.DailyAdStat
	for _, auth := range auths {
		req := wechat.GetAdposDetailReq{
			Page: 1, PageSize: 100,
			StartDate: yesterday, EndDate: yesterday,
		}
		resp, err := c.WxComponent.GetAdposDetail(componentToken, auth.AppID, req)
		if err != nil {
			log.Printf("[PullDailyAdStats] %s 拉取失败: %v", auth.AppID, err)
			continue
		}
		for _, item := range resp.List {
			batch = append(batch, model.DailyAdStat{
				AppID:         auth.AppID,
				Date:          yesterday,
				SlotID:        item.SlotID,
				AdSlot:        item.AdSlot,
				ExposureCount: item.ExposureCount,
				ClickCount:    item.ClickCount,
				Income:        item.Income,
				ECPM:          item.ECPM,
			})
		}
	}

	if len(batch) > 0 {
		if err := c.FinanceRepo.BatchCreateAdStats(batch); err != nil {
			return err
		}
		log.Printf("[PullDailyAdStats] 写入 %d 条", len(batch))
	}
	return nil
}

// PullSettlements 拉取上一结算期的收入数据
// 频率：每月 1 号 / 16 号 04:00
func (c *Container) PullSettlements(componentAppID, componentSecret string) error {
	componentToken, err := c.EnsureComponentToken(componentAppID, componentSecret)
	if err != nil {
		return err
	}

	month := time.Now().AddDate(0, 0, -1).Format("2006-01")
	auths, err := c.AuthRepo.ListAuthorized()
	if err != nil {
		return err
	}

	var batch []model.SettlementRecord
	for _, auth := range auths {
		req := wechat.GetSettlementReq{Page: 1, PageSize: 50, Month: month}
		resp, err := c.WxComponent.GetSettlement(componentToken, auth.AppID, req)
		if err != nil {
			log.Printf("[PullSettlements] %s 失败: %v", auth.AppID, err)
			continue
		}
		for _, item := range resp.List {
			batch = append(batch, model.SettlementRecord{
				AppID:         auth.AppID,
				UserID:        auth.UserID,
				SettlePeriod:  item.SettlementDate,
				TotalRevenue:  item.SettAmount,
				PlatformShare: int64(float64(item.SettAmount) * 0.28),
				AuthorShare:   int64(float64(item.SettAmount) * 0.72),
			})
		}
	}

	if len(batch) > 0 {
		if err := c.FinanceRepo.BatchCreateSettlements(batch); err != nil {
			return err
		}
		log.Printf("[PullSettlements] 写入 %d 条", len(batch))
	}

	// 拉完结算后，自动生成所有推广员的佣金结算单
	c.autoGenerateAllCommissions(month)
	return nil
}

// autoGenerateAllCommissions 为所有推广员自动生成本期佣金结算单
func (c *Container) autoGenerateAllCommissions(period string) {
	var promoters []model.User
	c.DB.Where("status = 1 AND is_promoter = true").Find(&promoters)
	for _, p := range promoters {
		if _, err := c.GenerateCommissionSettlement(p.ID, period); err != nil {
			log.Printf("[CommissionGen] 推广员 %d 期 %s 生成失败: %v", p.ID, period, err)
		}
	}
}

// RefreshAllAuthorizerTokens 主动刷新所有授权 token（防止事件推送丢失导致 token 过期）
// 频率：每小时
func (c *Container) RefreshAllAuthorizerTokens(componentAppID, componentSecret string) error {
	auths, err := c.AuthRepo.ListAuthorized()
	if err != nil {
		return err
	}
	for _, auth := range auths {
		if _, err := c.EnsureAuthorizerToken(componentAppID, componentSecret, auth.AppID); err != nil {
			log.Printf("[TokenRefresh] %s 刷新失败: %v", auth.AppID, err)
		}
	}
	return nil
}
