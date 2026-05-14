<template>
  <div class="my-revenue">
    <h2>我的收益</h2>
    <el-alert
      title="作者收益说明"
      type="info"
      :closable="false"
      description="您作为附件作者，可获得自有小程序流量主毛收入的 72%。平台留 28%（含腾讯返点 50% + 平台运营 22% + 推广员佣金 30%）。每日数据 T+1 同步，月初生成结算单。"
      class="info-alert"
    />

    <el-row :gutter="16" class="stat-row">
      <el-col :span="8">
        <el-card shadow="hover">
          <el-statistic title="本月预估收益(元)" :value="(stats.month_estimated / 100).toFixed(2)" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <el-statistic title="累计已结算(元)" :value="(stats.total_settled / 100).toFixed(2)" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <el-statistic title="附件下载总数" :value="stats.total_downloads" />
        </el-card>
      </el-col>
    </el-row>

    <el-card class="settlement-card">
      <template #header>结算明细</template>
      <el-table :data="settlements" stripe>
        <el-table-column prop="settle_period" label="期次" width="120" />
        <el-table-column label="平台总收入(元)" width="140">
          <template #default="{ row }">{{ (row.gross_income / 100).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="您的 72%(元)" width="140">
          <template #default="{ row }">{{ (row.author_share / 100).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'paid' ? 'success' : 'warning'" size="small">
              {{ row.status === 'paid' ? '已到账' : '处理中' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="paid_at" label="到账时间" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import http from '@/utils/http'

const stats = ref({ month_estimated: 0, total_settled: 0, total_downloads: 0 })
const settlements = ref([])

onMounted(async () => {
  try {
    const { data } = await http.get('/api/v1/author/revenue')
    if (data.code === 0) {
      stats.value = data.data.stats || stats.value
      settlements.value = data.data.settlements || []
    }
  } catch (e) { /* ignore */ }
})
</script>

<style scoped>
h2 { margin: 0 0 16px 0; }
.info-alert { margin-bottom: 20px; }
.stat-row { margin-bottom: 20px; }
.settlement-card { margin-top: 16px; }
</style>
