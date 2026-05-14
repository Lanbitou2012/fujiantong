package scheduler

import (
	"log"

	"fujiantong/internal/config"
	"fujiantong/internal/service"

	"github.com/robfig/cron/v3"
)

// Scheduler 全局定时任务调度器
// V3.0：仅保留 component_access_token 自动刷新（核心可用性依赖）。
// 数据同步类 cron（SyncAuthorizers / PullDailyAdStats / PullSettlements）挪到 V1.5。
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

	// component_access_token 自动刷新（每 30 分钟，提前过期）
	// 这是入驻流水线最关键的依赖：没有 component_token 就拼不出 auth_url。
	s.mustAdd("0 */30 * * * *", func() {
		if _, err := s.svc.EnsureComponentToken(wxCfg.AppID, wxCfg.AppSecret); err != nil {
			log.Printf("[Cron][ComponentToken] 刷新失败: %v", err)
		}
	}, "ComponentToken")

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
