<template>
  <div class="qr-page">
    <h2>我的专属推广二维码</h2>
    <el-alert
      title="将下方二维码分享给潜在作者"
      type="success"
      :closable="false"
      description="新作者扫码进入入驻流程，注册成功后将自动绑定为您的下级作者，您可从其流量主收益中获得 8.4% 推广佣金（V2.8 §4 资金流）"
      class="info"
    />

    <div class="qr-block">
      <div class="qr-wrap">
        <canvas ref="qrCanvas" />
        <div class="qr-info">
          <p class="label">推广员 ID</p>
          <p class="value">{{ userStore.user?.id }}</p>
          <p class="label">推广链接</p>
          <el-input v-model="inviteURL" readonly size="small">
            <template #append>
              <el-button @click="copyURL">复制</el-button>
            </template>
          </el-input>
          <el-button type="primary" @click="downloadQR" style="margin-top:12px;width:100%">
            <el-icon><Download /></el-icon>
            保存二维码图片
          </el-button>
        </div>
      </div>
    </div>

    <el-card class="tips" shadow="never">
      <template #header>📌 推广员使用须知</template>
      <ol>
        <li>把二维码贴在你自己的公众号文章、朋友圈、社群等位置</li>
        <li>新作者扫码后会经过：欢迎页 → 微信扫码登录 → 注册指引 → 授权 → 上线</li>
        <li>该作者发表的所有附件触发的广告收益，按月结算给你 8.4%</li>
        <li>佣金为单层提成（合规要求），不存在多级分销</li>
        <li>结算单状态变更可在「推广收益」页查看</li>
      </ol>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import QRCode from 'qrcode'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const qrCanvas = ref(null)

const inviteURL = computed(() => {
  const origin = window.location.origin
  return `${origin}/onboarding?p=${userStore.user?.id || ''}`
})

async function renderQR() {
  if (!qrCanvas.value) return
  try {
    await QRCode.toCanvas(qrCanvas.value, inviteURL.value, {
      width: 220, margin: 1,
      color: { dark: '#222', light: '#fff' },
    })
  } catch (e) {
    ElMessage.error('生成二维码失败')
  }
}

function copyURL() {
  navigator.clipboard.writeText(inviteURL.value)
  ElMessage.success('链接已复制')
}

function downloadQR() {
  const url = qrCanvas.value.toDataURL('image/png')
  const a = document.createElement('a')
  a.href = url
  a.download = `附件通推广二维码-${userStore.user?.id}.png`
  a.click()
}

onMounted(renderQR)
</script>

<style scoped>
.qr-page { padding: 0; }
h2 { margin: 0 0 16px 0; }
.info { margin-bottom: 20px; }
.qr-block {
  background: #fff; border-radius: 12px;
  padding: 32px; margin-bottom: 20px;
  display: flex; justify-content: center;
}
.qr-wrap {
  display: flex; gap: 32px; align-items: center;
}
.qr-info { min-width: 280px; }
.qr-info .label {
  color: #999; font-size: 12px; margin: 4px 0; letter-spacing: 1px;
}
.qr-info .value {
  color: #333; font-size: 18px; font-weight: 600; margin: 0 0 12px 0;
}
.tips ol {
  margin: 0; padding-left: 20px;
  color: #555; font-size: 13px; line-height: 1.9;
}
@media (max-width: 600px) {
  .qr-wrap { flex-direction: column; gap: 20px; }
  .qr-info { min-width: auto; width: 100%; }
}
</style>
