<template>
  <div class="dashboard">
    <h2>仪表盘</h2>

    <!-- 管理员视角 -->
    <template v-if="userStore.isAdmin">
      <el-row :gutter="16" class="stat-row">
        <el-col :span="6">
          <el-card shadow="hover">
            <el-statistic title="作者总数" :value="stats.total_authors" />
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <el-statistic title="推广员数" :value="stats.total_promoters" />
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <el-statistic title="本月平台收入(元)" :value="(stats.month_platform_income / 100).toFixed(2)" />
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <el-statistic title="待处理结算单" :value="stats.pending_commissions" />
          </el-card>
        </el-col>
      </el-row>

      <!-- 待办事项 -->
      <el-card class="todo-card" v-if="todo">
        <template #header>待办事项</template>
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="结算单待审核">{{ todo.settlements_to_approve }}</el-descriptions-item>
          <el-descriptions-item label="结算单待打款">{{ todo.settlements_to_pay }}</el-descriptions-item>
          <el-descriptions-item label="文件待复审">{{ todo.files_to_review }}</el-descriptions-item>
        </el-descriptions>
      </el-card>
    </template>

    <!-- 作者视角 -->
    <template v-else>
      <el-card>
        <template #header>快速操作</template>
        <el-space>
          <el-button type="primary" @click="$router.push('/files')">管理附件</el-button>
          <el-button @click="$router.push('/files')">上传新文件</el-button>
        </el-space>
      </el-card>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import http from '@/utils/http'

const userStore = useUserStore()
const stats = ref({})
const todo = ref(null)

onMounted(async () => {
  if (userStore.isAdmin) {
    try {
      const [dashRes, todoRes] = await Promise.all([
        http.get('/api/v1/admin/dashboard'),
        http.get('/api/v1/admin/todo'),
      ])
      if (dashRes.data.code === 0) stats.value = dashRes.data.data
      if (todoRes.data.code === 0) todo.value = todoRes.data.data
    } catch (e) { /* ignore */ }
  }
})
</script>

<style scoped>
.dashboard h2 {
  margin: 0 0 20px 0;
}
.stat-row {
  margin-bottom: 20px;
}
.todo-card {
  margin-top: 16px;
}
</style>
