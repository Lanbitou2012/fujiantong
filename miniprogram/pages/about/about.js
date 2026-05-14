const app = getApp()

Page({
  data: {
    authorName: '',
    primaryColor: '#3B82F6',
    customerServiceWxId: 'fjt_helper'
  },

  onLoad() {
    this.setData({
      authorName: app.globalData.authorName,
      primaryColor: app.globalData.primaryColor,
      customerServiceWxId: app.globalData.customerServiceWxId || 'fjt_helper'
    })
  },

  openCustomer() {
    wx.setClipboardData({
      data: this.data.customerServiceWxId,
      success: () => {
        wx.showToast({ title: '客服微信已复制', icon: 'success' })
      }
    })
  },

  openPrivacy() {
    wx.openPrivacyContract({
      fail: () => {
        wx.showToast({ title: '隐私协议加载失败', icon: 'none' })
      }
    })
  },

  openAgreement() {
    wx.navigateTo({
      url: '/pages/about/about?type=agreement',
      fail: () => {
        wx.showToast({ title: '协议加载失败', icon: 'none' })
      }
    })
  },

  openUgcDeclare() {
    wx.showModal({
      title: 'UGC 内容声明',
      content: '本小程序内容由作者自主上传，平台已对内容进行机审/复审。如发现违规内容请通过客服微信举报，平台将在 24 小时内处理。',
      showCancel: false,
      confirmText: '我知道了'
    })
  }
})
