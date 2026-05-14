<template>
  <div class="settings-page">
    <h2 class="page-title">系统设置</h2>

    <el-tabs v-model="activeTab" class="settings-tabs">
      <!-- ─── 修改密码 ─── -->
      <el-tab-pane label="修改密码" name="password">
        <el-card shadow="never" class="settings-card">
          <el-form
            ref="pwdFormRef"
            :model="pwdForm"
            :rules="pwdRules"
            label-width="120px"
            style="max-width: 520px"
          >
            <el-form-item label="原密码" prop="old_password">
              <el-input v-model="pwdForm.old_password" type="password" show-password placeholder="请输入当前密码" />
            </el-form-item>
            <el-form-item label="新密码" prop="new_password">
              <el-input v-model="pwdForm.new_password" type="password" show-password placeholder="至少 8 位" />
            </el-form-item>
            <el-form-item label="确认新密码" prop="confirm_password">
              <el-input v-model="pwdForm.confirm_password" type="password" show-password placeholder="再次输入新密码" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="pwdLoading" @click="submitPassword">保存</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- ─── 微信开放平台 ─── -->
      <el-tab-pane label="微信开放平台" name="wechat">
        <el-card shadow="never" class="settings-card">
          <el-alert
            type="info"
            :closable="false"
            show-icon
            title="提示"
            description="敏感字段（AppSecret / Token / EncodingAESKey）保存后只显示末 4 位，留空提交将保留原值；填入新值则覆盖。"
            style="margin-bottom: 20px"
          />

          <h4 class="section-title">第三方平台（小程序代运营）</h4>
          <el-form :model="wxForm" label-width="200px" style="max-width: 720px">
            <el-form-item label="Component AppID">
              <el-input v-model="wxForm.wx_component_appid" placeholder="wx开头的第三方平台AppID" />
            </el-form-item>
            <el-form-item label="Component AppSecret">
              <el-input
                v-model="wxForm.wx_component_appsecret"
                type="password"
                show-password
                :placeholder="secretSet.wx_component_appsecret ? '已配置（留空保留原值）' : '请输入第三方平台 AppSecret'"
              />
            </el-form-item>
            <el-form-item label="Token">
              <el-input
                v-model="wxForm.wx_component_token"
                type="password"
                show-password
                :placeholder="secretSet.wx_component_token ? '已配置（留空保留原值）' : '消息验证 Token'"
              />
            </el-form-item>
            <el-form-item label="EncodingAESKey">
              <el-input
                v-model="wxForm.wx_component_encoding_aes_key"
                type="password"
                show-password
                :placeholder="secretSet.wx_component_encoding_aes_key ? '已配置（留空保留原值）' : '43 位消息加解密密钥'"
              />
            </el-form-item>
          </el-form>

          <el-divider />

          <h4 class="section-title">网站应用（扫码登录）</h4>
          <el-form :model="wxForm" label-width="200px" style="max-width: 720px">
            <el-form-item label="网站应用 AppID">
              <el-input v-model="wxForm.wechat_open_appid" placeholder="open.weixin.qq.com 申请的网站应用 AppID" />
            </el-form-item>
            <el-form-item label="网站应用 AppSecret">
              <el-input
                v-model="wxForm.wechat_open_appsecret"
                type="password"
                show-password
                :placeholder="secretSet.wechat_open_appsecret ? '已配置（留空保留原值）' : '网站应用 AppSecret'"
              />
            </el-form-item>
            <el-form-item label="回调 URL">
              <el-input v-model="wxForm.wechat_open_redirect_uri" placeholder="https://fujian.5g6g.top/api/v1/auth/wechat/login" />
            </el-form-item>
          </el-form>

          <el-form-item>
            <el-button type="primary" :loading="saveLoading" @click="saveSettings">保存微信配置</el-button>
            <el-button @click="loadSettings">重新加载</el-button>
            <el-button type="danger" plain @click="clearWebsiteApp">清除网站应用配置</el-button>
          </el-form-item>
        </el-card>
      </el-tab-pane>

      <!-- ─── 对象存储 COS ─── -->
      <el-tab-pane label="对象存储（COS）" name="cos">
        <el-card shadow="never" class="settings-card">
          <el-alert
            type="warning"
            :closable="false"
            show-icon
            title="切换前请确认"
            description="从 local 切换到 cos 后，新上传的文件直传 COS；历史文件不会自动迁移，需手动同步。建议先在测试环境验证。"
            style="margin-bottom: 20px"
          />

          <el-form :model="cosForm" label-width="160px" style="max-width: 720px">
            <el-form-item label="存储驱动">
              <el-radio-group v-model="cosForm.storage_driver">
                <el-radio label="local">本地存储（默认）</el-radio>
                <el-radio label="cos">腾讯云 COS</el-radio>
              </el-radio-group>
            </el-form-item>

            <template v-if="cosForm.storage_driver === 'cos'">
              <el-form-item label="Bucket URL">
                <el-input v-model="cosForm.cos_bucket_url" placeholder="https://xxx-1234567890.cos.ap-guangzhou.myqcloud.com" />
              </el-form-item>
              <el-form-item label="SecretId">
                <el-input
                  v-model="cosForm.cos_secret_id"
                  type="password"
                  show-password
                  :placeholder="secretSet.cos_secret_id ? '已配置（留空保留原值）' : '腾讯云 SecretId'"
                />
              </el-form-item>
              <el-form-item label="SecretKey">
                <el-input
                  v-model="cosForm.cos_secret_key"
                  type="password"
                  show-password
                  :placeholder="secretSet.cos_secret_key ? '已配置（留空保留原值）' : '腾讯云 SecretKey'"
                />
              </el-form-item>
              <el-form-item label="CDN 加速域名">
                <el-input v-model="cosForm.cdn_base_url" placeholder="可选，如 https://cdn.fujian.5g6g.top" />
              </el-form-item>
            </template>

            <el-form-item>
              <el-button type="primary" :loading="saveLoading" @click="saveSettings">保存存储配置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import http from '@/utils/http'

const activeTab = ref('password')

// ─── 修改密码 ───
const pwdFormRef = ref()
const pwdLoading = ref(false)
const pwdForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})
const pwdRules = {
  old_password: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, message: '密码至少 8 位', trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (_, v, cb) => {
        if (v !== pwdForm.new_password) return cb(new Error('两次输入不一致'))
        cb()
      },
      trigger: 'blur',
    },
  ],
}

async function submitPassword() {
  await pwdFormRef.value.validate()
  pwdLoading.value = true
  try {
    const { data } = await http.post('/api/v1/admin/settings/password', {
      old_password: pwdForm.old_password,
      new_password: pwdForm.new_password,
    })
    if (data.code === 0) {
      ElMessage.success(data.msg || '密码修改成功')
      Object.assign(pwdForm, { old_password: '', new_password: '', confirm_password: '' })
    } else {
      ElMessage.error(data.msg || '修改失败')
    }
  } finally {
    pwdLoading.value = false
  }
}

// ─── 平台凭据 ───
const saveLoading = ref(false)
const secretSet = ref({})

const wxForm = reactive({
  wx_component_appid: '',
  wx_component_appsecret: '',
  wx_component_token: '',
  wx_component_encoding_aes_key: '',
  wechat_open_appid: '',
  wechat_open_appsecret: '',
  wechat_open_redirect_uri: '',
})

const cosForm = reactive({
  storage_driver: 'local',
  cos_bucket_url: '',
  cos_secret_id: '',
  cos_secret_key: '',
  cdn_base_url: '',
})

async function loadSettings() {
  const { data } = await http.get('/api/v1/admin/settings/platform')
  if (data.code !== 0) {
    ElMessage.error(data.msg || '加载失败')
    return
  }
  const d = data.data || {}
  secretSet.value = d._secret_set || {}

  // 非敏感字段直接填入；敏感字段清空（让占位符提示"已配置"）
  wxForm.wx_component_appid = d.wx_component_appid || ''
  wxForm.wechat_open_appid = d.wechat_open_appid || ''
  wxForm.wechat_open_redirect_uri = d.wechat_open_redirect_uri || ''
  wxForm.wx_component_appsecret = ''
  wxForm.wx_component_token = ''
  wxForm.wx_component_encoding_aes_key = ''
  wxForm.wechat_open_appsecret = ''

  cosForm.storage_driver = d.storage_driver || 'local'
  cosForm.cos_bucket_url = d.cos_bucket_url || ''
  cosForm.cdn_base_url = d.cdn_base_url || ''
  cosForm.cos_secret_id = ''
  cosForm.cos_secret_key = ''
}

async function saveSettings() {
  // 切换到 COS 前提示
  if (cosForm.storage_driver === 'cos' && !cosForm.cos_bucket_url) {
    ElMessage.warning('请先填写 Bucket URL')
    return
  }
  if (cosForm.storage_driver === 'cos') {
    try {
      await ElMessageBox.confirm('确认切换到腾讯云 COS 吗？历史文件不会自动迁移。', '切换存储驱动', {
        type: 'warning',
      })
    } catch {
      return
    }
  }

  saveLoading.value = true
  try {
    const payload = { ...wxForm, ...cosForm }
    const { data } = await http.put('/api/v1/admin/settings/platform', payload)
    if (data.code === 0) {
      ElMessage.success(data.msg || '保存成功')
      await loadSettings()
    } else {
      ElMessage.error(data.msg || '保存失败')
    }
  } finally {
    saveLoading.value = false
  }
}

async function clearWebsiteApp() {
  try {
    await ElMessageBox.confirm(
      '将清除网站应用（扫码登录）所有配置，第三方平台配置不受影响。继续吗？',
      '清除网站应用配置',
      { type: 'warning' }
    )
  } catch {
    return
  }
  saveLoading.value = true
  try {
    const { data } = await http.put('/api/v1/admin/settings/platform', {
      wechat_open_appid: '',
      wechat_open_appsecret: '__CLEAR__',
      wechat_open_redirect_uri: '',
    })
    if (data.code === 0) {
      ElMessage.success('已清除网站应用配置')
      await loadSettings()
    } else {
      ElMessage.error(data.msg || '清除失败')
    }
  } finally {
    saveLoading.value = false
  }
}

onMounted(loadSettings)
</script>

<style scoped>
.settings-page {
  padding: 0;
}
.page-title {
  margin: 0 0 16px;
  font-size: 18px;
  font-weight: 600;
}
.settings-tabs {
  background: #fff;
  padding: 16px 20px 24px;
  border-radius: 8px;
}
.settings-card {
  border: none;
  box-shadow: none;
}
.section-title {
  margin: 0 0 16px;
  font-size: 14px;
  font-weight: 600;
  color: #409eff;
}
</style>
