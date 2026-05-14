<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapsed ? '64px' : '220px'" class="layout-aside">
      <div class="logo">
        <span v-if="!isCollapsed">附件通</span>
        <span v-else>附</span>
      </div>
      <el-menu
        :default-active="$route.path"
        :collapse="isCollapsed"
        router
        class="layout-menu"
        background-color="#001529"
        text-color="rgba(255,255,255,0.65)"
        active-text-color="#fff"
      >
        <!-- 仪表盘（所有人） -->
        <el-menu-item index="/dashboard">
          <el-icon><DataAnalysis /></el-icon>
          <template #title>仪表盘</template>
        </el-menu-item>

        <!-- ─── 作者工作台（基础角色，所有 B 端用户均有） ─── -->
        <el-menu-item-group>
          <template #title>
            <span class="group-title">作者工作台</span>
          </template>
          <el-menu-item index="/files">
            <el-icon><Document /></el-icon>
            <template #title>附件管理</template>
          </el-menu-item>
          <el-menu-item index="/my-revenue">
            <el-icon><Coin /></el-icon>
            <template #title>我的收益</template>
          </el-menu-item>
        </el-menu-item-group>

        <!-- ─── 推广员工作台（is_promoter=true） ─── -->
        <el-menu-item-group v-if="userStore.isPromoter">
          <template #title>
            <span class="group-title">推广员工作台</span>
          </template>
          <el-menu-item index="/promoter/qrcode">
            <el-icon><Picture /></el-icon>
            <template #title>我的二维码</template>
          </el-menu-item>
          <el-menu-item index="/promoter/authors">
            <el-icon><UserFilled /></el-icon>
            <template #title>我的作者</template>
          </el-menu-item>
          <el-menu-item index="/promoter/commission">
            <el-icon><Wallet /></el-icon>
            <template #title>推广收益</template>
          </el-menu-item>
        </el-menu-item-group>

        <!-- ─── 平台管理（is_admin=true） ─── -->
        <el-menu-item-group v-if="userStore.isAdmin">
          <template #title>
            <span class="group-title">平台管理</span>
          </template>
          <el-menu-item index="/users">
            <el-icon><User /></el-icon>
            <template #title>用户管理</template>
          </el-menu-item>
          <el-menu-item index="/settlements">
            <el-icon><Money /></el-icon>
            <template #title>结算审核</template>
          </el-menu-item>
          <el-menu-item index="/deployments">
            <el-icon><Upload /></el-icon>
            <template #title>部署管理</template>
          </el-menu-item>
          <el-menu-item index="/settings">
            <el-icon><Setting /></el-icon>
            <template #title>系统设置</template>
          </el-menu-item>
        </el-menu-item-group>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="layout-header">
        <el-icon class="collapse-btn" @click="isCollapsed = !isCollapsed">
          <Fold v-if="!isCollapsed" />
          <Expand v-else />
        </el-icon>
        <div class="header-right">
          <el-tag v-if="userStore.isAdmin" type="danger" size="small" effect="dark">管理员</el-tag>
          <el-tag v-if="userStore.isPromoter" type="warning" size="small" effect="dark">推广员</el-tag>
          <el-tag v-if="!userStore.isAdmin && !userStore.isPromoter" type="info" size="small">作者</el-tag>
          <span class="username">{{ userStore.user?.nickname || userStore.user?.admin_username }}</span>
          <el-dropdown>
            <el-avatar :size="32" :src="userStore.user?.avatar_url || ''">
              {{ (userStore.user?.nickname || 'U')[0] }}
            </el-avatar>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="handleLogout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const router = useRouter()
const isCollapsed = ref(false)

function handleLogout() {
  userStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}
.layout-aside {
  background: #001529;
  transition: width 0.3s;
  overflow: hidden;
}
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 20px;
  font-weight: bold;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}
.layout-menu {
  border-right: none;
  background: #001529;
}
.layout-menu:not(.el-menu--collapse) {
  width: 220px;
}
:deep(.el-menu) {
  background: #001529;
}
:deep(.el-menu-item) {
  color: rgba(255,255,255,0.65);
}
:deep(.el-menu-item:hover),
:deep(.el-menu-item.is-active) {
  color: #fff;
  background: #1890ff;
}
.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #f0f0f0;
  padding: 0 20px;
  background: #fff;
}
.collapse-btn {
  font-size: 20px;
  cursor: pointer;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
.username {
  font-size: 14px;
  color: #333;
}
.layout-main {
  background: #f5f5f5;
  padding: 20px;
}
:deep(.el-menu-item-group__title) {
  padding-left: 20px !important;
  color: rgba(255,255,255,0.4) !important;
  font-size: 12px;
  letter-spacing: 1px;
  margin-top: 8px;
}
.group-title {
  font-weight: 600;
}
</style>
