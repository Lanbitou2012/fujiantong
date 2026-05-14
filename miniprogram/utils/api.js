// 后端 API 调用封装 — 统一走 ext.apiBaseUrl

const app = getApp()

function request(opts) {
  const baseUrl = (app && app.globalData && app.globalData.apiBaseUrl) || ''
  return new Promise((resolve, reject) => {
    wx.request({
      url: baseUrl + opts.url,
      method: opts.method || 'GET',
      data: opts.data || {},
      header: { 'Content-Type': 'application/json' },
      success: (res) => {
        if (res.statusCode === 200 && res.data && res.data.code === 0) {
          resolve(res.data.data)
        } else {
          reject(res.data || { msg: '请求失败' })
        }
      },
      fail: reject
    })
  })
}

// 通过 download_code 获取文件元信息（含作者昵称）
function getFileByCode(code) {
  return request({ url: `/api/v1/c/file/${code}` })
}

// 获取文件下载直链
function getDownloadUrl(code) {
  const baseUrl = (app && app.globalData && app.globalData.apiBaseUrl) || ''
  return `${baseUrl}/api/v1/c/download/${code}`
}

// UV 上报（用于推广员申请门槛检测）
function reportUV(payload) {
  return request({ url: '/api/v1/c/uv', method: 'POST', data: payload })
}

module.exports = {
  request,
  getFileByCode,
  getDownloadUrl,
  reportUV
}
