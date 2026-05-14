import axios from 'axios'
import { ElMessage } from 'element-plus'

const http = axios.create({
  timeout: 30000,
})

// 请求拦截器：自动附加 JWT token
http.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 公开路径白名单：在这些页面发生 401 不强制跳登录（例如入驻页调授权链接接口失败）
const PUBLIC_PATH_PREFIXES = ['/onboarding', '/login', '/auth-success', '/f/']

function isOnPublicPath() {
  const p = window.location.pathname || ''
  return PUBLIC_PATH_PREFIXES.some((pre) => p === pre || p.startsWith(pre + '/') || p.startsWith(pre))
}

// 响应拦截器：统一处理错误
http.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      // 仅在受保护路由触发时跳登录；公开页（入驻、登录、授权回调、C 端下载）不跳，让页面自行提示
      if (!isOnPublicPath()) {
        window.location.href = '/login'
        return
      }
    }
    const msg = error.response?.data?.msg || error.message || '请求失败'
    ElMessage.error(msg)
    return Promise.reject(error)
  }
)

export default http
