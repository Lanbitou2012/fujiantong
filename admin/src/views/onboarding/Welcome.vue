<template>
  <div class="join-page">
    <!-- Hero -->
    <section class="hero">
      <div class="hero-bg"></div>
      <div class="hero-content">
        <div class="brand">
          <div class="brand-logo">附</div>
          <span class="brand-name">附件通</span>
        </div>
        <h1 class="hero-title">让公众号文章<br />轻松挂载附件</h1>
        <p class="hero-sub">
          一次授权，自动部署小程序<br />
          流量主收益 72% 归您，平台不经手资金
        </p>
        <div class="hero-stats">
          <div class="stat-item">
            <div class="stat-num">0元</div>
            <div class="stat-label">入驻费用</div>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-item">
            <div class="stat-num">3 分钟</div>
            <div class="stat-label">自动部署</div>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-item">
            <div class="stat-num">72%</div>
            <div class="stat-label">收益归您</div>
          </div>
        </div>
      </div>
    </section>

    <!-- 推广员信息 -->
    <section v-if="promoterID" class="section">
      <div class="promoter-banner">
        <el-icon><User /></el-icon>
        <span>由推广员 <strong>{{ promoterID }}</strong> 邀请入驻，绑定后您的小程序收益将自动结算 8.4% 推广佣金给推荐人</span>
      </div>
    </section>

    <!-- 价值主张 -->
    <section class="section">
      <h2 class="section-title">为什么选择附件通？</h2>
      <div class="features">
        <div class="feature">
          <div class="feature-icon">🚀</div>
          <h3>3 分钟极速接入</h3>
          <p>授权 → 自动部署 → 提交审核，无需写任何代码</p>
        </div>
        <div class="feature">
          <div class="feature-icon">💰</div>
          <h3>收益归你 72%</h3>
          <p>微信广告半月自动分账，作者直收，平台不经手资金</p>
        </div>
        <div class="feature">
          <div class="feature-icon">📎</div>
          <h3>支持 Word/PDF/Excel</h3>
          <p>公众号文章一键挂载下载，文本超链接秒级生效</p>
        </div>
        <div class="feature">
          <div class="feature-icon">🛡️</div>
          <h3>合规与安全</h3>
          <p>持牌运营，数据加密传输，授权随时可撤销</p>
        </div>
      </div>
    </section>

    <!-- 接入流程 -->
    <section class="section section-alt">
      <h2 class="section-title">3 步开始变现</h2>
      <div class="steps">
        <div class="step">
          <div class="step-num">01</div>
          <div class="step-content">
            <h3>注册个人小程序</h3>
            <p>在 mp.weixin.qq.com 用本人微信注册个人主体小程序，拿到 AppID（首次约 10 分钟）</p>
            <el-link type="primary" href="https://mp.weixin.qq.com/cgi-bin/registermidpage?action=index" target="_blank">
              前往微信公众平台注册 →
            </el-link>
          </div>
        </div>
        <div class="step-arrow">→</div>
        <div class="step">
          <div class="step-num">02</div>
          <div class="step-content">
            <h3>授权附件通代运营</h3>
            <p>点击下方"立即授权"，用本人微信扫码 → 选择上一步注册的小程序 → 勾选"流量主代运营 + 代码管理"权限</p>
          </div>
        </div>
        <div class="step-arrow">→</div>
        <div class="step">
          <div class="step-num">03</div>
          <div class="step-content">
            <h3>系统自动部署</h3>
            <p>平台自动上传代码、配置广告位、提交审核。1-2 工作日通过后小程序上线，您即可在公众号文章插入附件</p>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="section join-section">
      <div class="join-card">
        <h2>立即接入</h2>
        <p v-if="promoterID" class="promoter-tip">
          <span>👤 推荐人 ID: <strong>{{ promoterID }}</strong></span>
        </p>

        <el-checkbox v-model="agreed" class="agreement">
          我已阅读并同意
          <a href="javascript:;" @click.stop="showAgreement = true">《附件通服务协议》</a>
          和
          <a href="javascript:;" @click.stop="showPrivacy = true">《隐私政策》</a>
        </el-checkbox>

        <button class="cta-btn" :disabled="!agreed || loading" @click="startAuth">
          <span v-if="!loading">立即授权接入</span>
          <span v-else>正在生成授权链接...</span>
        </button>

        <div class="security-tips">
          <span>🔒 微信官方授权</span>
          <span>·</span>
          <span>支持随时撤销</span>
          <span>·</span>
          <span>权限最小化</span>
        </div>

        <p class="help-text">
          没有小程序？请先 <el-link type="primary" href="https://mp.weixin.qq.com/cgi-bin/registermidpage?action=index" target="_blank">前往微信公众平台注册</el-link>
        </p>
      </div>
    </section>

    <!-- 服务协议弹窗 -->
    <el-dialog v-model="showAgreement" title="附件通服务协议" width="640px" align-center>
      <div class="agreement-content">
        <h4>一、服务内容</h4>
        <p>附件通（以下简称"平台"）为微信小程序运营者提供附件分发、广告变现等代运营服务。作者授权后，平台提供：</p>
        <ul>
          <li>代部署小程序代码、提交审核与发布</li>
          <li>开通流量主资格并自动配置广告位</li>
          <li>文件管理、用户运营、流量统计后台</li>
        </ul>
        <h4>二、收益分配</h4>
        <p>作者通过平台部署的小程序产生的广告收益，由微信广告系统每半月自动结算，按 <strong>作者 72% / 平台 28%</strong> 比例直接打入双方对应账户，<strong>平台不经手资金</strong>。</p>
        <h4>三、双方权责</h4>
        <p><strong>平台承诺：</strong>不修改作者已有内容、不滥用作者数据；作者可随时通过微信公众平台撤销授权。</p>
        <p><strong>作者承诺：</strong>所发布内容遵守法律法规，不发布违法违规、低俗、虚假信息。</p>
        <h4>四、协议终止</h4>
        <p>作者可随时通过微信公众平台撤销授权，平台亦保留对违规账号停止服务的权利。协议终止后，已产生的收益按上述比例继续结算至作者账户。</p>
      </div>
      <template #footer>
        <el-button type="primary" @click="showAgreement = false">我已阅读</el-button>
      </template>
    </el-dialog>

    <!-- 隐私政策弹窗 -->
    <el-dialog v-model="showPrivacy" title="隐私政策" width="640px" align-center>
      <div class="agreement-content">
        <h4>一、收集信息范围</h4>
        <ul>
          <li>小程序基本信息（名称、AppID、主体、二维码）</li>
          <li>小程序授权 token（用于代码部署与广告管理）</li>
          <li>广告结算数据（金额、日期）</li>
        </ul>
        <h4>二、信息使用</h4>
        <p>仅用于完成代运营服务，不会用于其他用途，不会向第三方披露（依法配合监管除外）。</p>
        <h4>三、存储与安全</h4>
        <p>所有敏感信息（如 refresh_token）使用 AES-256 加密存储，传输全程使用 HTTPS。</p>
        <h4>四、用户权利</h4>
        <p>您可随时通过微信公众平台撤销授权；亦可联系客服要求删除已收集的数据。</p>
      </div>
      <template #footer>
        <el-button type="primary" @click="showPrivacy = false">我已阅读</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import http from '@/utils/http'

const route = useRoute()
const promoterID = ref('')
const agreed = ref(false)
const loading = ref(false)
const showAgreement = ref(false)
const showPrivacy = ref(false)

onMounted(() => {
  promoterID.value = route.query.p || ''
})

// 兼容多种错误返回结构（503/400/500），把后端 msg 提取出来给作者看清"为啥不让授权"
function extractError(e) {
  return (
    e?.response?.data?.msg ||
    e?.response?.data?.error ||
    e?.message ||
    '生成授权链接失败'
  )
}

async function startAuth() {
  if (!agreed.value) {
    ElMessage.warning('请先阅读并同意服务协议')
    return
  }
  loading.value = true
  try {
    const { data } = await http.get('/api/v1/auth/wx-component/url', {
      params: promoterID.value ? { promoter_id: promoterID.value } : {},
    })
    if (data.code !== 0 || !data.data?.auth_url) {
      throw new Error(data.msg || '生成授权链接失败')
    }
    // 直跳微信公众平台授权页（同时勾选 135 流量主代运营 + 17 代码管理 双权限集）
    window.location.href = data.data.auth_url
  } catch (e) {
    loading.value = false
    const msg = extractError(e)
    // 平台尚未收到 component_verify_ticket 时给特殊提示，避免作者误以为"系统坏了"
    if (msg.includes('verify_ticket') || msg.includes('未配置')) {
      ElMessageBox.alert(
        '附件通平台正在初始化中（等待微信推送 component_verify_ticket，预计 10 分钟内完成）。请稍后再点击「立即授权接入」。',
        '平台初始化中',
        { confirmButtonText: '我知道了' }
      )
    } else {
      ElMessage.error(msg)
    }
  }
}
</script>

<style scoped>
.join-page {
  min-height: 100vh;
  background: #fff;
  font-family: -apple-system, BlinkMacSystemFont, 'PingFang SC', 'Microsoft YaHei', sans-serif;
}
.hero {
  position: relative;
  padding: 72px 24px 88px;
  overflow: hidden;
  color: #fff;
}
.hero-bg {
  position: absolute; inset: 0;
  background: linear-gradient(135deg, #059669 0%, #10B981 40%, #6366F1 100%);
  z-index: 0;
}
.hero-bg::before {
  content: ''; position: absolute; inset: 0;
  background: radial-gradient(circle at 20% 30%, rgba(255,255,255,0.15), transparent 40%),
              radial-gradient(circle at 80% 70%, rgba(255,255,255,0.1), transparent 40%);
}
.hero-content { position: relative; z-index: 1; max-width: 920px; margin: 0 auto; text-align: center; }
.brand { display: inline-flex; align-items: center; gap: 10px; margin-bottom: 28px; }
.brand-logo {
  width: 36px; height: 36px; border-radius: 10px;
  background: rgba(255,255,255,0.2);
  display: flex; align-items: center; justify-content: center;
  font-weight: 700; font-size: 18px;
}
.brand-name { font-size: 20px; font-weight: 600; letter-spacing: 1px; }
.hero-title { font-size: 44px; font-weight: 800; line-height: 1.2; margin: 0 0 18px; }
.hero-sub { font-size: 17px; opacity: 0.92; line-height: 1.7; margin: 0 0 40px; }
.hero-stats { display: inline-flex; gap: 0; padding: 22px 36px; background: rgba(255,255,255,0.15); border-radius: 16px; }
.stat-item { padding: 0 26px; }
.stat-num { font-size: 30px; font-weight: 800; }
.stat-label { font-size: 13px; opacity: 0.85; margin-top: 4px; }
.stat-divider { width: 1px; background: rgba(255,255,255,0.3); }

.section { padding: 64px 24px; }
.section-alt { background: #FAFAFA; }
.section-title { font-size: 28px; font-weight: 700; text-align: center; margin: 0 0 44px; color: #0F172A; }

.promoter-banner {
  max-width: 720px; margin: 0 auto;
  background: #EEF2FF; border: 1px solid #C7D2FE; border-radius: 12px;
  padding: 14px 20px; display: flex; align-items: center; gap: 10px;
  color: #4F46E5; font-size: 14px;
}

.features { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 20px; max-width: 1080px; margin: 0 auto; }
.feature { padding: 28px 22px; border-radius: 16px; background: #fff; border: 1px solid #F0F0F0; transition: all .3s; }
.feature:hover { transform: translateY(-4px); box-shadow: 0 12px 32px rgba(99,102,241,0.12); }
.feature-icon { font-size: 32px; margin-bottom: 12px; }
.feature h3 { font-size: 17px; font-weight: 700; margin: 0 0 10px; color: #0F172A; }
.feature p { font-size: 14px; color: #64748B; line-height: 1.7; margin: 0; }

.steps { display: flex; align-items: stretch; justify-content: center; gap: 12px; max-width: 1080px; margin: 0 auto; flex-wrap: wrap; }
.step { flex: 1; min-width: 240px; padding: 26px; background: #fff; border-radius: 16px; border: 2px solid #E5E7EB; }
.step-num { font-size: 32px; font-weight: 800; background: linear-gradient(135deg, #10B981, #6366F1); -webkit-background-clip: text; background-clip: text; color: transparent; margin-bottom: 10px; }
.step-content h3 { font-size: 17px; margin: 0 0 8px; color: #0F172A; }
.step-content p { font-size: 14px; color: #64748B; line-height: 1.7; margin: 0 0 8px; }
.step-arrow { font-size: 24px; color: #CBD5E1; align-self: center; font-weight: 300; }

.join-section { background: linear-gradient(180deg, #FAFAFA 0%, #fff 100%); }
.join-card { max-width: 480px; margin: 0 auto; padding: 40px 36px; background: #fff; border-radius: 20px; box-shadow: 0 20px 60px rgba(0,0,0,0.08); text-align: center; }
.join-card h2 { font-size: 26px; margin: 0 0 14px; color: #0F172A; }
.promoter-tip { font-size: 13px; color: #6366F1; background: #EEF2FF; padding: 7px 14px; border-radius: 999px; display: inline-block; margin-bottom: 20px; }
.agreement { display: block; margin: 18px 0; text-align: left; }
.agreement a { color: #6366F1; text-decoration: none; }
.agreement a:hover { text-decoration: underline; }

.cta-btn {
  width: 100%; padding: 15px; border: none; border-radius: 12px; cursor: pointer;
  background: linear-gradient(135deg, #10B981 0%, #6366F1 100%);
  color: #fff; font-size: 16px; font-weight: 600;
  transition: all .3s; box-shadow: 0 8px 24px rgba(99,102,241,0.3);
}
.cta-btn:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 12px 32px rgba(99,102,241,0.4); }
.cta-btn:disabled { opacity: 0.5; cursor: not-allowed; }

.security-tips { margin-top: 16px; font-size: 12px; color: #94A3B8; display: flex; justify-content: center; gap: 8px; flex-wrap: wrap; }
.help-text { margin-top: 20px; font-size: 13px; color: #94A3B8; }

.agreement-content h4 { font-size: 15px; margin: 14px 0 8px; color: #0F172A; }
.agreement-content p, .agreement-content li { font-size: 14px; line-height: 1.8; color: #475569; }
.agreement-content ul { padding-left: 20px; }

@media (max-width: 768px) {
  .hero-title { font-size: 30px; }
  .hero-sub { font-size: 15px; }
  .hero-stats { padding: 14px 20px; flex-wrap: wrap; gap: 6px; }
  .stat-item { padding: 0 12px; }
  .stat-num { font-size: 22px; }
  .step-arrow { display: none; }
  .section { padding: 48px 16px; }
}
</style>
