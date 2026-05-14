<template>
  <div class="v2-hub">
    <div class="v2-header">
      <h1>运维中心</h1>
      <p>一站式管理：用户 / 结算 / 平台设置（V3.0 MVP 合并版）</p>
    </div>

    <el-tabs v-model="activeTab" class="v2-tabs" type="border-card">
      <el-tab-pane label="用户管理" name="users">
        <UserList />
      </el-tab-pane>
      <el-tab-pane label="结算审核" name="settlements">
        <SettlementList />
      </el-tab-pane>
      <el-tab-pane label="系统设置" name="settings">
        <Settings />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import UserList from './UserList.vue'
import SettlementList from './SettlementList.vue'
import Settings from './Settings.vue'

const activeTab = ref('users')
</script>

<style scoped>
.v2-hub {
  padding: 0;
}
.v2-header h1 {
  font-size: 22px;
  font-weight: 700;
  color: #1890ff;
  margin: 0 0 4px 0;
}
.v2-header p {
  color: #64748b;
  margin: 0 0 16px 0;
  font-size: 13px;
}
.v2-tabs {
  background: #fff;
  border-radius: 8px;
}
:deep(.el-tabs__content) {
  padding: 16px;
}
</style>
