<template>
  <div class="user-list">
    <div class="page-header">
      <h2>用户管理</h2>
      <el-input v-model="keyword" placeholder="搜索昵称/AppID" style="width:240px" @keyup.enter="loadUsers" clearable>
        <template #append>
          <el-button icon="Search" @click="loadUsers" />
        </template>
      </el-input>
    </div>

    <el-table :data="users" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="nickname" label="昵称" width="120" />
      <el-table-column prop="bound_appid" label="绑定 AppID" width="180" />
      <el-table-column label="角色" width="150">
        <template #default="{ row }">
          <el-tag v-if="row.is_admin" type="danger" size="small">管理员</el-tag>
          <el-tag v-if="row.is_promoter" type="warning" size="small">推广员</el-tag>
          <el-tag v-if="!row.is_admin && !row.is_promoter" size="small">作者</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="推广员" width="120">
        <template #default="{ row }">
          {{ row.parent_promoter_user_id || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ row.status === 1 ? '正常' : '封禁' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button size="small" :type="row.is_promoter ? 'warning' : 'primary'" @click="togglePromoter(row)">
            {{ row.is_promoter ? '取消推广员' : '设为推广员' }}
          </el-button>
          <el-button size="small" :type="row.status === 1 ? 'danger' : 'success'" @click="toggleBan(row)">
            {{ row.status === 1 ? '封禁' : '解封' }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-if="total > size"
      :current-page="page"
      :page-size="size"
      :total="total"
      layout="prev, pager, next"
      @current-change="changePage"
      class="pagination"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import http from '@/utils/http'

const users = ref([])
const loading = ref(false)
const page = ref(1)
const size = ref(20)
const total = ref(0)
const keyword = ref('')

async function loadUsers() {
  loading.value = true
  try {
    const { data } = await http.get('/api/v1/admin/users', { params: { page: page.value, size: size.value, keyword: keyword.value } })
    if (data.code === 0) {
      users.value = data.data.list || []
      total.value = data.data.total
    }
  } finally {
    loading.value = false
  }
}

function changePage(p) {
  page.value = p
  loadUsers()
}

async function togglePromoter(row) {
  const enable = !row.is_promoter
  const action = enable ? '设为推广员' : '取消推广员'
  try {
    const { value: reason } = await ElMessageBox.prompt(`请输入${action}原因`, action, { confirmButtonText: '确定', cancelButtonText: '取消' })
    const { data } = await http.put(`/api/v1/admin/users/${row.id}/promoter`, { enable, reason })
    if (data.code === 0) {
      ElMessage.success('操作成功')
      loadUsers()
    } else {
      ElMessage.error(data.msg)
    }
  } catch (e) { /* cancelled */ }
}

async function toggleBan(row) {
  const ban = row.status === 1
  const action = ban ? '封禁' : '解封'
  try {
    await ElMessageBox.confirm(`确定要${action}该用户吗？`, '确认', { type: 'warning' })
    const { data } = await http.put(`/api/v1/admin/users/${row.id}/status`, { ban })
    if (data.code === 0) {
      ElMessage.success('操作成功')
      loadUsers()
    } else {
      ElMessage.error(data.msg)
    }
  } catch (e) { /* cancelled */ }
}

onMounted(loadUsers)
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.page-header h2 { margin: 0; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
