<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { request } from '../api.js'

const emit = defineEmits(['toast'])
const messages = ref([])
const question = ref('')
const loading = ref(false)
const chatEl = ref(null)

const suggestions = [
  '投递总人数是多少？',
  '哪个岗位最热门？',
  '列出 Java 后端岗位的候选人',
  '有哪些候选人技能包含 MySQL？'
]

const dataScopes = ['投递统计', '岗位热度', '候选人筛选', '简历档案', '聊天历史']

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
  question.value = ''
  messages.value.push({ question: q, answer: '' })
  loading.value = true
  await nextTick()
  scrollToBottom()
  try {
    const data = await request('/hr/ai', {
      method: 'POST',
      body: JSON.stringify({ question: q })
    })
    messages.value[messages.value.length - 1].answer = data.answer
  } catch (e) {
    messages.value.pop()
    emit('toast', { message: e.message, type: 'error' })
  } finally {
    loading.value = false
    await nextTick()
    scrollToBottom()
  }
}

async function clearHistory() {
  if (loading.value || !messages.value.length) return
  if (!window.confirm('确认清空当前账号的 AI 对话历史？')) return
  try {
    await request('/hr/ai/history', { method: 'DELETE' })
    messages.value = []
    emit('toast', { message: '对话历史已清空', type: 'success' })
  } catch (e) {
    emit('toast', { message: e.message, type: 'error' })
  }
}

function askSuggestion(text) {
  question.value = text
  send()
}

function formatAnswer(text) {
  return escapeHtml(text || '')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\n/g, '<br>')
}

function escapeHtml(text) {
  return String(text)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

onMounted(loadHistory)
</script>

<template>
  <div class="copilot-page">
    <div class="copilot-ambient one"></div>
    <div class="copilot-ambient two"></div>

    <header class="copilot-hero">
      <div>
        <h1>AI 数据分析助手</h1>
      </div>
    </header>

    <div class="copilot-shell">
      <aside class="glass-sidebar">
        <div class="assistant-badge">
          <div class="assistant-mark">AI</div>
          <div>
            <strong>招聘数据中枢</strong>
          </div>
        </div>

        <section class="side-section">
          <div class="side-heading">可查询数据</div>
          <div class="scope-list">
            <span v-for="scope in dataScopes" :key="scope">{{ scope }}</span>
          </div>
        </section>

        <section class="side-section">
          <div class="side-heading">推荐问题</div>
          <button
            v-for="item in suggestions"
            :key="item"
            type="button"
            class="suggestion"
            :disabled="loading"
            @click="askSuggestion(item)"
          >
            <span>{{ item }}</span>
            <b>↗</b>
          </button>
        </section>

        <div class="insight-panel">
          <strong>真实业务库</strong>
        </div>
      </aside>

      <main class="chat-card">
        <div class="chat-topbar">
          <div>
            <h2>Copilot 对话</h2>
          </div>
          <div class="top-actions">
            <button type="button" class="clear-btn" :disabled="loading || !messages.length" @click="clearHistory">
              清空
            </button>
            <span class="live-pill"><i></i>{{ loading ? '分析中' : '就绪' }}</span>
          </div>
        </div>

        <div ref="chatEl" class="chat-messages">
          <div v-if="!messages.length" class="empty-state">
            <div class="empty-orbit">
              <span></span>
            </div>
            <h3>开始一次招聘数据分析</h3>
          </div>

          <div v-for="(msg, i) in messages" :key="i" class="chat-pair">
            <div class="message-row user-row">
              <div class="bubble bubble-user">
                <div>{{ msg.question }}</div>
              </div>
            </div>
            <div class="message-row ai-row">
              <div class="avatar ai-avatar">AI</div>
              <div v-if="msg.answer" class="bubble bubble-ai">
                <div v-html="formatAnswer(msg.answer)"></div>
              </div>
              <div v-else class="bubble bubble-ai thinking">
                <span></span><span></span><span></span>
              </div>
            </div>
          </div>
        </div>

        <form class="floating-input" @submit.prevent="send">
          <div class="input-shell">
            <span class="input-icon">AI</span>
            <input v-model="question" placeholder="询问投递人数、热门岗位、候选人技能..." :disabled="loading" />
            <button type="submit" :disabled="loading || !question.trim()">
              {{ loading ? '分析中' : '发送' }}
            </button>
          </div>
        </form>
      </main>
    </div>
  </div>
</template>

<style scoped>
.copilot-page {
  position: relative;
  min-height: calc(100vh - 48px);
  padding-bottom: 0.5rem;
  overflow: hidden;
}
.copilot-page::before {
  content: '';
  position: absolute;
  inset: -2rem -2rem auto -2rem;
  height: 340px;
  background:
    linear-gradient(120deg, rgba(37, 99, 235, 0.1), rgba(79, 70, 229, 0.08), rgba(20, 184, 166, 0.08)),
    radial-gradient(circle at 22% 34%, rgba(255, 255, 255, 0.82), transparent 32%);
  border-radius: 0 0 40px 40px;
  pointer-events: none;
}
.copilot-ambient {
  position: absolute;
  border-radius: 999px;
  filter: blur(20px);
  opacity: 0.5;
  pointer-events: none;
}
.copilot-ambient.one {
  width: 180px;
  height: 180px;
  top: 76px;
  right: 12%;
  background: rgba(59, 130, 246, 0.14);
}
.copilot-ambient.two {
  width: 140px;
  height: 140px;
  bottom: 52px;
  left: 28%;
  background: rgba(20, 184, 166, 0.12);
}
.copilot-hero,
.copilot-shell {
  position: relative;
  z-index: 1;
}
.copilot-hero {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.95rem;
}
.copilot-hero h1 {
  font-size: clamp(1.35rem, 2vw, 1.8rem);
  letter-spacing: 0;
}
.copilot-shell {
  display: grid;
  grid-template-columns: minmax(240px, 320px) minmax(0, 1fr);
  gap: 1.35rem;
  align-items: stretch;
}
.glass-sidebar {
  min-height: calc(100vh - 178px);
  padding: 1.15rem;
  border: 1px solid rgba(255, 255, 255, 0.75);
  border-radius: 22px;
  background: linear-gradient(155deg, rgba(255, 255, 255, 0.82), rgba(248, 250, 252, 0.68));
  box-shadow: 0 20px 56px rgba(71, 85, 105, 0.12), inset 0 1px 0 rgba(255, 255, 255, 0.88);
  backdrop-filter: blur(18px);
}
.assistant-badge {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.8rem;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.72);
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.12);
}
.assistant-mark {
  width: 44px;
  height: 44px;
  border-radius: 15px;
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, #2563EB, #4F46E5);
  color: white;
  font-weight: 900;
  box-shadow: 0 14px 26px rgba(79, 70, 229, 0.28);
}
.assistant-badge strong {
  display: block;
}
.assistant-badge strong {
  color: #0F172A;
  font-size: 0.92rem;
}
.side-section {
  margin-top: 1.15rem;
}
.side-heading {
  margin-bottom: 0.65rem;
  color: #64748B;
  font-size: 0.74rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}
.scope-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.scope-list span {
  padding: 0.36rem 0.62rem;
  border-radius: 999px;
  background: rgba(37, 99, 235, 0.08);
  color: #3154C9;
  font-size: 0.78rem;
  font-weight: 800;
}
.suggestion {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.55rem;
  padding: 0.72rem 0.78rem;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 15px;
  background: rgba(255, 255, 255, 0.66);
  color: #1E293B;
  text-align: left;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.04);
}
.suggestion:hover:not(:disabled) {
  border-color: rgba(99, 102, 241, 0.3);
  background: rgba(255, 255, 255, 0.9);
  transform: translateY(-2px);
}
.suggestion b {
  color: #6366F1;
  font-size: 0.82rem;
}
.insight-panel {
  margin-top: 1.2rem;
  padding: 0.95rem 1rem;
  border-radius: 18px;
  background:
    linear-gradient(145deg, rgba(37, 99, 235, 0.9), rgba(79, 70, 229, 0.86)),
    #2563EB;
  color: white;
  box-shadow: 0 18px 38px rgba(37, 99, 235, 0.24);
}
.insight-panel strong {
  display: block;
  font-size: 1rem;
}
.chat-card {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 178px);
  min-height: 620px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(248, 250, 252, 0.88)),
    #FFFFFF;
  box-shadow: 0 28px 80px rgba(30, 41, 59, 0.14);
}
.chat-topbar {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
  padding: 1rem 1.15rem;
  border-bottom: 1px solid rgba(226, 232, 240, 0.8);
  background: rgba(255, 255, 255, 0.62);
  backdrop-filter: blur(12px);
}
.chat-topbar h2 {
  font-size: 1rem;
  letter-spacing: 0;
}
.top-actions {
  display: flex;
  align-items: center;
  gap: 0.55rem;
}
.clear-btn {
  padding: 0.42rem 0.62rem;
  border-radius: 999px;
  background: #FFFFFF;
  color: #64748B;
  font-size: 0.76rem;
  font-weight: 800;
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.18);
}
.clear-btn:hover:not(:disabled) {
  color: #2563EB;
  background: #F8FAFC;
  transform: translateY(-1px);
}
.live-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.42rem 0.62rem;
  border-radius: 999px;
  background: #F8FAFC;
  color: #475569;
  font-size: 0.76rem;
  font-weight: 800;
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.16);
}
.live-pill i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #22C55E;
}
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 1.35rem 1.25rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  background:
    linear-gradient(rgba(148, 163, 184, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 163, 184, 0.05) 1px, transparent 1px);
  background-size: 34px 34px;
}
.empty-state {
  margin: auto;
  max-width: 420px;
  text-align: center;
  color: #64748B;
}
.empty-orbit {
  width: 76px;
  height: 76px;
  margin: 0 auto 1rem;
  display: grid;
  place-items: center;
  border-radius: 24px;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.12), rgba(124, 58, 237, 0.12));
  box-shadow: inset 0 0 0 1px rgba(99, 102, 241, 0.12);
}
.empty-orbit span {
  width: 32px;
  height: 32px;
  border-radius: 12px;
  background: linear-gradient(135deg, #2563EB, #7C3AED);
  box-shadow: 0 14px 30px rgba(79, 70, 229, 0.25);
}
.empty-state h3 {
  color: #0F172A;
  font-size: 1.05rem;
  letter-spacing: 0;
}
.empty-state p {
  display: none;
}
.chat-pair {
  display: flex;
  flex-direction: column;
  gap: 0.7rem;
}
.message-row {
  display: flex;
  align-items: flex-start;
  gap: 0.65rem;
}
.user-row {
  justify-content: flex-end;
}
.ai-row {
  justify-content: flex-start;
}
.avatar {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  border-radius: 12px;
  font-size: 0.72rem;
  font-weight: 900;
}
.ai-avatar {
  background: #EEF2FF;
  color: #4F46E5;
  box-shadow: inset 0 0 0 1px rgba(99, 102, 241, 0.12);
}
.bubble {
  padding: 0.92rem 1.05rem;
  max-width: min(72%, 780px);
  font-size: 0.875rem;
  line-height: 1.6;
  animation: bubbleIn 0.3s var(--spring);
  white-space: pre-wrap;
}
@keyframes bubbleIn {
  from { opacity: 0; transform: translateY(8px) scale(0.97); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
.bubble-label {
  font-size: 0.7rem;
  font-weight: 800;
  text-transform: uppercase;
  margin-bottom: 0.25rem;
  opacity: 0.62;
}
.bubble-user {
  border-radius: 18px 18px 6px 18px;
  background: linear-gradient(135deg, #2563EB 0%, #4F46E5 100%);
  color: #fff;
  box-shadow: 0 18px 38px rgba(79, 70, 229, 0.24);
}
.bubble-user .bubble-label {
  color: rgba(255, 255, 255, 0.72);
}
.bubble-ai {
  border-radius: 18px 18px 18px 6px;
  background: rgba(255, 255, 255, 0.96);
  color: var(--text);
  border: 1px solid rgba(226, 232, 240, 0.9);
  box-shadow: 0 16px 36px rgba(15, 23, 42, 0.08);
}
.bubble-ai strong {
  font-weight: 900;
  color: #1D4ED8;
}
.thinking {
  display: inline-flex;
  align-items: center;
  gap: 0.32rem;
  min-width: 82px;
}
.thinking span {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #818CF8;
  animation: thinking 1s infinite ease-in-out;
}
.thinking span:nth-child(2) { animation-delay: 0.14s; }
.thinking span:nth-child(3) { animation-delay: 0.28s; }
@keyframes thinking {
  0%, 80%, 100% { transform: translateY(0); opacity: 0.35; }
  40% { transform: translateY(-5px); opacity: 1; }
}
.floating-input {
  padding: 1rem 1.25rem 1.15rem;
  background: linear-gradient(180deg, rgba(248, 250, 252, 0.2), rgba(255, 255, 255, 0.92));
}
.input-shell {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 16px 45px rgba(15, 23, 42, 0.12);
}
.input-icon {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  border-radius: 12px;
  background: #EEF2FF;
  color: #4F46E5;
  font-size: 0.72rem;
  font-weight: 900;
}
.input-shell input {
  border: 0;
  background: transparent;
  box-shadow: none;
  padding: 0.62rem 0.15rem;
}
.input-shell input:focus {
  box-shadow: none;
}
.input-shell button {
  min-width: 82px;
  border-radius: 13px;
  background: linear-gradient(135deg, #2563EB, #4F46E5);
  color: white;
  box-shadow: 0 12px 28px rgba(79, 70, 229, 0.26);
}
.input-shell button:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 16px 34px rgba(79, 70, 229, 0.32);
}
@media (max-width: 960px) {
  .copilot-hero {
    align-items: flex-start;
    flex-direction: column;
  }
  .copilot-shell {
    grid-template-columns: 1fr;
  }
  .glass-sidebar,
  .chat-card {
    min-height: auto;
  }
  .glass-sidebar {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }
  .assistant-badge,
  .insight-panel {
    grid-column: 1 / -1;
  }
  .chat-card {
    height: 680px;
  }
  .bubble {
    max-width: 88%;
  }
}
@media (max-width: 640px) {
  .glass-sidebar {
    display: block;
  }
  .live-pill {
    display: none;
  }
  .chat-topbar {
    padding: 1rem;
  }
  .chat-messages {
    padding: 1rem;
  }
  .input-shell {
    grid-template-columns: minmax(0, 1fr) auto;
  }
  .input-icon {
    display: none;
  }
}
</style>
