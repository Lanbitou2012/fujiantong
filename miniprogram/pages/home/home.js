const app = getApp()

Page({
  data: {
    authorName: '',
    primaryColor: '#3B82F6',
    recentFiles: []
  },

  onLoad() {
    this.setData({
      authorName: app.globalData.authorName,
      primaryColor: app.globalData.primaryColor
    })

    // V1.5: 加载作者近期附件列表
    // this.loadRecentFiles()
  },

  onFileTap(e) {
    const code = e.currentTarget.dataset.code
    wx.navigateTo({ url: `/pages/file-detail/file-detail?code=${code}` })
  },

  onShareAppMessage() {
    return {
      title: `${this.data.authorName} | 附件通`,
      path: '/pages/home/home'
    }
  }
})
