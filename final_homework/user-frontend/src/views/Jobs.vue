<script setup>
import { ref, onMounted } from 'vue'
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

onMounted(loadJobs)
</script>

<template>
  <div>
    <div class="page-header">
      <h1>公开岗位</h1>
      <p>{{ auth.loggedIn ? '点击投递按钮申请岗位' : '登录后可投递岗位' }}</p>
    </div>

    <div v-if="!jobs.length" class="empty card">暂无公开岗位，请等待 HR 发布</div>

    <div v-else class="job-grid stagger-in">
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
          <p class="description-text">{{ selectedJob.description || '暂无岗位描述' }}</p>
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
  padding: 1.5rem;
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  box-shadow: var(--shadow);
  transition: all var(--transition);
}
.job-card:hover {
  transform: translateY(-5px);
  box-shadow: var(--shadow-lg);
  border-color: rgba(22, 163, 74, 0.12);
}
.job-main {
  display: grid;
  gap: 0.9rem;
}
.job-kicker {
  display: inline-flex;
  width: fit-content;
  margin-bottom: 0.5rem;
  padding: 0.24rem 0.65rem;
  border-radius: 999px;
  background: var(--primary-light);
  color: var(--primary);
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.02em;
}
.job-main h3 {
  font-size: 1.1rem;
  line-height: 1.4;
  margin-bottom: 0.45rem;
  font-weight: 800;
}
.job-main p {
  color: var(--text-secondary);
  font-size: 0.88rem;
  line-height: 1.75;
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
  font-size: 0.84rem;
}
.detail-link:hover { text-decoration: underline; }
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
  background: rgba(15, 23, 42, 0.35);
  animation: maskIn 0.2s ease;
}
@keyframes maskIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
.detail-panel {
  width: min(720px, 100%);
  height: 100%;
  display: grid;
  grid-template-rows: auto 1fr auto;
  background: var(--bg-card);
  box-shadow: -16px 0 48px rgba(15, 23, 42, 0.15);
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
  padding: 1.5rem;
  border-bottom: 1px solid var(--border-light);
}
.detail-kicker {
  display: inline-block;
  color: var(--primary);
  font-size: 0.76rem;
  font-weight: 700;
  margin-bottom: 0.35rem;
  letter-spacing: 0.02em;
}
.detail-head h2 {
  font-size: 1.3rem;
  line-height: 1.4;
  font-weight: 800;
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
.description-text {
  color: var(--text-secondary);
  font-size: 0.92rem;
  line-height: 1.85;
  white-space: pre-wrap;
}
.detail-actions {
  display: flex;
  justify-content: flex-end;
  padding: 1rem 1.5rem 1.5rem;
  border-top: 1px solid var(--border-light);
  background: var(--bg-card);
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
