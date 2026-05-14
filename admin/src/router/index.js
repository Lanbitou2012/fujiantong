import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { public: true },
  },
  {
    path: '/',
    component: () => import('@/layout/MainLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘' },
      },
      {
        path: 'files',
        name: 'Files',
        component: () => import('@/views/author/FileList.vue'),
        meta: { title: '附件管理' },
      },
      {
        path: 'my-revenue',
        name: 'MyRevenue',
        component: () => import('@/views/author/MyRevenue.vue'),
        meta: { title: '我的收益' },
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/admin/UserList.vue'),
        meta: { title: '用户管理', role: 'admin' },
      },
      {
        path: 'settlements',
        name: 'Settlements',
        component: () => import('@/views/admin/SettlementList.vue'),
        meta: { title: '结算管理', role: 'admin' },
      },
      {
        path: 'deployments',
        name: 'Deployments',
        component: () => import('@/views/admin/DeploymentList.vue'),
        meta: { title: '部署管理', role: 'admin' },
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/admin/Settings.vue'),
        meta: { title: '系统设置', role: 'admin' },
      },
      {
        path: 'promoter/qrcode',
        name: 'PromoterQRCode',
        component: () => import('@/views/promoter/MyQRCode.vue'),
        meta: { title: '我的二维码', role: 'promoter' },
      },
      {
        path: 'promoter/authors',
        name: 'PromoterAuthors',
        component: () => import('@/views/promoter/AuthorList.vue'),
        meta: { title: '我的作者', role: 'promoter' },
      },
      {
        path: 'promoter/commission',
        name: 'PromoterCommission',
        component: () => import('@/views/promoter/Commission.vue'),
        meta: { title: '推广收益', role: 'promoter' },
      },
    ],
  },
  {
    path: '/auth-success',
    name: 'AuthSuccess',
    component: () => import('@/views/AuthSuccess.vue'),
    meta: { public: true },
  },
  {
    // C 端 H5 下载页（V2.8 §6 三重复制机制兜底路径）
    path: '/f/:code',
    name: 'FileDownload',
    component: () => import('@/views/public/FileDownload.vue'),
    meta: { public: true },
  },
  // ─── 作者 H5 入驻流程（V2.8 §15.2.3） ───
  {
    path: '/onboarding',
    name: 'OnboardingWelcome',
    component: () => import('@/views/onboarding/Welcome.vue'),
    meta: { public: true },
  },
  {
    path: '/onboarding/register-guide',
    name: 'OnboardingRegisterGuide',
    component: () => import('@/views/onboarding/RegisterGuide.vue'),
    // 该页面需登录后查看
  },
  {
    path: '/onboarding/authorize',
    name: 'OnboardingAuthorize',
    component: () => import('@/views/onboarding/Authorize.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 全局路由守卫：登录态 + 角色权限校验（前端兜底，后端中间件按铁律 4 实时校验为准）
router.beforeEach((to, from, next) => {
  if (to.meta.public) return next()
  const userStore = useUserStore()
  if (!userStore.token) {
    return next('/login')
  }
  // 角色拦截
  if (to.meta.role === 'admin' && !userStore.isAdmin) {
    return next('/dashboard')
  }
  if (to.meta.role === 'promoter' && !userStore.isPromoter) {
    return next('/dashboard')
  }
  next()
})

export default router
