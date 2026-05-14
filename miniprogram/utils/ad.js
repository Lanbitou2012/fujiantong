// 广告 SDK 封装 — 开屏 / 激励视频 / 插屏
// 广告位 ID 由平台通过 ext_json 注入，模板代码不写死

let splashAd = null
let rewardedVideoAd = null
let interstitialAd = null

/**
 * 初始化开屏广告（App.onLaunch 时调用）
 * 注意：开屏广告需在小程序启动时即创建实例
 */
function initSplashAd(adUnitId) {
  if (!adUnitId || !wx.createAppBoxAd) return // 暂不强求

  // V1.5 真正接入开屏广告需要 createInterstitialAd 或自定义启动页
  // MVP 阶段仅占位
}

/**
 * 创建/复用激励视频广告
 * @returns Promise<{ result: 'completed' | 'closed' | 'error', error?: any }>
 */
function showRewardedVideo(adUnitId) {
  return new Promise((resolve) => {
    if (!adUnitId) {
      console.warn('[Ad] 激励视频 adUnitId 未配置')
      return resolve({ result: 'error', error: 'no_ad_unit' })
    }

    if (!rewardedVideoAd) {
      rewardedVideoAd = wx.createRewardedVideoAd({ adUnitId })
      rewardedVideoAd.onError((err) => {
        console.error('[Ad] 激励视频 error', err)
      })
    }

    let completed = false
    const onClose = (res) => {
      // res.isEnded === true 表示用户完整观看
      rewardedVideoAd.offClose(onClose)
      completed = res && res.isEnded
      resolve({ result: completed ? 'completed' : 'closed' })
    }
    rewardedVideoAd.onClose(onClose)

    rewardedVideoAd.show().catch(() => {
      rewardedVideoAd.load()
        .then(() => rewardedVideoAd.show())
        .catch((err) => {
          console.error('[Ad] 激励视频加载失败', err)
          resolve({ result: 'error', error: err })
        })
    })
  })
}

/**
 * 显示插屏广告（在文件预览/列表切换时弹出）
 */
function showInterstitial(adUnitId) {
  if (!adUnitId) return

  if (!interstitialAd) {
    interstitialAd = wx.createInterstitialAd({ adUnitId })
    interstitialAd.onError((err) => {
      console.warn('[Ad] 插屏 error', err)
    })
  }
  interstitialAd.show().catch(() => {
    interstitialAd.load().then(() => interstitialAd.show())
  })
}

module.exports = {
  initSplashAd,
  showRewardedVideo,
  showInterstitial
}
