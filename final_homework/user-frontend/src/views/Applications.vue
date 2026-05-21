<script setup>
import { ref, onMounted } from 'vue'
import { request } from '../api.js'

const emit = defineEmits(['toast'])
const items = ref([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const data = await request('/candidate/applications')
    items.value = data.items || data.applications || []
  } catch (e) {
    emit('toast', { message: e.message, type: 'error' })
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="page-header">
      <h1>我的投递</h1>
      <p>查看已投递的岗位和状态</p>
    </div>

    <div v-if="loading" class="empty card">加载中...</div>
    <div v-else-if="!items.length" class="empty card">暂无投递记录，去「浏览岗位」投递吧</div>

    <div v-else class="app-list">
      <div v-for="app in items" :key="app.id" class="app-card card">
        <div class="app-header">
          <div>
            <span class="app-kicker">投递记录</span>
            <h3>{{ app.job?.title || '未知岗位' }}</h3>
          </div>
          <span class="badge badge-open">已投递</span>
        </div>
        <p class="app-desc">{{ app.job?.description }}</p>
        <div class="app-meta">
          <span>投递时间：{{ app.createdAt }}</span>
          <span>简历：{{ app.resume?.fileName || '未知' }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.app-list { display: flex; flex-direction: column; gap: 1rem; max-width: 920px; }
.app-card {
  min-height: 172px;
  background: linear-gradient(180deg, #FFFFFF 0%, #F8FBFF 100%);
  transition: transform var(--transition), box-shadow var(--transition);
}
.app-card:hover { transform: translateY(-2px); box-shadow: var(--shadow-md); }
.app-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 0.5rem; }
.app-kicker {
  color: var(--primary);
  font-size: 0.74rem;
  font-weight: 800;
}
.app-header h3 { font-size: 1.08rem; margin-top: 0.18rem; }
.app-desc { color: var(--text-secondary); font-size: 0.85rem; margin-bottom: 0.75rem; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.app-meta { display: flex; gap: 1.5rem; font-size: 0.8rem; color: var(--text-muted); }
</style>
