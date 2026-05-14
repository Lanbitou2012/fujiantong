<template>
  <div class="deployment-list">
    <h2>部署管理</h2>

    <el-table :data="list" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="appid" label="AppID" width="180" />
      <el-table-column prop="user_version" label="版本" width="120" />
      <el-table-column prop="status" label="状态" width="160">
        <template #default="{ row }">
          <el-tag :type="deployStatusType(row.status)" size="small">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="fail_reason" label="失败原因" min-width="200" show-overflow-tooltip />
      <el-table-column prop="created_at" label="创建时间" width="170" />
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
import http from '@/utils/http'

const list = ref([])
const loading = ref(false)
const page = ref(1)
const size = ref(20)
const total = ref(0)

async function loadList() {
  loading.value = true
  try {
    const { data } = await http.get('/api/v1/admin/deployments', { params: { page: page.value, size: size.value } })
    if (data.code === 0) {
      list.value = data.data.list || []
      total.value = data.data.total
    }
  } finally {
    loading.value = false
  }
}

function changePage(p) { page.value = p; loadList() }

function deployStatusType(s) {
  if (s === 'live') return 'success'
  if (s.includes('fail')) return 'danger'
  if (s.includes('waiting')) return 'warning'
  return 'info'
}

onMounted(loadList)
</script>

<style scoped>
h2 { margin: 0 0 16px 0; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
