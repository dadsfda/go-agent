<script setup>
import { computed, ref, onMounted } from 'vue'
import { request } from '../api.js'

const emit = defineEmits(['toast'])
const items = ref([])
const page = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const keyword = ref('')
const selected = ref(null)

async function load() {
  loading.value = true
  try {
    const data = await request(`/hr/applications?page=${page.value}&pageSize=${pageSize}`)
    items.value = data.items || data.applications || []
    total.value = data.total || 0
  } catch (e) {
    emit('toast', { message: e.message, type: 'error' })
  } finally {
    loading.value = false
  }
}

const totalPages = () => Math.max(1, Math.ceil(total.value / pageSize))
const filteredItems = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  if (!q) return items.value
  return items.value.filter(app => {
    const text = [
      app.profile?.name,
      app.profile?.phone,
      app.profile?.education,
      app.profile?.school,
      app.profile?.skills,
      app.job?.title,
      app.resume?.fileName
    ].join(' ').toLowerCase()
    return text.includes(q)
  })
})
const jobCount = computed(() => new Set(items.value.map(app => app.job?.id).filter(Boolean)).size)

function prev() { if (page.value > 1) { page.value--; load() } }
function next() { if (page.value < totalPages()) { page.value++; load() } }
function textPreview(text, length = 36) {
  const value = String(text || '').trim()
  if (!value) return '未填写'
  return value.length > length ? value.slice(0, length) + '...' : value
}
function skillTags(skills) {
  return String(skills || '')
    .split(/[,，、\s]+/)
    .map(x => x.trim())
    .filter(Boolean)
    .slice(0, 5)
}
function openDetail(app) {
  selected.value = app
}
function closeDetail() {
  selected.value = null
}

onMounted(load)
</script>

<template>
  <div>
    <div class="page-header">
      <h1>候选人台账</h1>
      <p>查看投递候选人的档案、技能和云端简历</p>
    </div>

    <div class="card">
      <div class="ledger-toolbar">
        <div class="ledger-stats">
          <div>
            <strong>{{ total }}</strong>
            <span>投递记录</span>
          </div>
          <div>
            <strong>{{ jobCount }}</strong>
            <span>相关岗位</span>
          </div>
        </div>
        <div class="ledger-actions">
          <input v-model="keyword" placeholder="搜索姓名、电话、岗位或技能" />
          <button class="btn-secondary" :disabled="loading" @click="load">{{ loading ? '刷新中...' : '刷新' }}</button>
        </div>
      </div>

      <div v-if="loading && !items.length" class="empty">加载中...</div>
      <div v-else-if="!items.length" class="empty">暂无投递记录</div>
      <div v-else class="table-wrap">
        <table class="ledger-table">
          <thead>
            <tr>
              <th>候选人</th>
              <th>投递岗位</th>
              <th>联系方式</th>
              <th>教育背景</th>
              <th>经历预览</th>
              <th>技能标签</th>
              <th>简历</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="app in filteredItems" :key="app.id">
              <td>
                <div class="candidate-name">{{ app.profile?.name || '未命名' }}</div>
                <div class="muted">投递于 {{ app.createdAt || '-' }}</div>
              </td>
              <td>
                <span class="job-pill">{{ app.job?.title || '未知岗位' }}</span>
              </td>
              <td>{{ app.profile?.phone || '-' }}</td>
              <td>
                <div>{{ app.profile?.education || '-' }}</div>
                <div class="muted">{{ app.profile?.school || '-' }}</div>
              </td>
              <td class="preview-cell">{{ textPreview(app.profile?.experience) }}</td>
              <td>
                <div class="skill-list">
                  <span v-for="tag in skillTags(app.profile?.skills)" :key="tag" class="skill-tag">{{ tag }}</span>
                  <span v-if="!skillTags(app.profile?.skills).length" class="muted">未填写</span>
                </div>
              </td>
              <td>
                <a :href="app.resume?.signedUrl" target="_blank" class="resume-link">
                  {{ textPreview(app.resume?.fileName, 12) }}
                </a>
              </td>
              <td>
                <button class="btn-secondary btn-sm" @click="openDetail(app)">查看详情</button>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="!filteredItems.length" class="empty compact">没有匹配的候选人</div>
      </div>

      <div v-if="total > pageSize" class="pager">
        <button class="btn-secondary btn-sm" :disabled="page <= 1" @click="prev">上一页</button>
        <span>第 {{ page }} / {{ totalPages() }} 页，共 {{ total }} 条</span>
        <button class="btn-secondary btn-sm" :disabled="page >= totalPages()" @click="next">下一页</button>
      </div>
    </div>

    <div v-if="selected" class="detail-mask" @click.self="closeDetail">
      <div class="detail-panel">
        <div class="detail-head">
          <div>
            <h2>{{ selected.profile?.name || '候选人详情' }}</h2>
            <p>{{ selected.job?.title }} · {{ selected.profile?.phone }}</p>
          </div>
          <button class="btn-ghost" @click="closeDetail">关闭</button>
        </div>
        <div class="detail-grid">
          <section>
            <span>教育背景</span>
            <p>{{ selected.profile?.education || '-' }} / {{ selected.profile?.school || '-' }}</p>
          </section>
          <section>
            <span>云端简历</span>
            <p>
              <a :href="selected.resume?.signedUrl" target="_blank" class="resume-link">
                {{ selected.resume?.fileName || '查看简历' }}
              </a>
            </p>
          </section>
          <section class="wide">
            <span>工作 / 项目经历</span>
            <p>{{ selected.profile?.experience || '未填写' }}</p>
          </section>
          <section class="wide">
            <span>核心技能</span>
            <p>{{ selected.profile?.skills || '未填写' }}</p>
          </section>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ledger-toolbar {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
  margin-bottom: 1rem;
}
.ledger-stats {
  display: flex;
  gap: 0.75rem;
}
.ledger-stats > div {
  min-width: 92px;
  min-height: 76px;
  padding: 0.85rem 0.95rem;
  border-radius: 12px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-sm);
  transition: all var(--transition);
}
.ledger-stats > div:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow);
}
.ledger-stats strong {
  display: block;
  font-size: 1.15rem;
}
.ledger-stats span,
.muted {
  color: var(--text-muted);
  font-size: 0.78rem;
}
.ledger-actions {
  display: grid;
  grid-template-columns: minmax(220px, 320px) auto;
  gap: 0.75rem;
}
.ledger-table {
  table-layout: fixed;
}
.ledger-table th:nth-child(1) { width: 13%; }
.ledger-table th:nth-child(2) { width: 13%; }
.ledger-table th:nth-child(3) { width: 11%; }
.ledger-table th:nth-child(4) { width: 14%; }
.ledger-table th:nth-child(5) { width: 18%; }
.ledger-table th:nth-child(6) { width: 18%; }
.ledger-table th:nth-child(7) { width: 9%; }
.ledger-table th:nth-child(8) { width: 8%; }
.candidate-name {
  font-weight: 700;
}
.job-pill {
  display: inline-flex;
  max-width: 100%;
  padding: 0.25rem 0.55rem;
  border-radius: 999px;
  background: #EAF2FF;
  color: var(--primary);
  font-size: 0.8rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.preview-cell {
  color: var(--text-secondary);
}
.skill-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}
.skill-tag {
  display: inline-flex;
  max-width: 120px;
  padding: 0.28rem 0.65rem;
  border-radius: 999px;
  background: var(--primary-light);
  color: var(--primary);
  border: 1px solid rgba(31, 115, 241, 0.1);
  font-size: 0.76rem;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.resume-link {
  color: var(--primary);
  font-size: 0.85rem;
  font-weight: 600;
  text-decoration: none;
}
.resume-link:hover {
  text-decoration: underline;
}
.compact {
  padding: 1.5rem 1rem;
}
.detail-mask {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: rgba(15, 23, 42, 0.35);
  display: flex;
  justify-content: flex-end;
}
.detail-panel {
  width: min(620px, 100%);
  height: 100%;
  overflow-y: auto;
  background: var(--bg-card);
  box-shadow: -16px 0 48px rgba(15, 23, 42, 0.15);
  padding: 1.5rem;
  animation: panelIn 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}
@keyframes panelIn {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}
.detail-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border);
  margin-bottom: 1rem;
}
.detail-head h2 {
  font-size: 1.25rem;
}
.detail-head p {
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin-top: 0.2rem;
}
.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
.detail-grid section {
  padding: 1rem;
  border-radius: 12px;
  background: #F8FAFC;
  border: 1px solid var(--border-light);
}
.detail-grid section.wide {
  grid-column: 1 / -1;
}
.detail-grid span {
  color: var(--text-muted);
  font-size: 0.78rem;
  font-weight: 700;
}
.detail-grid p {
  margin-top: 0.5rem;
  line-height: 1.75;
  white-space: pre-wrap;
}
@media (max-width: 900px) {
  .ledger-toolbar {
    align-items: stretch;
    flex-direction: column;
  }
  .ledger-actions {
    grid-template-columns: 1fr;
  }
}
</style>
