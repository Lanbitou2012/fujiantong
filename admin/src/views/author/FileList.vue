<template>
  <div class="file-list">
    <div class="page-header">
      <h2>附件管理</h2>
      <el-upload
        :action="'/api/v1/author/files/upload'"
        :headers="{ Authorization: `Bearer ${userStore.token}` }"
        :on-success="handleUploadSuccess"
        :on-error="handleUploadError"
        :show-file-list="false"
      >
        <el-button type="primary" icon="Upload">上传附件</el-button>
      </el-upload>
    </div>

    <el-table :data="files" v-loading="loading" stripe>
      <el-table-column prop="original_name" label="文件名" min-width="200" />
      <el-table-column prop="ext" label="类型" width="80" />
      <el-table-column label="大小" width="100">
        <template #default="{ row }">{{ formatSize(row.size) }}</template>
      </el-table-column>
      <el-table-column prop="view_count" label="浏览" width="80" />
      <el-table-column prop="download_count" label="下载" width="80" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
            {{ row.status === 'active' ? '正常' : row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="handleCopy(row)">一键复制</el-button>
          <el-button size="small" @click="handleReplace(row)">替换</el-button>
          <el-button size="small" @click="handlePreview(row)">预览链接</el-button>
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

    <!-- 替换文件弹窗 -->
    <el-dialog v-model="replaceDialog" title="替换附件" width="400">
      <p>替换后原有下载链接不变，仅更新文件内容。</p>
      <el-upload
        :action="`/api/v1/author/files/${replaceFileId}`"
        :headers="{ Authorization: `Bearer ${userStore.token}` }"
        method="PUT"
        :on-success="handleReplaceSuccess"
        :show-file-list="false"
        name="file"
      >
        <el-button type="primary">选择新文件</el-button>
      </el-upload>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import http from '@/utils/http'

const userStore = useUserStore()
const files = ref([])
const loading = ref(false)
const page = ref(1)
const size = ref(20)
const total = ref(0)
const replaceDialog = ref(false)
const replaceFileId = ref(0)

async function loadFiles() {
  loading.value = true
  try {
    const { data } = await http.get('/api/v1/author/files', { params: { page: page.value, size: size.value } })
    if (data.code === 0) {
      files.value = data.data.list || []
      total.value = data.data.total
    }
  } finally {
    loading.value = false
  }
}

function changePage(p) {
  page.value = p
  loadFiles()
}

function handleUploadSuccess(res) {
  if (res.code === 0) {
    ElMessage.success('上传成功')
    loadFiles()
  } else {
    ElMessage.error(res.msg || '上传失败')
  }
}

function handleUploadError() {
  ElMessage.error('上传失败')
}

function handleReplace(row) {
  replaceFileId.value = row.id
  replaceDialog.value = true
}

function handleReplaceSuccess(res) {
  if (res.code === 0) {
    ElMessage.success('替换成功')
    replaceDialog.value = false
    loadFiles()
  } else {
    ElMessage.error(res.msg)
  }
}

function handlePreview(row) {
  const url = `${window.location.origin}/api/v1/c/file/${row.download_code}`
  navigator.clipboard?.writeText(url)
  ElMessage.success(`下载链接已复制: /f/${row.download_code}`)
}

// ─── 核心：一键复制富文本到剪贴板（三级回退） ───
async function handleCopy(row) {
  // 先获取分享卡片
  try {
    const { data } = await http.post(`/api/v1/author/files/${row.id}/share-card`)
    if (data.code !== 0) {
      ElMessage.error(data.msg || '生成分享卡片失败')
      return
    }
    const cards = data.data
    // 优先使用 weapp_text_link
    const card = cards.find(c => c.card_type === 'weapp_text_link') || cards[0]
    if (!card) {
      ElMessage.error('未生成分享卡片')
      return
    }
    await copyRichText(card.html)
    ElMessage.success('已复制到剪贴板，粘贴到公众号文章即可')
  } catch (e) {
    ElMessage.error('复制失败: ' + (e.message || ''))
  }
}

// 三级回退复制机制（复用自上代项目）
async function copyRichText(html) {
  // Level 1: ClipboardItem API (Chrome 66+)
  if (navigator.clipboard && window.ClipboardItem) {
    try {
      const blob = new Blob([html], { type: 'text/html' })
      const item = new ClipboardItem({ 'text/html': blob })
      await navigator.clipboard.write([item])
      return
    } catch (e) {
      // fallthrough
    }
  }

  // Level 2: contentEditable + document.execCommand
  const el = document.createElement('div')
  el.contentEditable = 'true'
  el.innerHTML = html
  el.style.position = 'fixed'
  el.style.left = '-9999px'
  document.body.appendChild(el)
  const range = document.createRange()
  range.selectNodeContents(el)
  const sel = window.getSelection()
  sel.removeAllRanges()
  sel.addRange(range)
  try {
    document.execCommand('copy')
  } catch (e) {
    // Level 3: 纯文本
    await navigator.clipboard?.writeText(html)
  }
  document.body.removeChild(el)
}

function formatSize(bytes) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let size = bytes
  while (size >= 1024 && i < units.length - 1) {
    size /= 1024
    i++
  }
  return size.toFixed(i > 0 ? 1 : 0) + ' ' + units[i]
}

onMounted(loadFiles)
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.page-header h2 {
  margin: 0;
}
.pagination {
  margin-top: 16px;
  justify-content: flex-end;
}
</style>
