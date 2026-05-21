<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { request } from '../api.js'
import { auth } from '../stores/auth.js'

const emit = defineEmits(['toast'])
const messages = ref([])
const question = ref('')
const loading = ref(false)
const chatEl = ref(null)

async function loadHistory() {
  try {
    const data = await request('/hr/ai/history')
    messages.value = data.messages || []
    await nextTick()
    scrollToBottom()
  } catch {}
}

function scrollToBottom() {
  if (chatEl.value) chatEl.value.scrollTop = chatEl.value.scrollHeight
}

async function send() {
  const q = question.value.trim()
  if (!q || loading.value) return
  loading.value = true
  try {
    const data = await request('/hr/ai', {
      method: 'POST',
      body: JSON.stringify({ question: q })
    })
    messages.value.push({ question: q, answer: data.answer })
    question.value = ''
    await nextTick()
    scrollToBottom()
  } catch (e) {
    emit('toast', { message: e.message, type: 'error' })
  } finally {
    loading.value = false
  }
}

onMounted(loadHistory)
</script>

<template>
  <div>
    <div class="page-header">
      <h1>AI 业务问答</h1>
      <p>用自然语言查询投递统计、岗位热度等 MySQL 业务数据</p>
    </div>

    <div class="ai-layout">
      <div class="card ai-side">
        <div class="ai-side-title">可询问内容</div>
        <div class="ai-chip">投递统计</div>
        <div class="ai-chip">岗位热度</div>
        <div class="ai-chip">候选人分布</div>
        <p>回答基于当前业务数据生成，适合快速查看招聘运营情况。</p>
      </div>

    <div class="card chat-card">
      <div ref="chatEl" class="chat-messages">
        <div v-if="!messages.length" class="empty" style="padding:2rem">
          还没有对话记录，试试问：投递总人数是多少？
        </div>
        <div v-for="(msg, i) in messages" :key="i" class="chat-pair">
          <div class="bubble bubble-user">
            <div class="bubble-label">HR</div>
            <div>{{ msg.question }}</div>
          </div>
          <div class="bubble bubble-ai">
            <div class="bubble-label">AI</div>
            <div>{{ msg.answer }}</div>
          </div>
        </div>
      </div>

      <form class="chat-input" @submit.prevent="send">
        <input v-model="question" placeholder="例如：投递总人数是多少？" :disabled="loading" />
        <button type="submit" class="btn-primary" :disabled="loading || !question.trim()">
          {{ loading ? '思考中...' : '发送' }}
        </button>
      </form>
    </div>
    </div>
  </div>
</template>

<style scoped>
.ai-layout {
  display: grid;
  grid-template-columns: minmax(220px, 300px) minmax(0, 1fr);
  gap: 1.25rem;
}
.ai-side,
.chat-card {
  min-height: calc(100vh - 180px);
}
.ai-side {
  background: linear-gradient(180deg, #FFFFFF 0%, #F8FBFF 100%);
}
.ai-side-title {
  font-size: 0.88rem;
  font-weight: 800;
  color: var(--text-secondary);
  margin-bottom: 0.85rem;
}
.ai-chip {
  width: fit-content;
  margin-bottom: 0.55rem;
  padding: 0.35rem 0.75rem;
  border-radius: 999px;
  background: var(--primary-light);
  color: var(--primary);
  font-size: 0.8rem;
  font-weight: 800;
}
.ai-side p {
  margin-top: 1rem;
  color: var(--text-secondary);
  font-size: 0.9rem;
  line-height: 1.8;
}
.chat-card {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 180px);
  background: linear-gradient(180deg, #FFFFFF 0%, #F8FBFF 100%);
}
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.chat-pair { display: flex; flex-direction: column; gap: 0.5rem; }
.bubble {
  padding: 0.85rem 1rem;
  border-radius: 16px;
  max-width: 80%;
  font-size: 0.875rem;
  line-height: 1.6;
}
.bubble-label {
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  margin-bottom: 0.25rem;
  opacity: 0.6;
}
.bubble-user {
  align-self: flex-end;
  background: linear-gradient(135deg, #3B82F6, #1D4ED8);
  color: #fff;
}
.bubble-user .bubble-label { color: rgba(255,255,255,0.7); }
.bubble-ai {
  align-self: flex-start;
  background: #fff;
  color: var(--text);
  border: 1px solid var(--border);
}
.chat-input {
  display: flex;
  gap: 0.75rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border);
}
.chat-input input { flex: 1; }
@media (max-width: 960px) {
  .ai-layout {
    grid-template-columns: 1fr;
  }
  .ai-side,
  .chat-card {
    min-height: auto;
  }
  .chat-card {
    height: calc(100vh - 260px);
  }
}
</style>
