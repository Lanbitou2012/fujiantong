<template>
  <div class="auth-success">
    <div class="success-card">
      <template v-if="!error">
        <div class="icon-wrap">
          <el-icon :size="56" color="#10B981"><CircleCheckFilled /></el-icon>
        </div>
        <h1>授权成功</h1>
        <p class="sub">您的小程序已成功授权给附件通平台</p>
        <ul class="info-list">
          <li v-if="appid"><span>小程序 AppID</span><strong>{{ appid }}</strong></li>
          <li v-if="userID"><span>作者 ID</span><strong>{{ userID }}</strong></li>
          <li><span>当前阶段</span><strong style="color:#6366F1">系统正在自动部署小程序代码</strong></li>
        </ul>
        <p class="hint">平台将在后台完成「广告位配置 → 类目设置 → 代码上传 → 提交审核」全流程，预计 1-2 工作日内审核通过并自动上线，期间您可以进入后台查看进度。</p>
        <el-button type="primary" size="large" style="width: 100%" @click="goDashboard">
          进入作者后台
        </el-button>
      </template>

      <template v-else>
        <div class="icon-wrap">
          <el-icon :size="56" color="#EF4444"><CircleCloseFilled /></el-icon>
        </div>
        <h1>授权失败</h1>
        <p class="sub error-msg">{{ errorMsg }}</p>
        <el-button size="large" style="width: 100%" @click="$router.push('/onboarding')">
          返回重试
        </el-button>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const appid = ref('')
const userID = ref('')
const error = ref('')

const errorMap = {
  missing_auth_code: '未收到授权码，授权流程未完成',
  token_failed: '签发登录态失败，请重试',
}
const errorMsg = ref('')

onMounted(() => {
  appid.value = route.query.appid || ''
  userID.value = route.query.user_id || ''
  error.value = route.query.error || ''
  if (error.value) {
    errorMsg.value = errorMap[error.value] || decodeURIComponent(error.value)
    return
  }
  // 从 hash 取 token 并落 localStorage
  const hash = window.location.hash || ''
  const m = hash.match(/token=([^&]+)/)
  if (m && m[1]) {
    localStorage.setItem('token', m[1])
    localStorage.setItem('user', JSON.stringify({
      id: Number(userID.value) || 0,
      bound_appid: appid.value,
      role_name: 'author',
    }))
    // 清掉 URL 上的 token
    history.replaceState(null, '', route.fullPath.split('#')[0])
  }
})

function goDashboard() {
  router.push('/dashboard')
}
</script>

<style scoped>
.auth-success {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #F0FDF4 0%, #EEF2FF 100%);
  padding: 20px;
  font-family: -apple-system, BlinkMacSystemFont, 'PingFang SC', 'Microsoft YaHei', sans-serif;
}
.success-card {
  width: 100%;
  max-width: 480px;
  background: #fff;
  border-radius: 20px;
  padding: 48px 36px 40px;
  box-shadow: 0 20px 60px rgba(16, 185, 129, 0.15);
  text-align: center;
}
.icon-wrap { margin-bottom: 16px; }
h1 { font-size: 24px; margin: 0 0 8px; color: #0F172A; }
.sub { color: #64748B; margin: 0 0 24px; font-size: 14px; }
.error-msg { color: #EF4444; }
.info-list {
  list-style: none; padding: 0; margin: 0 0 20px;
  background: #F9FAFB; border-radius: 12px; padding: 16px 20px; text-align: left;
}
.info-list li { display: flex; justify-content: space-between; padding: 6px 0; font-size: 13px; color: #475569; }
.info-list li span { color: #94A3B8; }
.info-list li strong { color: #0F172A; font-weight: 600; }
.hint {
  background: #FEF3C7; color: #92400E; font-size: 13px; line-height: 1.6;
  padding: 12px 16px; border-radius: 10px; margin: 0 0 24px; text-align: left;
}
</style>
