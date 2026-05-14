// 附件通模板小程序 — App 入口
// 关键：通过 wx.getExtConfigSync() 读取 ext_json 注入的差异化配置
// V2.8 §15.2.8：模板代码 1 套，配置 N 套

const ad = require('./utils/ad.js')

App({
  globalData: {
    extConfig: null,
    apiBaseUrl: '',
    primaryColor: '#3B82F6',
    authorName: '附件通',
    platformUserId: '',
    adUnitIds: {},
    privacyVersion: 'v2.1',
    customerServiceWxId: ''
  },

  onLaunch() {
    // 1. 读取 ext_json 注入配置
    const ext = wx.getExtConfigSync ? wx.getExtConfigSync() : {}
    const cfg = ext.ext || {}

    this.globalData.extConfig = cfg
    this.globalData.apiBaseUrl = cfg.apiBaseUrl || 'https://fujian.5g6g.top'
    this.globalData.primaryColor = cfg.primaryColor || '#3B82F6'
    this.globalData.authorName = cfg.authorName || '附件通'
    this.globalData.platformUserId = cfg.platformUserId || ''
    this.globalData.adUnitIds = cfg.adUnitIds || {}
    this.globalData.privacyVersion = cfg.privacyVersion || 'v2.1'
    this.globalData.customerServiceWxId = cfg.customerServiceWxId || ''

    // 2. 预初始化广告（开屏广告需要在 launch 时即触发）
    ad.initSplashAd(this.globalData.adUnitIds.splash)

    // 3. 静默登录（用 code 换 openid 给后端 UV 上报使用）
    wx.login({
      success: (res) => {
        if (res.code) {
          this.globalData.wxLoginCode = res.code
        }
      }
    })

    console.log('[App] 已加载作者配置:', this.globalData.authorName)
  },

  onShow() {
    // 每次回到前台触发
  },

  onError(err) {
    console.error('[App] error:', err)
  }
})
