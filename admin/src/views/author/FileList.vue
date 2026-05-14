<template>
  <div class="author-files" @dragenter.prevent="onDragEnter" @dragover.prevent @drop.prevent="onDrop">
    <!-- 拖拽遮罩 -->
    <div v-if="dragActive" class="drop-mask" @dragleave="onDragLeave">
      <div class="drop-mask__box">
        <div class="drop-mask__icon">📥</div>
        <div class="drop-mask__title">松开即可上传</div>
        <div class="drop-mask__sub">支持多文件，单个不超过 20MB</div>
      </div>
    </div>

    <!-- 数据概览（前端聚合） -->
    <div class="stats-row">
      <div class="stat-card stat-card--files">
        <div class="stat-card__header">
          <div class="stat-card__icon">📁</div>
          <div class="stat-card__label">附件总数</div>
        </div>
        <div class="stat-card__value">{{ total }}</div>
      </div>
      <div class="stat-card stat-card--views">
        <div class="stat-card__header">
          <div class="stat-card__icon">👁</div>
          <div class="stat-card__label">本页累计浏览</div>
        </div>
        <div class="stat-card__value">{{ totalViews }}</div>
      </div>
      <div class="stat-card stat-card--downloads">
        <div class="stat-card__header">
          <div class="stat-card__icon">⬇️</div>
          <div class="stat-card__label">本页累计下载</div>
        </div>
        <div class="stat-card__value">{{ totalDownloads }}</div>
      </div>
      <div class="stat-card stat-card--recent">
        <div class="stat-card__header">
          <div class="stat-card__icon">🕐</div>
          <div class="stat-card__label">本页文件</div>
        </div>
        <div class="stat-card__value">{{ files.length }}</div>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar__left">
        <h2 class="toolbar__title">附件管理</h2>
      </div>
      <div class="toolbar__right">
        <el-tooltip content="或者直接把文件拖到当前页面任意位置">
          <el-button type="primary" size="large" @click="onPickFile">
            <span style="font-size: 18px; margin-right: 6px;">📤</span>
            <span>上传文件（支持多选）</span>
          </el-button>
        </el-tooltip>
        <input ref="fileInput" type="file" multiple style="display: none" @change="onFilesChosen" />
      </div>
    </div>

    <!-- 上传队列 -->
    <div v-if="uploadQueue.length > 0" class="upload-queue">
      <div class="upload-queue__title">
        ⬆️ 上传队列 ({{ uploadQueueDone }}/{{ uploadQueue.length }})
      </div>
      <div v-for="u in uploadQueue" :key="u.id" class="upload-item">
        <div class="upload-item__name">📄 {{ u.name }}</div>
        <el-progress
          :percentage="u.progress"
          :status="u.status === 'error' ? 'exception' : (u.status === 'done' ? 'success' : '')"
          style="flex: 1; margin: 0 12px;"
        />
        <span class="upload-item__status">
          <template v-if="u.status === 'uploading'">{{ u.progress }}%</template>
          <template v-else-if="u.status === 'done'">✓ 完成</template>
          <template v-else-if="u.status === 'error'">✗ {{ u.error }}</template>
        </span>
      </div>
    </div>

    <!-- 文件列表 -->
    <div class="file-list" v-loading="loading">
      <el-empty v-if="!loading && files.length === 0" description="还没有上传过附件">
        <el-button type="primary" @click="onPickFile">📤 上传第一个文件</el-button>
      </el-empty>

      <div v-for="f in files" :key="f.id" class="file-row">
        <div class="file-row__icon" :style="{ background: iconColor(f.ext) }">
          {{ iconLabel(f.ext) }}
        </div>
        <div class="file-row__main">
          <div class="file-row__name">
            <el-tooltip :content="f.original_name" placement="top">
              <span>{{ f.original_name }}</span>
            </el-tooltip>
          </div>
          <div class="file-row__meta">
            <span>{{ formatSize(f.size) }}</span>
            <span class="sep">·</span>
            <span>👁 {{ f.view_count || 0 }}</span>
            <span class="sep">·</span>
            <span>⬇ {{ f.download_count || 0 }}</span>
            <span class="sep">·</span>
            <el-tag :type="f.status === 'active' ? 'success' : 'danger'" size="small">
              {{ f.status === 'active' ? '正常' : f.status }}
            </el-tag>
          </div>
        </div>
        <div class="file-row__actions">
          <el-button type="primary" @click="copyShareCard(f)">
            📋 一键粘贴
          </el-button>
          <el-tooltip content="复制 /f/ 兜底链接，建议作为文章末尾备用下载链接">
            <el-button @click="copyH5BackupLink(f)">
              🔗 H5备份
            </el-button>
          </el-tooltip>
          <el-button @click="handleReplace(f)">🔄 替换</el-button>
        </div>
      </div>

      <div v-if="totalPages > 1" class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          :page-size="size"
          :total="total"
          layout="prev, pager, next, total"
          @current-change="loadFiles"
        />
      </div>
    </div>

    <!-- 替换文件 input -->
    <input ref="replaceInput" type="file" style="display: none" @change="onReplaceChosen" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import http from '@/utils/http'

const userStore = useUserStore()
const files = ref([])
const loading = ref(false)
const page = ref(1)
const size = ref(20)
const total = ref(0)
const totalPages = computed(() => Math.ceil(total.value / size.value))
const totalViews = computed(() => files.value.reduce((s, f) => s + (f.view_count || 0), 0))
const totalDownloads = computed(() => files.value.reduce((s, f) => s + (f.download_count || 0), 0))

const fileInput = ref(null)
const replaceInput = ref(null)
const replacingFile = ref(null)

const dragActive = ref(false)
let dragDepth = 0

const uploadQueue = ref([])
const uploadQueueDone = computed(() => uploadQueue.value.filter(u => u.status !== 'uploading').length)

const ALLOWED = ['xlsx', 'xls', 'csv', 'doc', 'docx', 'pdf', 'ppt', 'pptx', 'txt', 'zip', 'rar']
const MAX_SIZE = 20 * 1024 * 1024

// ─── 加载文件列表 ───
async function loadFiles() {
  loading.value = true
  try {
    const { data } = await http.get('/api/v1/author/files', {
      params: { page: page.value, size: size.value },
    })
    if (data.code === 0) {
      files.value = data.data.list || []
      total.value = data.data.total || 0
    }
  } finally {
    loading.value = false
  }
}

// ─── 上传 ───
function onPickFile() {
  fileInput.value && fileInput.value.click()
}

function onFilesChosen(e) {
  const list = Array.from(e.target.files || [])
  e.target.value = ''
  if (list.length) enqueueUploads(list)
}

function onDragEnter(e) {
  if (e.dataTransfer && e.dataTransfer.types && e.dataTransfer.types.includes('Files')) {
    dragDepth++
    dragActive.value = true
  }
}
function onDragLeave() {
  dragDepth--
  if (dragDepth <= 0) {
    dragDepth = 0
    dragActive.value = false
  }
}
function onDrop(e) {
  dragDepth = 0
  dragActive.value = false
  const list = Array.from(e.dataTransfer?.files || [])
  if (list.length) enqueueUploads(list)
}

function enqueueUploads(list) {
  const valid = []
  for (const f of list) {
    const ext = (f.name.split('.').pop() || '').toLowerCase()
    if (!ALLOWED.includes(ext)) {
      ElMessage.warning(`不支持的文件类型：${f.name}`)
      continue
    }
    if (f.size > MAX_SIZE) {
      ElMessage.warning(`${f.name} 超过 20MB 上限`)
      continue
    }
    valid.push(f)
  }
  if (!valid.length) return

  valid.forEach(file => {
    const task = {
      id: Date.now() + '_' + Math.random().toString(36).slice(2),
      name: file.name,
      file,
      status: 'uploading',
      progress: 0,
      error: '',
    }
    uploadQueue.value.push(task)
    startUpload(task)
  })
}

async function startUpload(task) {
  const fd = new FormData()
  fd.append('file', task.file)
  try {
    await http.post('/api/v1/author/files/upload', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (e.total) task.progress = Math.round((e.loaded / e.total) * 100)
      },
    })
    task.status = 'done'
    task.progress = 100
    if (uploadQueue.value.every(t => t.status !== 'uploading')) {
      setTimeout(() => {
        uploadQueue.value = []
        page.value = 1
        loadFiles()
      }, 800)
    }
  } catch (e) {
    task.status = 'error'
    task.error = e.response?.data?.msg || e.message || '上传失败'
  }
}

// ─── 一键复制（三级回退） ───
async function copyShareCard(f) {
  try {
    const { data } = await http.post(`/api/v1/author/files/${f.id}/share-card`)
    if (data.code !== 0) {
      ElMessage.error(data.msg || '生成分享卡片失败')
      return
    }
    const cards = data.data || []
    const card = cards.find(c => c.card_type === 'weapp_text_link') || cards[0]
    if (!card) {
      ElMessage.error('未生成分享卡片')
      return
    }
    const ok = await copyHtml(card.html, f.original_name || '点击下载附件')
    if (ok) {
      ElMessage.success({
        message: '已复制！切到公众号编辑器（mp.weixin.qq.com）正文里 Ctrl+V 粘贴即可',
        duration: 3500,
      })
    } else {
      ElMessage.error('复制失败，请手动复制')
    }
  } catch (e) {
    ElMessage.error('复制失败: ' + (e.message || ''))
  }
}

// 三级回退复制：ClipboardItem (text/html + text/plain) → contentEditable + execCommand → 纯文本
async function copyHtml(html, plainFallback) {
  const plain = plainFallback || html
  // Level 1: 异步 ClipboardItem API
  try {
    if (window.ClipboardItem && navigator.clipboard && navigator.clipboard.write) {
      const item = new ClipboardItem({
        'text/html': new Blob([html], { type: 'text/html' }),
        'text/plain': new Blob([plain], { type: 'text/plain' }),
      })
      await navigator.clipboard.write([item])
      return true
    }
  } catch (e) { /* fallthrough */ }
  // Level 2: contentEditable + execCommand
  try {
    const div = document.createElement('div')
    div.contentEditable = 'true'
    div.innerHTML = html
    div.style.position = 'fixed'
    div.style.left = '-9999px'
    div.style.top = '0'
    document.body.appendChild(div)
    const range = document.createRange()
    range.selectNodeContents(div)
    const sel = window.getSelection()
    sel.removeAllRanges()
    sel.addRange(range)
    const ok = document.execCommand('copy')
    sel.removeAllRanges()
    document.body.removeChild(div)
    if (ok) return true
  } catch (e) { /* fallthrough */ }
  // Level 3: 纯文本
  try {
    await navigator.clipboard.writeText(plain)
    return true
  } catch (e) {
    return false
  }
}

async function copyH5BackupLink(f) {
  const url = `${window.location.origin}/f/${f.download_code || f.id}`
  try {
    await navigator.clipboard.writeText(url)
    ElMessageBox.alert(
      `H5 备份链接已复制：<br><code style="word-break:break-all">${url}</code><br><br>建议在公众号文章末尾小字附上：<br><strong>备用下载链接：</strong>${url}`,
      'H5 备份链接已复制',
      { dangerouslyUseHTMLString: true, confirmButtonText: '知道了' }
    )
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}

// ─── 替换文件 ───
function handleReplace(f) {
  replacingFile.value = f
  replaceInput.value && replaceInput.value.click()
}

async function onReplaceChosen(e) {
  const file = e.target.files && e.target.files[0]
  e.target.value = ''
  if (!file || !replacingFile.value) return

  const ext = (file.name.split('.').pop() || '').toLowerCase()
  if (!ALLOWED.includes(ext)) {
    ElMessage.warning('不支持的文件类型')
    return
  }
  if (file.size > MAX_SIZE) {
    ElMessage.warning('文件超过 20MB 上限')
    return
  }

  const id = replacingFile.value.id
  replacingFile.value = null
  const fd = new FormData()
  fd.append('file', file)
  const loadingMsg = ElMessage({ message: '替换中...', duration: 0, type: 'info' })
  try {
    const { data } = await http.put(`/api/v1/author/files/${id}`, fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    loadingMsg.close()
    if (data.code === 0) {
      ElMessage.success('替换成功，下载链接不变')
      loadFiles()
    } else {
      ElMessage.error(data.msg || '替换失败')
    }
  } catch (e) {
    loadingMsg.close()
    ElMessage.error(e.response?.data?.msg || '替换失败')
  }
}

// ─── 工具方法 ───
const ICON_COLORS = {
  xlsx: '#10B981', xls: '#10B981', csv: '#10B981',
  doc: '#3B82F6', docx: '#3B82F6',
  pdf: '#EF4444',
  ppt: '#F59E0B', pptx: '#F59E0B',
  txt: '#64748B', zip: '#A855F7', rar: '#A855F7',
}
function iconColor(t) {
  return ICON_COLORS[(t || '').toLowerCase()] || '#94A3B8'
}
function iconLabel(t) {
  const ext = (t || '').toLowerCase()
  const map = { xlsx: 'XLS', xls: 'XLS', docx: 'DOC', pptx: 'PPT' }
  return (map[ext] || ext || 'FILE').toUpperCase()
}
function formatSize(n) {
  if (!n) return '0 B'
  if (n < 1024) return n + ' B'
  if (n < 1024 * 1024) return (n / 1024).toFixed(1) + ' KB'
  return (n / 1024 / 1024).toFixed(2) + ' MB'
}

onMounted(loadFiles)
</script>

<style scoped>
.author-files {
  position: relative;
  min-height: calc(100vh - 112px);
}

/* 拖拽遮罩 */
.drop-mask {
  position: fixed;
  inset: 0;
  background: rgba(21, 128, 61, 0.92);
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
  animation: fadeIn 0.15s ease;
}
.drop-mask__box {
  background: #fff;
  padding: 64px 80px;
  border-radius: 28px;
  text-align: center;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}
.drop-mask__icon { font-size: 72px; margin-bottom: 16px; }
.drop-mask__title { font-size: 24px; font-weight: 700; color: #0f172a; margin-bottom: 8px; }
.drop-mask__sub { font-size: 14px; color: #64748b; }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }

/* 统计卡片 */
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}
.stat-card {
  background: #fff;
  padding: 20px 24px;
  border-radius: 16px;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.05);
  border: 1px solid #f1f5f9;
  transition: all 0.2s;
}
.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(15, 23, 42, 0.08);
}
.stat-card__header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}
.stat-card__icon {
  width: 36px; height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}
.stat-card--files .stat-card__icon { background: #ecfdf5; }
.stat-card--views .stat-card__icon { background: #eff6ff; }
.stat-card--downloads .stat-card__icon { background: #fef3c7; }
.stat-card--recent .stat-card__icon { background: #f0f4ff; }
.stat-card__label {
  font-size: 13px;
  color: #64748b;
  font-weight: 500;
}
.stat-card__value {
  font-size: 30px;
  font-weight: 800;
  letter-spacing: -1px;
}
.stat-card--files .stat-card__value { color: #15803d; }
.stat-card--views .stat-card__value { color: #2563eb; }
.stat-card--downloads .stat-card__value { color: #d97706; }
.stat-card--recent .stat-card__value { color: #7c3aed; }

/* 工具栏 */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  background: #fff;
  padding: 14px 20px;
  border-radius: 14px;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.05);
  border: 1px solid #f1f5f9;
}
.toolbar__title { margin: 0; font-size: 18px; font-weight: 600; }
.toolbar__right { display: flex; gap: 12px; }

/* 上传队列 */
.upload-queue {
  background: #fff;
  padding: 16px 20px;
  border-radius: 14px;
  margin-bottom: 20px;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.05);
  border: 1px solid #f1f5f9;
}
.upload-queue__title {
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 12px;
}
.upload-item {
  display: flex;
  align-items: center;
  padding: 8px 0;
  border-top: 1px solid #f1f5f9;
}
.upload-item:first-of-type { border-top: none; }
.upload-item__name {
  width: 280px;
  font-size: 13px;
  color: #334155;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.upload-item__status {
  font-size: 12px;
  color: #64748b;
  min-width: 80px;
  text-align: right;
  font-weight: 600;
}

/* 文件卡片列表 */
.file-list {
  background: #fff;
  border-radius: 14px;
  padding: 4px 0;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.05);
  border: 1px solid #f1f5f9;
  min-height: 280px;
}
.file-row {
  display: flex;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #f8fafc;
  transition: background 0.15s;
}
.file-row:last-child { border-bottom: none; }
.file-row:hover { background: #f8fafc; }

.file-row__icon {
  width: 48px; height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.5px;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
}
.file-row__main {
  flex: 1;
  margin-left: 18px;
  min-width: 0;
  overflow: hidden;
}
.file-row__name {
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 560px;
}
.file-row__meta {
  font-size: 12px;
  color: #94a3b8;
  display: flex;
  align-items: center;
  gap: 4px;
}
.file-row__meta .sep {
  margin: 0 6px;
  color: #e2e8f0;
}
.file-row__actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}
.pagination-wrap {
  padding: 18px;
  display: flex;
  justify-content: center;
}
</style>
