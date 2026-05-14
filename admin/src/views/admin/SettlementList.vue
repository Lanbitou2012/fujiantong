<template>
  <div class="settlement-list">
    <h2>结算管理</h2>

    <el-table :data="list" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="promoter_id" label="推广员ID" width="100" />
      <el-table-column prop="settle_period" label="结算期" width="120" />
      <el-table-column label="平台分成(元)" width="120">
        <template #default="{ row }">{{ (row.platform_fee / 100).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="佣金(元)" width="100">
        <template #default="{ row }">{{ (row.commission / 100).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.status === 'pending'" size="small" type="primary" @click="approve(row)">审核通过</el-button>
          <el-button v-if="row.status === 'approved'" size="small" type="success" @click="markPaid(row)">标记已打款</el-button>
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

const list = ref([])
const loading = ref(false)
const page = ref(1)
const size = ref(20)
const total = ref(0)

async function loadList() {
  loading.value = true
  try {
    const { data } = await http.get('/api/v1/admin/settlements', { params: { page: page.value, size: size.value } })
    if (data.code === 0) {
      list.value = data.data.list || []
      total.value = data.data.total
    }
  } finally {
    loading.value = false
  }
}

function changePage(p) { page.value = p; loadList() }

function statusType(s) {
  return { pending: 'warning', approved: 'primary', paid: 'success', rejected: 'danger' }[s] || 'info'
}
function statusLabel(s) {
  return { pending: '待审核', approved: '待打款', paid: '已打款', rejected: '已驳回' }[s] || s
}

async function approve(row) {
  try {
    await ElMessageBox.confirm('确定审核通过该结算单？', '确认')
    const { data } = await http.post(`/api/v1/admin/settlements/${row.id}/approve`)
    data.code === 0 ? (ElMessage.success('已通过'), loadList()) : ElMessage.error(data.msg)
  } catch (e) { /* cancelled */ }
}

async function markPaid(row) {
  try {
    const { value } = await ElMessageBox.prompt('请输入转账凭证（可选）', '标记打款')
    const { data } = await http.post(`/api/v1/admin/settlements/${row.id}/mark-paid`, { transfer_proof: value || '' })
    data.code === 0 ? (ElMessage.success('已标记'), loadList()) : ElMessage.error(data.msg)
  } catch (e) { /* cancelled */ }
}

onMounted(loadList)
</script>

<style scoped>
h2 { margin: 0 0 16px 0; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
