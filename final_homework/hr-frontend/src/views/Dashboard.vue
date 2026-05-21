<script setup>
import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { request } from '../api.js'
import { auth } from '../stores/auth.js'
import { buildChartPoints, buildJobRanking, buildTrendDays, countOwnedJobs, formatTrendRange } from '../dashboardMetrics.js'

const router = useRouter()
const stats = ref({ jobs: 0, applications: 0 })
const applications = ref([])
const recent = ref([])

onMounted(async () => {
  try {
    const [jobsData, appsData] = await Promise.all([
      request('/hr/jobs'),
      request('/hr/applications?page=1&pageSize=100')
    ])
    const items = appsData.items || appsData.applications || []
    stats.value.jobs = countOwnedJobs(jobsData.jobs, auth.user?.id)
    stats.value.applications = appsData.total || 0
    applications.value = items
    recent.value = items.slice(0, 5)
  } catch {}
})

const trendDays = computed(() => buildTrendDays(applications.value))
const trendRange = computed(() => formatTrendRange(trendDays.value))
const chartPoints = computed(() => buildChartPoints(trendDays.value))
const jobRanking = computed(() => buildJobRanking(applications.value))
const todayCount = computed(() => trendDays.value.at(-1)?.count || 0)
</script>

<template>
  <div>
    <div class="page-header">
      <div>
        <h1>&#24037;&#20316;&#21488;</h1>
        <p>&#27426;&#36814;&#22238;&#26469;&#65292;{{ auth.user?.email }}</p>
      </div>
    </div>

    <div class="stat-grid stagger-in">
      <div class="stat-card" @click="router.push('/jobs')">
        <div class="stat-icon orange">J</div>
        <div>
          <div class="stat-label">&#25105;&#21457;&#24067;&#30340;&#23703;&#20301;</div>
          <div class="stat-value">{{ stats.jobs }} <span>&#20010;</span></div>
        </div>
      </div>
      <div class="stat-card" @click="router.push('/applications')">
        <div class="stat-icon blue">C</div>
        <div>
          <div class="stat-label">&#25237;&#36882;&#24635;&#20154;&#25968;</div>
          <div class="stat-value">{{ stats.applications }} <span>&#20154;</span></div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon green">T</div>
        <div>
          <div class="stat-label">&#20170;&#26085;&#26032;&#25237;&#36882;</div>
          <div class="stat-value">{{ todayCount }} <span>&#20154;</span></div>
        </div>
      </div>
      <div class="stat-card" @click="router.push('/ai')">
        <div class="stat-icon purple">AI</div>
        <div>
          <div class="stat-label">&#24453;&#31579;&#36873;&#31616;&#21382;</div>
          <div class="stat-value">{{ stats.applications }} <span>&#20221;</span></div>
        </div>
      </div>
    </div>

    <div class="dashboard-grid">
      <section class="card chart-card">
        <div class="section-head">
          <h3>&#25237;&#36882;&#36235;&#21183;&#65288;&#36817;7&#22825;&#65289;</h3>
          <span>{{ trendRange }}</span>
        </div>
        <div class="chart-stage">
          <div class="grid-lines"></div>
          <svg viewBox="0 0 640 220" role="img" aria-label="trend chart">
            <polyline v-if="chartPoints" :points="chartPoints" fill="none" stroke="rgba(31,115,241,.14)" stroke-width="18" stroke-linecap="round" stroke-linejoin="round" />
            <polyline v-if="chartPoints" :points="chartPoints" fill="none" stroke="#1F73F1" stroke-width="5" stroke-linecap="round" stroke-linejoin="round" />
            <g fill="#1F73F1">
              <circle
                v-for="point in chartPoints.split(' ')"
                :key="point"
                :cx="point.split(',')[0]"
                :cy="point.split(',')[1]"
                r="5"
              />
            </g>
          </svg>
          <div class="chart-labels">
            <span v-for="day in trendDays" :key="day.key">{{ day.label }} · {{ day.count }}</span>
          </div>
        </div>
      </section>

      <section class="card rank-card">
        <div class="section-head">
          <h3>&#23703;&#20301;&#25237;&#36882; TOP5</h3>
          <button class="btn-ghost btn-sm" @click="router.push('/jobs')">&#26356;&#22810;</button>
        </div>
        <ol v-if="jobRanking.length" class="rank-list">
          <li v-for="item in jobRanking" :key="item.title">
            <span>{{ item.title }}</span>
            <strong>{{ item.count }}</strong>
          </li>
        </ol>
        <div v-else class="empty compact">&#26242;&#26080;&#25237;&#36882;&#25968;&#25454;</div>
      </section>
    </div>

    <section class="card recent-card">
      <div class="section-head">
        <h3>&#26368;&#36817;&#25237;&#36882;&#30340;&#20505;&#36873;&#20154;</h3>
        <button class="btn-ghost btn-sm" @click="router.push('/applications')">&#26356;&#22810;</button>
      </div>
      <div v-if="!recent.length" class="empty compact">&#26242;&#26080;&#25237;&#36882;&#35760;&#24405;</div>
      <table v-else>
        <thead>
          <tr>
            <th>&#22995;&#21517;</th>
            <th>&#24212;&#32856;&#23703;&#20301;</th>
            <th>&#23398;&#21382;</th>
            <th>&#25237;&#36882;&#26102;&#38388;</th>
            <th>&#25805;&#20316;</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="app in recent" :key="app.id">
            <td>{{ app.profile?.name || '-' }}</td>
            <td>{{ app.job?.title || '-' }}</td>
            <td>{{ app.profile?.education || '-' }}</td>
            <td>{{ app.createdAt || '-' }}</td>
            <td><button class="btn-secondary btn-sm" @click="router.push('/applications')">&#26597;&#30475;</button></td>
          </tr>
        </tbody>
      </table>
    </section>
  </div>
</template>

<style scoped>
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.9rem;
  margin-bottom: 1rem;
}
.stat-card {
  min-height: 96px;
  display: flex;
  align-items: center;
  gap: 0.9rem;
  padding: 1.1rem;
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  border: 1px solid var(--border);
  box-shadow: var(--shadow);
  cursor: pointer;
  transition: all var(--transition);
}
.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
  border-color: rgba(37, 99, 235, 0.12);
}
.stat-icon {
  width: 46px;
  height: 46px;
  display: grid;
  place-items: center;
  border-radius: 12px;
  font-size: 0.82rem;
  font-weight: 900;
}
.stat-icon.orange { background: linear-gradient(135deg, #FFF1DD, #FFE8C4); color: #EA580C; box-shadow: 0 4px 12px rgba(234, 88, 12, 0.15); }
.stat-icon.blue { background: linear-gradient(135deg, #DBEAFE, #BFDBFE); color: #2563EB; box-shadow: 0 4px 12px rgba(37, 99, 235, 0.15); }
.stat-icon.green { background: linear-gradient(135deg, #D1FAE5, #BBF7D0); color: #16A34A; box-shadow: 0 4px 12px rgba(22, 163, 74, 0.15); }
.stat-icon.purple { background: linear-gradient(135deg, #E8E0FF, #DDD6FE); color: #7C3AED; box-shadow: 0 4px 12px rgba(124, 58, 237, 0.15); }
.stat-label {
  color: var(--text-secondary);
  font-size: 0.78rem;
  margin-bottom: 0.2rem;
  font-weight: 500;
}
.stat-value {
  color: var(--text);
  font-size: 1.6rem;
  font-weight: 900;
  line-height: 1.1;
  letter-spacing: -0.02em;
}
.stat-value span {
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--text-muted);
}
.dashboard-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.7fr) minmax(260px, 0.9fr);
  gap: 1rem;
  margin-bottom: 1rem;
}
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.8rem;
  margin-bottom: 0.85rem;
}
.section-head h3 {
  font-size: 0.95rem;
  font-weight: 700;
}
.section-head span {
  color: var(--text-muted);
  font-size: 0.76rem;
}
.chart-stage {
  position: relative;
  height: 240px;
  overflow: hidden;
  border-radius: 12px;
  background: #F8FAFC;
}
.grid-lines {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(215, 226, 241, 0.4) 1px, transparent 1px),
    linear-gradient(90deg, rgba(215, 226, 241, 0.4) 1px, transparent 1px);
  background-size: 100% 48px, 80px 100%;
}
.chart-stage svg {
  position: relative;
  z-index: 1;
  width: 100%;
  height: calc(100% - 30px);
  padding: 12px;
}
.chart-labels {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 0.35rem;
  padding: 0 0.7rem 0.55rem;
  color: var(--text-muted);
  font-size: 0.72rem;
  text-align: center;
}
.rank-list {
  display: grid;
  gap: 0.55rem;
  list-style: none;
  counter-reset: rank;
}
.rank-list li {
  counter-increment: rank;
  display: grid;
  grid-template-columns: 26px 1fr auto;
  align-items: center;
  gap: 0.6rem;
  min-height: 34px;
  padding: 0.3rem 0;
  color: var(--text-secondary);
  font-size: 0.84rem;
  border-bottom: 1px solid rgba(215, 226, 241, 0.3);
}
.rank-list li:last-child { border-bottom: none; }
.rank-list li::before {
  content: counter(rank);
  width: 24px;
  height: 24px;
  display: grid;
  place-items: center;
  border-radius: 7px;
  background: var(--primary-light);
  color: var(--primary);
  font-size: 0.72rem;
  font-weight: 900;
}
.rank-list strong {
  color: var(--text);
  font-size: 0.82rem;
  font-weight: 700;
}
.recent-card {
  padding-bottom: 1rem;
}
.compact {
  padding: 1.3rem 1rem;
}
@media (max-width: 1080px) {
  .stat-grid,
  .dashboard-grid {
    grid-template-columns: 1fr 1fr;
  }
}
@media (max-width: 720px) {
  .stat-grid,
  .dashboard-grid {
    grid-template-columns: 1fr;
  }
}
</style>
