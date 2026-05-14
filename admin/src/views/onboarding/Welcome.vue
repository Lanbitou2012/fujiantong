<template>
  <div class="onboarding-page">
    <div class="card">
      <div class="hero">
        <h1>附件通</h1>
        <p class="slogan">让公众号附件，变成你的小程序流量</p>
      </div>

      <div v-if="promoter" class="invite-banner">
        <el-icon><User /></el-icon>
        <span>由 <strong>{{ promoter.nickname || '推广员' }}</strong> 推荐入驻</span>
      </div>

      <div class="benefit">
        <div class="benefit-item">
          <div class="num">72%</div>
          <div class="lbl">流量主收益归你</div>
        </div>
        <div class="benefit-item">
          <div class="num">3 步</div>
          <div class="lbl">入驻即开通独立小程序</div>
        </div>
        <div class="benefit-item">
          <div class="num">0 元</div>
          <div class="lbl">永久免费使用</div>
        </div>
      </div>

      <ol class="steps">
        <li><strong>注册个人主体小程序</strong>（10 分钟，附详细教程）</li>
        <li><strong>回填 AppID + 授权代运营</strong>（自动配置广告位、提交审核）</li>
        <li><strong>上传附件 + 一键复制到公众号</strong>（30 秒发文）</li>
      </ol>

      <el-button type="primary" size="large" class="cta" @click="handleStart">
        立即开始入驻
        <el-icon><ArrowRight /></el-icon>
      </el-button>

      <p class="login-tip">已是作者？<el-link type="primary" @click="$router.push('/login')">前往登录</el-link></p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'

const route = useRoute()
const router = useRouter()
const promoter = ref(null)
const promoterID = route.query.p || ''

onMounted(async () => {
  if (promoterID) {
    try {
      // 简化：仅展示推广员邀请信息，不调用后端（推广员资料从扫码登录后绑定）
      promoter.value = { nickname: '附件通推荐人' }
    } catch (e) { /* ignore */ }
  }
})

async function handleStart() {
  // 跳转到微信扫码登录，带 promoter_id
  try {
    const { data } = await axios.get('/api/v1/auth/scan/url', { params: { state: promoterID ? `promoter_${promoterID}` : 'onboarding' } })
    if (data.code === 0) {
      window.location.href = data.data.url
    }
  } catch (e) {
    // 微信开放平台未配置时降级为账号登录
    router.push('/login')
  }
}
</script>

<style scoped>
.onboarding-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex; align-items: center; justify-content: center;
  padding: 20px;
}
.card {
  width: 100%; max-width: 480px;
  background: #fff; border-radius: 16px; padding: 40px 32px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.2);
}
.hero { text-align: center; margin-bottom: 24px; }
.hero h1 { margin: 0; font-size: 32px; color: #333; letter-spacing: 2px; }
.slogan { margin: 8px 0 0; color: #666; font-size: 14px; }
.invite-banner {
  background: #ecf5ff; border: 1px solid #b3d8ff; border-radius: 8px;
  padding: 10px 14px; margin-bottom: 20px;
  display: flex; align-items: center; gap: 8px;
  color: #409eff; font-size: 13px;
}
.benefit {
  display: flex; justify-content: space-around;
  background: #fafbfc; border-radius: 12px;
  padding: 20px 12px; margin-bottom: 24px;
}
.benefit-item { text-align: center; }
.benefit-item .num {
  font-size: 26px; font-weight: 700;
  background: linear-gradient(90deg, #667eea, #764ba2);
  -webkit-background-clip: text; -webkit-text-fill-color: transparent;
}
.benefit-item .lbl { font-size: 12px; color: #888; margin-top: 4px; }
.steps {
  padding-left: 22px; margin: 0 0 28px;
  color: #555; font-size: 14px; line-height: 1.9;
}
.steps li { margin-bottom: 4px; }
.cta {
  width: 100%; height: 50px; font-size: 16px; border-radius: 10px;
}
.login-tip {
  text-align: center; margin: 16px 0 0;
  color: #999; font-size: 13px;
}
</style>
