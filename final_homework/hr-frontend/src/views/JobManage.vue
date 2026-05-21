<script setup>
import { computed, ref, onMounted } from 'vue'
import { request } from '../api.js'
import { auth } from '../stores/auth.js'

const emit = defineEmits(['toast'])
const jobs = ref([])
const form = ref({ title: '', description: '' })
const editing = ref(null)
const loading = ref(false)
const keyword = ref('')
const statusFilter = ref('all')
const currentUserId = computed(() => Number(auth.user?.id || 0))

async function loadJobs() {
  try {
    const data = await request('/hr/jobs')
    jobs.value = data.jobs || []
  } catch (e) {
    emit('toast', { message: e.message, type: 'error' })
  }
}

async function handleSubmit() {
  if (!form.value.title.trim() || !form.value.description.trim()) {
    return emit('toast', { message: '请填写岗位名称和描述', type: 'error' })
  }
  loading.value = true
  try {
    if (editing.value) {
      await request(`/hr/jobs/${editing.value.id}`, {
        method: 'PUT',
        body: JSON.stringify({ ...form.value, status: editing.value.status })
      })
      emit('toast', { message: '岗位已更新', type: 'success' })
    } else {
      await request('/hr/jobs', {
        method: 'POST',
        body: JSON.stringify(form.value)
      })
      emit('toast', { message: '岗位已发布', type: 'success' })
    }
    form.value = { title: '', description: '' }
    editing.value = null
    await loadJobs()
  } catch (e) {
    emit('toast', { message: e.message, type: 'error' })
  } finally {
    loading.value = false
  }
}

function startEdit(job) {
  if (!isOwnJob(job)) {
    return emit('toast', { message: '只能编辑自己创建的岗位', type: 'error' })
  }
  editing.value = job
  form.value = { title: job.title, description: job.description }
}

function cancelEdit() {
  editing.value = null
  form.value = { title: '', description: '' }
}

async function closeJob(job) {
  if (!isOwnJob(job)) {
    return emit('toast', { message: '只能下架自己创建的岗位', type: 'error' })
  }
  try {
    await request(`/hr/jobs/${job.id}`, {
      method: 'PUT',
      body: JSON.stringify({ title: job.title, description: job.description, status: 'closed' })
    })
    emit('toast', { message: '岗位已下架', type: 'success' })
    await loadJobs()
  } catch (e) {
    emit('toast', { message: e.message, type: 'error' })
  }
}

async function republishJob(job) {
  if (!isOwnJob(job)) {
    return emit('toast', { message: '只能重新上架自己创建的岗位', type: 'error' })
  }
  try {
    await request(`/hr/jobs/${job.id}`, {
      method: 'PUT',
      body: JSON.stringify({ title: job.title, description: job.description, status: 'open' })
    })
    emit('toast', { message: '岗位已重新上架', type: 'success' })
    await loadJobs()
  } catch (e) {
    emit('toast', { message: e.message, type: 'error' })
  }
}

async function deleteJob(job) {
  if (!isOwnJob(job)) {
    return emit('toast', { message: '只能删除自己创建的岗位', type: 'error' })
  }
  if (!confirm(`确定要删除岗位「${job.title}」吗？删除后不可恢复。`)) return
  try {
    await request(`/hr/jobs/${job.id}`, { method: 'DELETE' })
    emit('toast', { message: '岗位已删除', type: 'success' })
    await loadJobs()
  } catch (e) {
    emit('toast', { message: e.message, type: 'error' })
  }
}

const totalJobs = computed(() => jobs.value.length)
const openJobs = computed(() => jobs.value.filter(job => job.status === 'open').length)
const closedJobs = computed(() => jobs.value.filter(job => job.status === 'closed').length)
const filteredJobs = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  return jobs.value.filter(job => {
    const matchesStatus = statusFilter.value === 'all' || job.status === statusFilter.value
    const matchesKeyword = !q || [job.title, job.description, job.id].join(' ').toLowerCase().includes(q)
    return matchesStatus && matchesKeyword
  })
})
const descriptionLength = computed(() => form.value.description.trim().length)

function isOwnJob(job) {
  return Number(job.ownerHrId) === currentUserId.value
}

function ownerText(job) {
  return isOwnJob(job) ? '我发布的' : `HR #${job.ownerHrId}`
}

function statusText(job) {
  return job.status === 'open' ? '候选人可见，可继续投递' : '岗位已隐藏，重新上架后恢复投递'
}

function textPreview(text, length = 52) {
  const value = String(text || '').trim()
  if (!value) return '暂无岗位描述'
  return value.length > length ? value.slice(0, length) + '...' : value
}

onMounted(loadJobs)
</script>

<template>
  <div>
    <div class="page-header">
      <h1>岗位管理</h1>
      <p>发布招聘需求、查看状态，并持续维护候选人可见的岗位信息</p>
    </div>

    <div class="job-layout">
      <div class="card publish-card">
        <div class="section-head">
          <div>
            <h3>{{ editing ? '编辑岗位' : '发布新岗位' }}</h3>
            <p>{{ editing ? '更新标题、职责与候选人看到的岗位说明' : '先补齐核心信息，再对外开放投递入口' }}</p>
          </div>
          <span class="panel-tag">{{ editing ? '编辑中' : '待发布' }}</span>
        </div>
      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label>岗位名称</label>
          <input v-model="form.title" placeholder="如：高级 Go 开发工程师" />
        </div>
        <div class="form-group">
          <div class="field-head">
            <label>岗位描述</label>
            <span>{{ descriptionLength }} 字</span>
          </div>
          <textarea v-model="form.description" placeholder="岗位职责、任职要求、加分项等，建议分点描述" rows="5"></textarea>
        </div>
        <div class="form-actions">
          <button type="submit" class="btn-primary" :disabled="loading">
            {{ loading ? '提交中...' : editing ? '保存修改' : '发布岗位' }}
          </button>
          <button v-if="editing" type="button" class="btn-secondary" @click="cancelEdit">取消编辑</button>
        </div>
      </form>
      </div>

      <div class="card summary-card">
        <div class="section-head compact-head">
          <div>
            <h3>岗位概览</h3>
            <p>先看总量，再处理需要维护的岗位</p>
          </div>
        </div>
        <div class="summary-grid">
          <div>
            <strong>{{ totalJobs }}</strong>
            <span>全部岗位</span>
          </div>
          <div>
            <strong>{{ openJobs }}</strong>
            <span>招聘中</span>
          </div>
          <div>
            <strong>{{ closedJobs }}</strong>
            <span>已下架</span>
          </div>
        </div>
        <div class="summary-note">
          <strong>管理提示</strong>
          <p>全部岗位可查看；只有自己发布的岗位可以编辑、下架或删除。</p>
        </div>
      </div>
    </div>

    <div class="card ledger-card">
      <div class="ledger-header">
        <div>
          <h3>全部岗位</h3>
          <p>可查看全部岗位，维护动作仅限本人发布的岗位</p>
        </div>
        <div class="ledger-actions">
          <input v-model="keyword" placeholder="搜索岗位名称或描述" />
          <select v-model="statusFilter">
            <option value="all">全部状态</option>
            <option value="open">招聘中</option>
            <option value="closed">已下架</option>
          </select>
          <button class="btn-secondary" @click="loadJobs">刷新</button>
        </div>
      </div>

      <div v-if="!jobs.length" class="empty jobs-empty">
        <strong>还没有岗位</strong>
        <span>先发布第一个岗位，候选人才能在前台浏览和投递。</span>
      </div>
      <div v-else class="table-wrap">
        <table class="jobs-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>岗位名称</th>
              <th>归属</th>
              <th>描述</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="job in filteredJobs" :key="job.id">
              <td>{{ job.id }}</td>
              <td>
                <div class="job-title">{{ job.title }}</div>
                <div class="muted">候选人可见字段</div>
              </td>
              <td>
                <span class="owner-tag" :class="{ external: !isOwnJob(job) }">{{ ownerText(job) }}</span>
              </td>
              <td class="description-cell">{{ textPreview(job.description) }}</td>
              <td>
                <div class="status-stack">
                  <span class="badge" :class="job.status === 'open' ? 'badge-open' : 'badge-closed'">
                    {{ job.status === 'open' ? '招聘中' : '已下架' }}
                  </span>
                  <span>{{ statusText(job) }}</span>
                </div>
              </td>
              <td>
                <div v-if="isOwnJob(job)" class="action-group">
                  <button class="btn-secondary btn-sm" @click="startEdit(job)">编辑</button>
                  <button v-if="job.status === 'open'" class="btn-danger btn-sm" @click="closeJob(job)">下架</button>
                  <button v-if="job.status === 'closed'" class="btn-primary btn-sm" @click="republishJob(job)">重新上架</button>
                  <button class="btn-outline-danger btn-sm" @click="deleteJob(job)">删除</button>
                </div>
                <span v-else class="readonly-text">仅可查看</span>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="!filteredJobs.length" class="empty compact">
          没有匹配当前筛选条件的岗位
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.job-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.6fr) minmax(280px, 0.9fr);
  gap: 1.5rem;
  margin-bottom: 1.5rem;
}
.publish-card,
.summary-card,
.ledger-card {
  border: 1px solid var(--border-light);
}
.publish-card,
.summary-card {
  min-height: 360px;
}
.section-head,
.ledger-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 1.1rem;
}
.section-head h3,
.ledger-header h3 {
  font-size: 1.05rem;
  line-height: 1.25;
}
.section-head p,
.ledger-header p,
.summary-note p,
.muted {
  margin-top: 0.25rem;
  color: var(--text-secondary);
  font-size: 0.82rem;
}
.panel-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 64px;
  height: 30px;
  padding: 0 0.75rem;
  border-radius: 999px;
  color: var(--primary);
  background: var(--primary-light);
  font-size: 0.76rem;
  font-weight: 700;
}
.field-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.35rem;
}
.field-head label {
  margin-bottom: 0;
}
.field-head span {
  color: var(--text-muted);
  font-size: 0.78rem;
}
.form-actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.compact-head {
  margin-bottom: 1rem;
}
.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
  margin-bottom: 1rem;
}
.summary-grid > div {
  min-height: 92px;
  padding: 1rem 0.85rem;
  border-radius: 12px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-sm);
  transition: all var(--transition);
}
.summary-grid > div:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow);
}
.summary-grid strong {
  display: block;
  font-size: 1.3rem;
  line-height: 1.2;
}
.summary-grid span {
  color: var(--text-muted);
  font-size: 0.78rem;
}
.summary-note {
  padding: 1rem;
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--bg-card);
}
.summary-note strong {
  font-size: 0.88rem;
}
.ledger-header {
  align-items: center;
}
.ledger-actions {
  display: grid;
  grid-template-columns: minmax(220px, 320px) 140px auto;
  gap: 0.75rem;
}
.jobs-table {
  table-layout: fixed;
}
.jobs-table th:nth-child(1) { width: 8%; }
.jobs-table th:nth-child(2) { width: 18%; }
.jobs-table th:nth-child(3) { width: 12%; }
.jobs-table th:nth-child(4) { width: 28%; }
.jobs-table th:nth-child(5) { width: 21%; }
.jobs-table th:nth-child(6) { width: 13%; }
.job-title {
  font-weight: 700;
}
.owner-tag {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 0 0.65rem;
  border-radius: 999px;
  color: #0F7A43;
  background: #EAF8EF;
  font-size: 0.78rem;
  font-weight: 800;
  white-space: nowrap;
}
.owner-tag.external {
  color: var(--text-secondary);
  background: #F1F5F9;
}
.description-cell {
  color: var(--text-secondary);
  line-height: 1.7;
}
.status-stack {
  display: grid;
  gap: 0.45rem;
}
.status-stack span:last-child {
  color: var(--text-secondary);
  font-size: 0.78rem;
  line-height: 1.55;
}
.action-group {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.readonly-text {
  color: var(--text-muted);
  font-size: 0.8rem;
  font-weight: 700;
}
.btn-outline-danger {
  background: #fff;
  color: var(--danger);
  border: 1px solid rgba(239, 68, 68, 0.45);
}
.btn-outline-danger:hover {
  transform: translateY(-1px);
  background: #FEF2F2;
}
.jobs-empty {
  display: grid;
  gap: 0.3rem;
}
.jobs-empty strong {
  color: var(--text);
  font-size: 1rem;
}
.jobs-empty span {
  color: var(--text-secondary);
}
.compact {
  padding: 1.4rem 1rem;
}
@media (max-width: 1080px) {
  .job-layout {
    grid-template-columns: 1fr;
  }
  .ledger-header {
    align-items: stretch;
    flex-direction: column;
  }
  .ledger-actions {
    grid-template-columns: 1fr;
  }
}
</style>
