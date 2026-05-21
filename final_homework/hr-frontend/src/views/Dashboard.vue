<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { request } from '../api.js'
import { auth } from '../stores/auth.js'

const router = useRouter()
const stats = ref({ jobs: 0, applications: 0 })

onMounted(async () => {
  try {
    const [jobsData, appsData] = await Promise.all([
      request('/hr/jobs'),
      request('/hr/applications?page=1&pageSize=1')
    ])
    stats.value.jobs = jobsData.jobs?.length || 0
    stats.value.applications = appsData.total || 0
  } catch {}
})
</script>

<template>
  <div>
    <div class="page-header">
      <h1>工作台</h1>
      <p>欢迎回来，{{ auth.user?.email }}</p>
    </div>

    <div class="stat-grid">
      <div class="stat-card" @click="router.push('/jobs')">
        <div class="stat-icon jobs">💼</div>
        <div>
          <div class="stat-value">{{ stats.jobs }}</div>
          <div class="stat-label">在招岗位</div>
        </div>
      </div>
      <div class="stat-card" @click="router.push('/applications')">
        <div class="stat-icon people">👥</div>
        <div>
          <div class="stat-value">{{ stats.applications }}</div>
          <div class="stat-label">投递候选</div>
        </div>
      </div>
      <div class="stat-card" @click="router.push('/ai')">
        <div class="stat-icon ai">🤖</div>
        <div>
          <div class="stat-value">AI</div>
          <div class="stat-label">智能问答</div>
        </div>
      </div>
    </div>

    <div class="card quick-card">
      <div class="quick-head">
        <h3>快速开始</h3>
        <span>HR 工作流</span>
      </div>
      <ul>
        <li>1. 在「岗位管理」中发布招聘岗位</li>
        <li>2. 候选人投递后在「候选人台账」查看</li>
        <li>3. 使用「AI 问答」查询投递统计数据</li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
.stat-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}
.stat-card {
  background: var(--card);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  border: 1px solid rgba(220, 229, 242, 0.92);
  min-height: 148px;
  padding: 1.4rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  cursor: pointer;
  transition: transform var(--transition), box-shadow var(--transition);
}
.stat-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-md);
}
.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  flex-shrink: 0;
}
.stat-icon.jobs { background: #EAF2FF; color: #2563EB; }
.stat-icon.people { background: #DCFCE7; color: #15803D; }
.stat-icon.ai { background: #FEF3C7; color: #B45309; }
.stat-value { font-size: 1.5rem; font-weight: 700; color: var(--text); }
.stat-label { font-size: 0.8rem; color: var(--text-muted); }
.quick-card {
  margin-top: 1.5rem;
  min-height: 220px;
}
.quick-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}
.quick-head h3 {
  font-size: 1.05rem;
}
.quick-head span {
  padding: 0.3rem 0.7rem;
  border-radius: 999px;
  background: var(--primary-light);
  color: var(--primary);
  font-size: 0.78rem;
  font-weight: 700;
}
.quick-card ul {
  color: var(--text-secondary);
  font-size: 0.9rem;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}
.quick-card li {
  padding: 0.9rem 1rem;
  border-radius: 14px;
  background: #F8FBFF;
  border: 1px solid rgba(220, 229, 242, 0.9);
}
</style>
