<template>
  <div class="login-page">
    <div class="login-card">
      <h1 class="login-title">附件通</h1>
      <p class="login-subtitle">B 端管理后台</p>

      <el-tabs v-model="loginType">
        <el-tab-pane label="管理员登录" name="admin">
          <el-form @submit.prevent="handleAdminLogin" class="login-form">
            <el-form-item>
              <el-input v-model="form.username" placeholder="管理员账号" prefix-icon="User" size="large" />
            </el-form-item>
            <el-form-item>
              <el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" size="large" show-password />
            </el-form-item>
            <el-button type="primary" size="large" :loading="loading" @click="handleAdminLogin" style="width:100%">
              登 录
            </el-button>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="微信扫码登录" name="wechat">
          <div class="wechat-login">
            <el-button type="success" size="large" @click="handleWechatLogin" :loading="loading">
              <el-icon><Connection /></el-icon>
              获取微信扫码链接
            </el-button>
            <p class="tip">点击后跳转至微信扫码授权页面</p>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import http from '@/utils/http'

const router = useRouter()
const userStore = useUserStore()

const loginType = ref('admin')
const loading = ref(false)
const form = ref({ username: '', password: '' })

async function handleAdminLogin() {
  if (!form.value.username || !form.value.password) {
    ElMessage.warning('请输入账号密码')
    return
  }
  loading.value = true
  try {
    await userStore.adminLogin(form.value.username, form.value.password)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (e) {
    ElMessage.error(e.message || '登录失败')
  } finally {
    loading.value = false
  }
}

async function handleWechatLogin() {
  loading.value = true
  try {
    const { data } = await http.get('/api/v1/auth/scan/url')
    if (data.code === 0) {
      window.location.href = data.data.url
    } else {
      ElMessage.error(data.msg)
    }
  } catch (e) {
    ElMessage.error('获取扫码链接失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.login-card {
  width: 400px;
  padding: 40px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.15);
}
.login-title {
  text-align: center;
  font-size: 28px;
  margin: 0 0 4px 0;
  color: #333;
}
.login-subtitle {
  text-align: center;
  color: #999;
  margin: 0 0 24px 0;
}
.login-form {
  margin-top: 16px;
}
.wechat-login {
  text-align: center;
  padding: 30px 0;
}
.tip {
  color: #999;
  font-size: 12px;
  margin-top: 12px;
}
</style>
