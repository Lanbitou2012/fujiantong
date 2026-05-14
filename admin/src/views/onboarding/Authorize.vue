<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="header">
        <h2>授权附件通代运营你的小程序</h2>
        <p class="sub">下一步：授权后平台将<strong>全自动</strong>为你完成配置 → 部署 → 提审 → 上线</p>
      </div>

      <el-alert title="授权后平台将自动完成以下动作（无需你操作）" type="success" :closable="false">
        <ul class="auto-list">
          <li>✅ 设置默认分账比例 28%（平台），你拿 72%</li>
          <li>✅ 创建开屏 / 激励视频 / 插屏三种广告位</li>
          <li>✅ 配置「工具 → 办公」类目</li>
          <li>✅ 配置小程序服务器域名 + 隐私指引</li>
          <li>✅ 部署模板代码（注入你的品牌名）</li>
          <li>✅ 自动提交审核（最快 24 小时通过）</li>
          <li>✅ 审核通过后自动发布上线</li>
        </ul>
      </el-alert>

      <el-divider />

      <div class="perm-block">
        <h3>⚠️ 重要：授权时务必勾选这两组权限集</h3>
        <ul>
          <li>✅ <strong>流量主代运营</strong>（权限集 135）</li>
          <li>✅ <strong>小程序代码管理</strong>（权限集 17）</li>
        </ul>
        <p class="warn">两组权限缺一不可，否则平台无法完成代部署！</p>
      </div>

      <div class="btn-block">
        <el-button v-if="!authURL" type="primary" size="large" @click="genURL" :loading="loading" class="cta">
          生成授权链接
        </el-button>

        <template v-else>
          <p class="tip">复制下方链接到 <strong>电脑浏览器</strong> 打开 → 用微信扫码完成授权</p>
          <el-input v-model="authURL" readonly type="textarea" :rows="3" />
          <div class="btn-row">
            <el-button type="primary" size="large" @click="copyURL" class="cta">复制授权链接</el-button>
            <el-button size="large" @click="goOpen" class="cta" plain>直接打开</el-button>
          </div>
          <p class="hint">授权完成后自动跳转回附件通后台</p>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import http from '@/utils/http'

const route = useRoute()
const loading = ref(false)
const authURL = ref('')

async function genURL() {
  loading.value = true
  try {
    const { data } = await http.get('/api/v1/auth/wx-component/url', {
      params: { biz_appid: route.query.appid || '' },
    })
    if (data.code === 0) {
      authURL.value = data.data.auth_url
      ElMessage.success('已生成授权链接')
    } else {
      ElMessage.error(data.msg)
    }
  } catch (e) {
    ElMessage.error('生成失败：' + (e.message || ''))
  } finally {
    loading.value = false
  }
}

function copyURL() {
  navigator.clipboard?.writeText(authURL.value)
  ElMessage.success('授权链接已复制')
}

function goOpen() {
  window.open(authURL.value, '_blank')
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh; background: #f5f7fa;
  display: flex; align-items: center; justify-content: center;
  padding: 30px 20px;
}
.auth-card {
  width: 100%; max-width: 600px;
  background: #fff; border-radius: 12px; padding: 36px 28px;
  box-shadow: 0 8px 30px rgba(0,0,0,0.06);
}
.header { text-align: center; margin-bottom: 20px; }
.header h2 { margin: 0; color: #333; }
.header .sub { margin: 8px 0 0; color: #888; font-size: 13px; line-height: 1.6; }
.auto-list { margin: 6px 0 0; padding-left: 18px; line-height: 1.9; }
.auto-list li { color: #67c23a; font-size: 13px; }
.perm-block { background: #fef0f0; border-radius: 8px; padding: 16px; margin-bottom: 20px; }
.perm-block h3 { margin: 0 0 8px; font-size: 15px; color: #f56c6c; }
.perm-block ul { margin: 0; padding-left: 22px; color: #555; line-height: 1.8; }
.perm-block .warn { margin: 8px 0 0; color: #f56c6c; font-size: 13px; font-weight: 600; }
.btn-block { text-align: center; }
.cta { width: 100%; height: 50px; font-size: 16px; border-radius: 8px; }
.btn-row { display: flex; gap: 12px; margin-top: 12px; }
.btn-row .cta { flex: 1; }
.tip { color: #555; margin-bottom: 12px; font-size: 14px; }
.hint { color: #999; font-size: 12px; margin: 16px 0 0; }
</style>
