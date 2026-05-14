const app = getApp()
const api = require('../../utils/api.js')
const ad = require('../../utils/ad.js')

Page({
  data: {
    loading: true,
    error: '',
    file: null,
    unlocked: false,
    adLoading: false,
    primaryColor: '#3B82F6',
    customerServiceWxId: ''
  },

  onLoad(options) {
    this.code = options.code || options.id || ''
    this.setData({
      primaryColor: app.globalData.primaryColor,
      customerServiceWxId: app.globalData.customerServiceWxId
    })

    if (!this.code) {
      this.setData({ loading: false, error: '无效的文件链接' })
      return
    }

    this.loadFile()

    // UV 上报（用于推广员申请门槛检测，V2.8 §15.2.4）
    api.reportUV({ code: this.code }).catch(() => {})
  },

  async loadFile() {
    try {
      const data = await api.getFileByCode(this.code)
      data.size_text = formatSize(data.size)
      this.setData({ file: data, loading: false })
    } catch (e) {
      this.setData({
        loading: false,
        error: (e && e.msg) || '文件不存在或已下架'
      })
    }
  },

  async onWatchAd() {
    const adUnitId = app.globalData.adUnitIds.rewardedVideo
    if (!adUnitId) {
      // 广告位未配置时，给老用户/调试模式一个降级口
      wx.showModal({
        title: '广告未就绪',
        content: '当前无可用广告，是否直接下载？',
        success: (res) => {
          if (res.confirm) this.setData({ unlocked: true })
        }
      })
      return
    }

    this.setData({ adLoading: true })
    const result = await ad.showRewardedVideo(adUnitId)
    this.setData({ adLoading: false })

    if (result.result === 'completed') {
      wx.showToast({ title: '解锁成功 🎉', icon: 'success' })
      this.setData({ unlocked: true })
    } else if (result.result === 'closed') {
      wx.showToast({ title: '需完整观看广告才能解锁', icon: 'none' })
    } else {
      wx.showToast({ title: '广告加载失败，请重试', icon: 'none' })
    }
  },

  onDownload() {
    const url = api.getDownloadUrl(this.code)
    wx.showLoading({ title: '下载中...' })
    wx.downloadFile({
      url,
      success: (res) => {
        wx.hideLoading()
        if (res.statusCode === 200) {
          wx.openDocument({
            filePath: res.tempFilePath,
            showMenu: true,
            success: () => {},
            fail: () => {
              wx.showModal({
                title: '无法预览',
                content: '该类型文件无法在小程序内预览，已保存到本地，可在「微信 → 文件」查看',
                showCancel: false
              })
            }
          })
        } else {
          wx.showToast({ title: '下载失败', icon: 'none' })
        }
      },
      fail: () => {
        wx.hideLoading()
        wx.showToast({ title: '下载失败', icon: 'none' })
      }
    })
  },

  copyCustomer() {
    wx.setClipboardData({ data: this.data.customerServiceWxId })
  },

  goHome() {
    wx.switchTab({ url: '/pages/home/home' })
  },

  onShareAppMessage() {
    return {
      title: this.data.file ? this.data.file.name : '附件通',
      path: `/pages/file-detail/file-detail?code=${this.code}`
    }
  }
})

function formatSize(bytes) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0, size = bytes
  while (size >= 1024 && i < units.length - 1) { size /= 1024; i++ }
  return size.toFixed(i > 0 ? 1 : 0) + ' ' + units[i]
}
