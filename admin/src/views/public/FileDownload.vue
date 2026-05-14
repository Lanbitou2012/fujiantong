<template>
  <div class="download-page">
    <div class="download-card" :class="{ 'is-pc': env.isPC }">
      <div class="brand">
        <h1>附件通</h1>
        <p class="tagline">公众号附件极速分享</p>
      </div>

      <div v-if="loading" class="state">
        <el-icon class="is-loading" :size="32"><Loading /></el-icon>
        <p>加载中...</p>
      </div>

      <div v-else-if="error" class="state error">
        <el-icon :size="48"><CircleCloseFilled /></el-icon>
        <h2>{{ error }}</h2>
        <p>请联系作者或返回公众号文章</p>
      </div>

      <div v-else-if="file" class="layout">
        <!-- 主信息区 -->
        <div class="file-info">
          <div class="file-icon" :class="extClass">
            {{ file.ext.toUpperCase() }}
          </div>
          <h2 class="filename">{{ file.name }}</h2>
          <div class="meta">
            <span><el-icon><User /></el-icon> {{ file.author_name || '匿名作者' }}</span>
            <span><el-icon><Files /></el-icon> {{ formatSize(file.size) }}</span>
            <span><el-icon><View /></el-icon> {{ file.view_count }} 浏览</span>
            <span><el-icon><Download /></el-icon> {{ file.download_count }} 下载</span>
          </div>

          <!-- ─── 场景 1: 微信内 ─── -->
          <template v-if="env.isWeixin">
            <div v-if="hasLaunch" class="action-block">
              <wx-open-launch-weapp
                v-if="wxLaunchReady"
                :username="launchInfo.appid"
                :path="`${launchInfo.path}?${launchInfo.query}`"
                @launch="onLaunchSuccess"
                @error="onLaunchError"
                class="launch-tag"
              >
                <component :is="'script'" type="text/wxtag-template">
                  <button style="width:100%;height:48px;font-size:16px;background:#07c160;color:#fff;border:none;border-radius:8px;">在小程序中打开 →</button>
                </component>
              </wx-open-launch-weapp>
              <div v-else class="weixin-hint">
                <p>📱 您正在微信内浏览</p>
                <p class="sub">右上角 <strong>···</strong> → <strong>用浏览器打开</strong>，即可一键拉起小程序查看完整体验</p>
              </div>
            </div>
          </template>

          <!-- ─── 场景 2: 手机移动浏览器（非微信） ─── -->
          <template v-else-if="env.isMobile && hasLaunch">
            <el-button type="success" size="large" class="primary-btn" @click="launchMobile">
              <el-icon><Promotion /></el-icon>
              在微信小程序中打开
            </el-button>
            <p class="device-tip">点击后将拉起手机微信</p>
          </template>

          <!-- ─── 场景 3: PC 浏览器（非微信）─── -->
          <template v-else-if="env.isPC && hasLaunch">
            <el-button type="success" size="large" class="primary-btn" @click="launchPC">
              <el-icon><Promotion /></el-icon>
              在 PC 微信中打开
            </el-button>
            <p class="device-tip">需先登录 PC 端微信 · 或扫描右侧二维码用手机微信打开</p>
          </template>

          <!-- ─── H5 兜底下载 ─── -->
          <el-button
            :type="hasLaunch ? 'default' : 'primary'"
            :size="hasLaunch ? 'default' : 'large'"
            @click="handleDownload"
            :loading="downloading"
            class="fallback-btn"
            :class="{ 'primary-btn': !hasLaunch }"
            :plain="hasLaunch"
          >
            <el-icon><Download /></el-icon>
            {{ hasLaunch ? '直接下载附件（H5）' : '立即下载' }}
          </el-button>

          <p class="hint" v-if="hasLaunch">
            ⭐ 推荐在小程序内打开 · 作者会持续更新内容
          </p>
        </div>

        <!-- ─── PC 端二维码（与按钮并存）─── -->
        <div v-if="env.isPC && hasLaunch && qrDataUrl" class="qr-pane">
          <div class="qr-title">手机扫码打开</div>
          <img :src="qrDataUrl" alt="小程序二维码" class="qr-img" />
          <p class="qr-tip">用手机微信扫一扫</p>
        </div>
      </div>

      <div class="footer">
        <p>由 <strong>附件通</strong> 提供技术支持</p>
        <p class="privacy">
          <el-link type="info" :underline="false" href="https://fujian.5g6g.top/privacy" target="_blank">
            隐私政策
          </el-link>
          ·
          <el-link type="info" :underline="false" href="https://fujian.5g6g.top/terms" target="_blank">
            服务条款
          </el-link>
        </p>
      </div>
    </div>

    <!-- 超时引导弹窗 -->
    <el-dialog v-model="schemeTimeoutDialog" title="未能打开微信？" width="320" align-center>
      <p>可能原因：</p>
      <ul style="padding-left:18px;margin:8px 0;">
        <li>未安装 PC 微信</li>
        <li>未登录 PC 微信</li>
        <li>浏览器拦截了协议跳转</li>
      </ul>
      <el-button type="success" plain @click="downloadWxClient">下载 PC 微信</el-button>
      <el-button @click="schemeTimeoutDialog = false; handleDownload()">改用 H5 下载</el-button>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'
import QRCode from 'qrcode'

const route = useRoute()
const code = route.params.code

const loading = ref(true)
const error = ref('')
const downloading = ref(false)
const file = ref(null)
const launchInfo = ref(null)
const wxLaunchReady = ref(false)
const qrDataUrl = ref('')
const schemeTimeoutDialog = ref(false)

// 环境检测
const env = computed(() => {
  const ua = navigator.userAgent.toLowerCase()
  const isWeixin = /micromessenger/.test(ua)
  const isMobile = /android|iphone|ipad|ipod/.test(ua)
  return {
    isWeixin,
    isMobile: isMobile && !isWeixin,
    isPC: !isMobile && !isWeixin,
    isAndroid: /android/.test(ua),
    isIOS: /iphone|ipad|ipod/.test(ua),
  }
})

const hasLaunch = computed(() => launchInfo.value && launchInfo.value.appid)

const extClass = computed(() => {
  if (!file.value) return ''
  const e = file.value.ext.toLowerCase()
  if (['pdf'].includes(e)) return 'ext-pdf'
  if (['doc', 'docx'].includes(e)) return 'ext-doc'
  if (['xls', 'xlsx'].includes(e)) return 'ext-xls'
  if (['ppt', 'pptx'].includes(e)) return 'ext-ppt'
  if (['zip', 'rar', '7z'].includes(e)) return 'ext-zip'
  return 'ext-default'
})

function formatSize(bytes) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0, size = bytes
  while (size >= 1024 && i < units.length - 1) { size /= 1024; i++ }
  return size.toFixed(i > 0 ? 1 : 0) + ' ' + units[i]
}

function handleDownload() {
  downloading.value = true
  window.location.href = `/api/v1/c/download/${code}`
  setTimeout(() => {
    downloading.value = false
    if (file.value) file.value.download_count++
  }, 1500)
}

// 手机移动浏览器：直接跳 url_scheme
function launchMobile() {
  const url = launchInfo.value.url_scheme || launchInfo.value.url_link
  if (url) window.location.href = url
}

// PC 浏览器：跳 url_scheme + 超时检测
function launchPC() {
  const url = launchInfo.value.url_scheme || launchInfo.value.url_link
  if (!url) return

  // 借鉴 xzfzs：用隐藏 iframe 触发 scheme，避免页面跳转
  const iframe = document.createElement('iframe')
  iframe.style.display = 'none'
  iframe.src = url
  document.body.appendChild(iframe)

  let visibilityChanged = false
  const onVisible = () => {
    if (document.hidden) visibilityChanged = true
  }
  document.addEventListener('visibilitychange', onVisible)

  // 2.5 秒检测：若页面仍在前台，说明 scheme 未被拉起
  setTimeout(() => {
    document.removeEventListener('visibilitychange', onVisible)
    document.body.removeChild(iframe)
    if (!visibilityChanged) {
      schemeTimeoutDialog.value = true
    }
  }, 2500)
}

function downloadWxClient() {
  window.open('https://pc.weixin.qq.com/', '_blank')
}

function onLaunchSuccess() { console.log('成功拉起小程序') }
function onLaunchError(e) { console.warn('拉起小程序失败', e) }

async function initWxLaunch() {
  if (!env.value.isWeixin) return
  // V1.5 接入 JSSDK signature 后启用：当前仅展示静态引导文案
  // wxLaunchReady.value = true
}

// 生成 PC 端二维码（用 url_link，因 https 链接在手机微信扫码后能正常拉起）
async function genQR() {
  if (!env.value.isPC || !launchInfo.value) return
  const target = launchInfo.value.url_link || launchInfo.value.url_scheme
  if (!target) return
  try {
    qrDataUrl.value = await QRCode.toDataURL(target, {
      width: 180,
      margin: 1,
      color: { dark: '#222', light: '#fff' },
    })
  } catch (e) {
    console.warn('二维码生成失败', e)
  }
}

onMounted(async () => {
  try {
    const [fileRes, launchRes] = await Promise.all([
      axios.get(`/api/v1/c/file/${code}`),
      axios.get(`/api/v1/c/launch/${code}`).catch(() => ({ data: { code: 1 } })),
    ])

    if (fileRes.data.code === 0) {
      file.value = fileRes.data.data
    } else {
      error.value = fileRes.data.msg || '文件不存在'
      return
    }

    if (launchRes.data.code === 0) {
      launchInfo.value = launchRes.data.data
    }

    await initWxLaunch()
    await genQR()
  } catch (e) {
    error.value = e.response?.data?.msg || '文件不存在或已下架'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.download-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}
.download-card {
  width: 100%;
  max-width: 480px;
  background: #fff;
  border-radius: 16px;
  padding: 36px 28px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.2);
}
.download-card.is-pc {
  max-width: 720px;
}
.brand { text-align: center; margin-bottom: 28px; }
.brand h1 { margin: 0; font-size: 28px; color: #333; letter-spacing: 2px; }
.brand .tagline { margin: 4px 0 0; color: #999; font-size: 13px; }

.state { text-align: center; padding: 60px 0; color: #999; }
.state.error { color: #f56c6c; }
.state.error h2 { margin: 16px 0 8px; font-size: 20px; }

.layout {
  display: flex;
  gap: 28px;
  align-items: stretch;
}
.file-info {
  flex: 1;
  text-align: center;
  min-width: 0;
}
.qr-pane {
  width: 220px;
  padding: 20px 16px;
  border-left: 1px dashed #eee;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.qr-title {
  font-size: 14px;
  color: #666;
  margin-bottom: 10px;
  font-weight: 600;
}
.qr-img {
  width: 180px;
  height: 180px;
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 8px;
  background: #fff;
}
.qr-tip {
  font-size: 12px;
  color: #999;
  margin: 10px 0 0;
}

.file-icon {
  width: 80px; height: 80px; margin: 0 auto 18px;
  border-radius: 12px; display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 18px; font-weight: 700; letter-spacing: 1px;
}
.ext-pdf { background: #f56c6c; }
.ext-doc { background: #409eff; }
.ext-xls { background: #67c23a; }
.ext-ppt { background: #e6a23c; }
.ext-zip { background: #909399; }
.ext-default { background: #5e72e4; }

.filename {
  margin: 0 0 10px; font-size: 17px; color: #333;
  word-break: break-all; line-height: 1.5;
}
.meta {
  display: flex; flex-wrap: wrap; justify-content: center;
  gap: 8px 14px; font-size: 12px; color: #888; margin-bottom: 24px;
}
.meta span { display: inline-flex; align-items: center; gap: 4px; }

.action-block { margin-bottom: 12px; }
.launch-tag { display: block; width: 100%; }

.primary-btn {
  width: 100%; height: 48px; font-size: 16px;
  border-radius: 8px; margin-bottom: 8px;
}
.fallback-btn {
  width: 100%; height: 40px; font-size: 14px;
  border-radius: 8px; margin-top: 4px;
}
.device-tip {
  margin: 4px 0 12px; font-size: 12px; color: #aaa;
}
.weixin-hint {
  background: #f0f9eb; border: 1px solid #c2e7b0;
  border-radius: 8px; padding: 14px 16px; margin-bottom: 12px;
  font-size: 13px; color: #67c23a;
}
.weixin-hint p { margin: 0; line-height: 1.6; }
.weixin-hint .sub { color: #555; font-size: 12px; margin-top: 4px; }

.hint { margin-top: 14px; color: #aaa; font-size: 12px; }
.footer {
  margin-top: 28px; padding-top: 18px; border-top: 1px solid #eee;
  text-align: center; font-size: 12px; color: #999;
}
.footer p { margin: 4px 0; }
.privacy { color: #ccc; }

@media (max-width: 600px) {
  .layout { flex-direction: column; }
  .qr-pane { display: none; }
}
</style>
