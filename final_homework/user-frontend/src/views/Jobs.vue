<script setup>
import { computed, ref, onMounted } from 'vue'
import { request } from '../api.js'
import { auth } from '../stores/auth.js'

const emit = defineEmits(['toast'])
const jobs = ref([])
const appliedIds = ref(new Set())
const selectedJob = ref(null)

async function loadApplications() {
  if (!auth.loggedIn) { appliedIds.value = new Set(); return }
  try {
    const data = await request('/candidate/applications')
    const items = data.items || data.applications || []
    appliedIds.value = new Set(items.map(a => Number(a.jobId)))
  } catch {}
}

async function loadJobs() {
  await loadApplications()
  try {
    const data = await request('/jobs')
    jobs.value = data.jobs || []
  } catch (e) {
    emit('toast', { message: e.message, type: 'error' })
  }
}

async function applyJob(jobId) {
  if (!auth.loggedIn) return emit('toast', { message: '请先登录再投递', type: 'error' })
  try {
    await request('/candidate/applications', {
      method: 'POST',
      body: JSON.stringify({ jobId })
    })
    appliedIds.value.add(jobId)
    emit('toast', { message: '投递成功', type: 'success' })
  } catch (e) {
    emit('toast', { message: e.message, type: 'error' })
  }
}

const selectedDescriptionBlocks = computed(() => buildDescriptionBlocks(selectedJob.value?.description))

function openDetail(job) {
  selectedJob.value = job
}

function closeDetail() {
  selectedJob.value = null
}

function previewDescription(text, length = 76) {
  const value = String(text || '').replace(/\s+/g, ' ').trim()
  if (!value) return '暂无岗位描述'
  return value.length > length ? value.slice(0, length) + '...' : value
}

function normalizeDescription(text) {
  return String(text || '')
    .replace(/\r/g, '')
    .replace(/([。；;])\s*(?=(?:\d+[.、]|[（(]\d+[）)]))/g, '$1\n')
    .trim()
}

function buildDescriptionBlocks(text) {
  const lines = normalizeDescription(text)
    .split(/\n+/)
    .map(line => line.trim())
    .filter(Boolean)

  const blocks = []
  let listItems = []
  const flushList = () => {
    if (!listItems.length) return
    blocks.push({ type: 'list', items: listItems })
    listItems = []
  }

  lines.forEach(line => {
    const matched = line.match(/^(?:[-*•]|(?:\d+[.、])|(?:[（(]\d+[）)]))\s*(.+)$/)
    if (matched) {
      listItems.push(matched[1])
      return
    }
    flushList()
    blocks.push({ type: 'paragraph', text: line })
  })
  flushList()
  return blocks.length ? blocks : [{ type: 'paragraph', text: '暂无岗位描述' }]
}

onMounted(loadJobs)
</script>

<template>
  <div>
    <div class="page-header">
      <h1>公开岗位</h1>
      <p>{{ auth.loggedIn ? '点击投递按钮申请岗位' : '登录后可投递岗位' }}</p>
    </div>

    <div v-if="!jobs.length" class="empty card">暂无公开岗位，请等待 HR 发布</div>

    <div v-else class="job-grid">
      <div v-for="job in jobs" :key="job.id" class="job-card">
        <div class="job-main">
          <div>
            <span class="job-kicker">开放岗位</span>
            <h3>{{ job.title }}</h3>
            <p>{{ previewDescription(job.description) }}</p>
          </div>
          <button class="detail-link" @click="openDetail(job)">查看详情</button>
        </div>
        <div class="job-foot">
          <span class="meta">ID: {{ job.id }}</span>
          <div class="job-actions">
            <button class="btn-secondary btn-sm" @click="openDetail(job)">展开</button>
            <button
              v-if="auth.loggedIn"
              class="btn-primary btn-sm"
              :disabled="appliedIds.has(Number(job.id))"
              @click="applyJob(Number(job.id))"
            >
              {{ appliedIds.has(Number(job.id)) ? '已投递' : '一键投递' }}
            </button>
            <router-link v-else to="/login" class="btn-secondary btn-sm login-link">
              登录后投递
            </router-link>
          </div>
        </div>
      </div>
    </div>

    <div v-if="selectedJob" class="detail-mask" @click.self="closeDetail">
      <aside class="detail-panel">
        <div class="detail-head">
          <div>
            <span class="detail-kicker">岗位详情</span>
            <h2>{{ selectedJob.title }}</h2>
            <p>ID: {{ selectedJob.id }}</p>
          </div>
          <button class="btn-ghost close-btn" @click="closeDetail">关闭</button>
        </div>

        <div class="detail-body">
          <section
            v-for="(block, index) in selectedDescriptionBlocks"
            :key="`${block.type}-${index}`"
            class="description-block"
          >
            <p v-if="block.type === 'paragraph'">{{ block.text }}</p>
            <ul v-else>
              <li v-for="item in block.items" :key="item">{{ item }}</li>
            </ul>
          </section>
        </div>

        <div class="detail-actions">
          <button
            v-if="auth.loggedIn"
            class="btn-primary"
            :disabled="appliedIds.has(Number(selectedJob.id))"
            @click="applyJob(Number(selectedJob.id))"
          >
            {{ appliedIds.has(Number(selectedJob.id)) ? '已投递' : '投递这个岗位' }}
          </button>
          <router-link v-else to="/login" class="btn-primary detail-login-link">
            登录后投递
          </router-link>
        </div>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.job-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 1.15rem;
}
.job-card {
  min-height: 236px;
  display: grid;
  gap: 1rem;
  align-content: space-between;
  padding: 1.45rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: linear-gradient(180deg, #FFFFFF 0%, #F8FBFF 100%);
  box-shadow: var(--shadow);
  transition: transform var(--transition), box-shadow var(--transition), border-color var(--transition);
}
.job-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-md);
  border-color: rgba(37, 99, 235, 0.22);
}
.job-main {
  display: grid;
  gap: 0.9rem;
}
.job-kicker {
  display: inline-flex;
  margin-bottom: 0.5rem;
  padding: 0.22rem 0.58rem;
  border-radius: 999px;
  background: var(--primary-light);
  color: var(--primary);
  font-size: 0.74rem;
  font-weight: 800;
}
.job-main h3 {
  font-size: 1.12rem;
  line-height: 1.4;
  margin-bottom: 0.45rem;
}
.job-main p {
  color: var(--text-secondary);
  font-size: 0.9rem;
  line-height: 1.8;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.detail-link {
  width: fit-content;
  padding: 0;
  border-radius: 0;
  background: transparent;
  color: var(--primary);
  font-weight: 700;
}
.job-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}
.job-actions {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  flex-wrap: wrap;
  justify-content: flex-end;
}
.login-link,
.detail-login-link {
  text-decoration: none;
}
.detail-mask {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  justify-content: flex-end;
  background: rgba(15, 23, 42, 0.36);
}
.detail-panel {
  width: min(720px, 100%);
  height: 100%;
  display: grid;
  grid-template-rows: auto 1fr auto;
  background: linear-gradient(180deg, #FFFFFF 0%, #F8FBFF 100%);
  box-shadow: -12px 0 36px rgba(15, 23, 42, 0.18);
}
.detail-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.5rem;
  border-bottom: 1px solid var(--border);
}
.detail-kicker {
  display: inline-block;
  color: var(--primary);
  font-size: 0.78rem;
  font-weight: 700;
  margin-bottom: 0.35rem;
}
.detail-head h2 {
  font-size: 1.35rem;
  line-height: 1.4;
}
.detail-head p {
  margin-top: 0.25rem;
  color: var(--text-muted);
  font-size: 0.85rem;
}
.close-btn {
  flex: 0 0 auto;
}
.detail-body {
  overflow-y: auto;
  padding: 1.5rem;
}
.description-block {
  padding: 1rem 1.1rem;
  border-radius: 16px;
  background: #fff;
  border: 1px solid rgba(220, 229, 242, 0.92);
}
.description-block + .description-block {
  margin-top: 1rem;
}
.description-block p,
.description-block li {
  color: var(--text-secondary);
  font-size: 0.95rem;
  line-height: 1.95;
}
.description-block p {
  white-space: pre-line;
}
.description-block ul {
  padding-left: 1.3rem;
}
.description-block li + li {
  margin-top: 0.45rem;
}
.detail-actions {
  display: flex;
  justify-content: flex-end;
  padding: 1rem 1.5rem 1.5rem;
  border-top: 1px solid var(--border);
  background: #fff;
}
@media (max-width: 720px) {
  .job-grid {
    grid-template-columns: 1fr;
  }
  .job-foot {
    align-items: flex-start;
    flex-direction: column;
  }
  .job-actions {
    justify-content: flex-start;
  }
}
</style>
