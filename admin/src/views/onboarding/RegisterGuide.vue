<template>
  <div class="guide-page">
    <div class="guide-container">
      <div class="page-header">
        <h2>注册你的个人主体小程序</h2>
        <p class="sub">按下面 5 步在微信公众平台完成，约 10 分钟</p>
      </div>

      <el-steps :active="active" finish-status="success" align-center class="steps-bar">
        <el-step title="进入注册页" />
        <el-step title="填写邮箱" />
        <el-step title="选择小程序" />
        <el-step title="实名认证" />
        <el-step title="完成 + 回填 AppID" />
      </el-steps>

      <div class="step-content">
        <el-card v-show="active === 0" shadow="never">
          <h3>第 1 步：进入小程序注册页</h3>
          <p>访问：<el-link type="primary" href="https://mp.weixin.qq.com/wxopen/waregister" target="_blank">https://mp.weixin.qq.com/wxopen/waregister</el-link></p>
          <p class="tip">⚠️ 请使用电脑浏览器访问，**不要用微信内置浏览器**</p>
          <p>点击右上角"立即注册" → 选择"小程序"。</p>
        </el-card>

        <el-card v-show="active === 1" shadow="never">
          <h3>第 2 步：填写一个未注册过的邮箱</h3>
          <p>该邮箱将作为小程序的登录账号。<strong>必须从未在微信公众平台注册过任何账号</strong>（小程序/公众号/企业号都算）。</p>
          <p class="tip">💡 建议新建一个专用邮箱：QQ邮箱 / 163 / Gmail 都可以</p>
          <p>填写密码（建议 12 位以上、包含大小写+数字） → 收激活邮件 → 点邮件中的链接激活。</p>
        </el-card>

        <el-card v-show="active === 2" shadow="never">
          <h3>第 3 步：选择"个人"主体</h3>
          <p>激活后跳转到信息填写页：</p>
          <ul>
            <li>主体类型：选 <strong>"个人"</strong>（重要！选企业要营业执照）</li>
            <li>填写真实姓名、身份证号</li>
            <li>填写手机号（用于人脸识别）</li>
          </ul>
          <p class="tip">⚠️ 个人主体小程序<strong>不能用同一身份证再注册第二个</strong>，请确认是否已绑定过其他小程序</p>
        </el-card>

        <el-card v-show="active === 3" shadow="never">
          <h3>第 4 步：扫码完成人脸识别</h3>
          <p>用 <strong>本人微信</strong> 扫码 → 微信里跳转"实名认证小程序" → 摄像头人脸识别。</p>
          <p>识别完成后注册成功，此时获得一个 AppID（在小程序后台"开发管理 → 开发设置"中查看）。</p>
        </el-card>

        <el-card v-show="active === 4" shadow="never">
          <h3>第 5 步：回填 AppID 到附件通</h3>
          <p>复制您新注册小程序的 AppID（格式：wxXXXXXXXXXXXXXXXX），点击下方按钮回填：</p>
          <el-form :model="form" label-position="top" class="appid-form">
            <el-form-item label="小程序 AppID">
              <el-input v-model="form.appid" placeholder="wxXXXXXXXXXXXXXXXX" size="large" />
            </el-form-item>
            <el-button type="primary" size="large" @click="submitAppID" :loading="submitting" style="width:100%">
              提交并进入授权步骤 →
            </el-button>
          </el-form>
        </el-card>
      </div>

      <div class="nav-bar">
        <el-button v-if="active > 0" @click="active--" plain>上一步</el-button>
        <el-button v-if="active < 4" type="primary" @click="active++">下一步</el-button>
      </div>

      <el-card class="help-card" shadow="never">
        <template #header>
          <span>❓ 卡在哪一步？</span>
        </template>
        <p>联系您的推广员或附件通客服，远程屏幕共享指导：</p>
        <p>📧 <strong>support@fujian.5g6g.top</strong></p>
        <p>💬 微信：<strong>fjt_helper</strong></p>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import http from '@/utils/http'

const router = useRouter()
const active = ref(0)
const submitting = ref(false)
const form = ref({ appid: '' })

async function submitAppID() {
  const appid = form.value.appid.trim()
  if (!/^wx[a-zA-Z0-9]{16}$/.test(appid)) {
    ElMessage.warning('请输入合法的 AppID（wx 开头共 18 位）')
    return
  }
  submitting.value = true
  try {
    const { data } = await http.post('/api/v1/author/appid', { appid })
    if (data.code === 0) {
      ElMessage.success('AppID 回填成功')
      router.push({ path: '/onboarding/authorize', query: { appid } })
    } else {
      ElMessage.error(data.msg)
    }
  } catch (e) {
    ElMessage.error(e.message || '提交失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.guide-page {
  min-height: 100vh;
  background: #f5f7fa;
  padding: 30px 20px;
}
.guide-container {
  max-width: 720px; margin: 0 auto;
}
.page-header { text-align: center; margin-bottom: 24px; }
.page-header h2 { margin: 0; color: #333; }
.page-header .sub { color: #888; margin-top: 4px; font-size: 13px; }
.steps-bar { margin-bottom: 32px; }
.step-content { min-height: 280px; }
.step-content h3 { margin: 0 0 12px; color: #333; font-size: 18px; }
.step-content p { line-height: 1.7; color: #555; margin: 8px 0; }
.step-content ul { padding-left: 22px; color: #555; line-height: 1.9; }
.tip {
  background: #fdf6ec; border-left: 3px solid #e6a23c;
  padding: 8px 12px; color: #875a13 !important;
  border-radius: 4px;
}
.appid-form { margin-top: 16px; }
.nav-bar {
  display: flex; justify-content: space-between; gap: 12px;
  margin-top: 24px;
}
.nav-bar :deep(.el-button) { flex: 1; }
.help-card { margin-top: 32px; }
.help-card p { margin: 4px 0; color: #666; font-size: 13px; }
</style>
