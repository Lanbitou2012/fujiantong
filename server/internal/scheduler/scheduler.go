package scheduler

import (
	"log"

	"fujiantong/internal/config"
	"fujiantong/internal/service"

	"github.com/robfig/cron/v3"
)

// Scheduler 全局定时任务调度器
// 按 V2.8 §15.2 表执行 6 个核心 cron job
type Scheduler struct {
	cron *cron.Cron
	svc  *service.Container
}

func New(svc *service.Container) *Scheduler {
	// 使用本地时区 + 秒级精度
	c := cron.New(cron.WithLocation(loadTZ()), cron.WithSeconds())
	return &Scheduler{cron: c, svc: svc}
}

// Start 注册并启动所有 cron job
func (s *Scheduler) Start() {
	wxCfg := config.GetComponentConfig()
	if wxCfg.AppID == "" {
		log.Println("[Scheduler] WX_COMPONENT_APPID 未配置，跳过微信相关定时任务")
		return
	}

	// 1. component_access_token 自动刷新（每 30 分钟，提前过期）
	s.mustAdd("0 */30 * * * *", func() {
		if _, err := s.svc.EnsureComponentToken(wxCfg.AppID, wxCfg.AppSecret); err != nil {
			log.Printf("[Cron][ComponentToken] 刷新失败: %v", err)
		}
	}, "ComponentToken")

	// 2. 所有 authorizer_access_token 主动刷新（每小时）
	s.mustAdd("0 0 * * * *", func() {
		if err := s.svc.RefreshAllAuthorizerTokens(wxCfg.AppID, wxCfg.AppSecret); err != nil {
			log.Printf("[Cron][AuthorizerToken] 失败: %v", err)
		}
	}, "AuthorizerTokens")

	// 3. getAuthorizerList 同步（每日 02:00）
	s.mustAdd("0 0 2 * * *", func() {
		if err := s.svc.SyncAuthorizers(wxCfg.AppID, wxCfg.AppSecret); err != nil {
			log.Printf("[Cron][SyncAuthorizers] 失败: %v", err)
		}
	}, "SyncAuthorizers")

	// 4. 每日广告数据拉取（每日 03:00）
	s.mustAdd("0 0 3 * * *", func() {
		if err := s.svc.PullDailyAdStats(wxCfg.AppID, wxCfg.AppSecret); err != nil {
			log.Printf("[Cron][PullDailyAdStats] 失败: %v", err)
		}
	}, "PullDailyAdStats")

	// 5. 结算数据拉取 + 自动生成佣金结算单（每月 1 号 / 16 号 04:00）
	s.mustAdd("0 0 4 1,16 * *", func() {
		if err := s.svc.PullSettlements(wxCfg.AppID, wxCfg.AppSecret); err != nil {
			log.Printf("[Cron][PullSettlements] 失败: %v", err)
		}
	}, "PullSettlements")

	s.cron.Start()
	log.Printf("[Scheduler] 已启动，注册 %d 个定时任务", len(s.cron.Entries()))
}

func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
}

func (s *Scheduler) mustAdd(spec string, cmd func(), name string) {
	if _, err := s.cron.AddFunc(spec, cmd); err != nil {
		log.Fatalf("[Scheduler] 注册任务 %s 失败: %v", name, err)
	}
	log.Printf("[Scheduler] 已注册: %s @ %s", name, spec)
}
