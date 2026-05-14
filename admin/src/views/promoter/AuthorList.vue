<template>
  <div class="author-list">
    <h2>我的作者</h2>
    <el-table :data="list" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="nickname" label="昵称" width="140" />
      <el-table-column prop="bound_appid" label="AppID" width="200" />
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ row.status === 1 ? '正常' : '封禁' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="注册时间" width="170" />
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
    const { data } = await http.get('/api/v1/promoter/authors', { params: { page: page.value, size: size.value } })
    if (data.code === 0) {
      list.value = data.data.list || []
      total.value = data.data.total
    }
  } finally {
    loading.value = false
  }
}

function changePage(p) { page.value = p; loadList() }

onMounted(loadList)
</script>

<style scoped>
h2 { margin: 0 0 16px 0; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
