import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import http from '@/utils/http'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const isAdmin = computed(() => user.value?.is_admin === true)
  const isPromoter = computed(() => user.value?.is_promoter === true)
  const roleName = computed(() => {
    if (isAdmin.value) return 'admin'
    if (isPromoter.value) return 'promoter'
    return 'author'
  })

  function setAuth(t, u) {
    token.value = t
    user.value = u
    localStorage.setItem('token', t)
    localStorage.setItem('user', JSON.stringify(u))
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  async function adminLogin(username, password) {
    const { data } = await http.post('/api/v1/auth/admin/login', { username, password })
    if (data.code === 0) {
      setAuth(data.data.token, data.data.user)
      return true
    }
    throw new Error(data.msg)
  }

  async function wechatLogin(code) {
    const { data } = await http.post('/api/v1/auth/wechat/login', { code })
    if (data.code === 0) {
      setAuth(data.data.token, data.data.user)
      return true
    }
    throw new Error(data.msg)
  }

  return { token, user, isAdmin, isPromoter, roleName, setAuth, logout, adminLogin, wechatLogin }
})
