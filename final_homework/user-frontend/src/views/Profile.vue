<script setup>
import { ref, onMounted } from 'vue'
import { request } from '../api.js'

const emit = defineEmits(['toast'])
const form = ref({
  name: '', phone: '', education: '', school: '', experience: '', skills: ''
})
const loading = ref(false)
const fromResume = ref(false)
const fileInput = ref(null)
const selectedFile = ref(null)
const uploadedFile = ref('')
const uploading = ref(false)
const parsing = ref(false)
const parseStatus = ref('')
const savedAt = ref('')
const savedFlash = ref(false)
const editing = ref(true)

async function loadProfile() {
  try {
    const data = await request('/candidate/profile')
    let hasProfile = false
    Object.keys(form.value).forEach(k => {
      if (data[k]) {
        form.value[k] = data[k]
        hasProfile = true
      }
    })
    if (hasProfile) editing.value = false
  } catch {}
}

function onFileChange(e) {
  selectedFile.value = e.target.files[0] || null
  parseStatus.value = ''
}

function applyParsedProfile(profile) {
  let filled = 0
  let changed = 0
  Object.keys(form.value).forEach(k => {
    if (profile[k]) {
      filled += 1
      if (form.value[k] !== profile[k]) changed += 1
      form.value[k] = profile[k]
    }
  })
  if (!filled) {
    throw new Error('未从简历中提取到姓名、电话、学历、学校、经历或技能')
  }
  fromResume.value = true
  editing.value = true
  parseStatus.value = `已自动填入 ${filled} 个字段，其中 ${changed} 个字段发生变化`
  emit('toast', { message: `${parseStatus.value}，请确认后保存`, type: 'info' })
}

async function uploadResume() {
  if (!selectedFile.value) return emit('toast', { message: '请选择简历文件', type: 'error' })
  const ext = selectedFile.value.name.split('.').pop().toLowerCase()
  if (!['pdf', 'doc', 'docx'].includes(ext)) {
    return emit('toast', { message: '仅支持 PDF、DOC、DOCX 格式', type: 'error' })
  }

  uploading.value = true
  parsing.value = false
  parseStatus.value = '正在上传简历...'
  try {
    const uploadForm = new FormData()
    uploadForm.append('resume', selectedFile.value)
    const data = await request('/candidate/resume', { method: 'POST', body: uploadForm })
    uploadedFile.value = data.fileName

    if (ext !== 'pdf') {
      parseStatus.value = '简历已上传，当前自动解析仅支持文字版 PDF'
      emit('toast', { message: parseStatus.value + '，请手动填写档案', type: 'info' })
      return
    }

    uploading.value = false
    parsing.value = true
    parseStatus.value = '正在解析简历并提取字段...'
    const parseForm = new FormData()
    parseForm.append('resume', selectedFile.value)
    const profile = await request('/candidate/resume/parse', { method: 'POST', body: parseForm })
    applyParsedProfile(profile)
    selectedFile.value = null
    if (fileInput.value) fileInput.value.value = ''
  } catch (e) {
    parseStatus.value = '解析失败，请手动填写档案'
    emit('toast', { message: e.message, type: 'error' })
  } finally {
    uploading.value = false
    parsing.value = false
  }
}

async function save() {
  const { name, phone, education, school, experience, skills } = form.value
  if (!name.trim() || !phone.trim() || !education.trim() || !school.trim() || !experience.trim() || !skills.trim()) {
    return emit('toast', { message: '所有字段均为必填', type: 'error' })
  }
  loading.value = true
  try {
    await request('/candidate/profile', {
      method: 'POST',
      body: JSON.stringify(form.value)
    })
    savedAt.value = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
    savedFlash.value = true
    editing.value = false
    setTimeout(() => { savedFlash.value = false }, 1800)
    emit('toast', { message: '档案已保存', type: 'success' })
  } catch (e) {
    emit('toast', { message: e.message, type: 'error' })
  } finally {
    loading.value = false
  }
}

onMounted(loadProfile)
</script>

<template>
  <div>
    <div class="page-header">
      <h1>结构化档案</h1>
      <p>填写完整资料后才能投递岗位（所有字段必填）</p>
      <p v-if="fromResume" class="parse-hint">以下信息已从简历中自动提取，请核对后保存</p>
    </div>

    <div class="profile-shell">
      <div class="card profile-side">
        <div class="side-title">档案完整度</div>
        <div class="side-score">6 项</div>
        <p>上传文字版 PDF 可自动提取关键信息，保存后即可投递岗位。</p>
      </div>

    <div class="card profile-card">
      <div v-if="savedAt" class="save-banner" :class="{ flash: savedFlash }">
        <div>
          <div class="save-title">档案已保存</div>
          <div class="save-sub">最近保存时间：{{ savedAt }}</div>
        </div>
        <button v-if="!editing" type="button" class="btn-ghost btn-sm" @click="editing = true">重新编辑</button>
        <span v-else class="save-check">✓</span>
      </div>

      <div v-if="editing" class="resume-panel">
        <div>
          <div class="resume-title">上传简历自动填入</div>
          <div class="resume-sub">上传文字版 PDF 后，系统会解析并填入下方档案字段</div>
        </div>
        <input ref="fileInput" type="file" accept=".pdf,.doc,.docx" class="file-input" @change="onFileChange" />
        <button type="button" class="btn-ghost" :disabled="uploading || parsing" @click="fileInput.click()">
          选择文件
        </button>
        <button type="button" class="btn-primary" :disabled="uploading || parsing || !selectedFile" @click="uploadResume">
          {{ uploading ? '上传中...' : parsing ? '解析中...' : '上传并解析' }}
        </button>
      </div>
      <div v-if="editing && (selectedFile || uploadedFile || parseStatus)" class="parse-status" :class="{ active: uploading || parsing }">
        <span class="status-dot"></span>
        <span>{{ parsing ? '正在解析并自动填入字段...' : uploading ? '正在上传简历...' : parseStatus || ('已选择：' + selectedFile.name) }}</span>
      </div>
      <div v-if="editing && uploadedFile" class="upload-line">已上传：{{ uploadedFile }}</div>

      <div v-if="!editing" class="profile-summary">
        <div class="summary-row"><span>姓名</span><strong>{{ form.name }}</strong></div>
        <div class="summary-row"><span>联系电话</span><strong>{{ form.phone }}</strong></div>
        <div class="summary-row"><span>最高学历</span><strong>{{ form.education }}</strong></div>
        <div class="summary-row"><span>毕业院校</span><strong>{{ form.school }}</strong></div>
        <div class="summary-block">
          <span>工作 / 项目经历</span>
          <p>{{ form.experience }}</p>
        </div>
        <div class="summary-block">
          <span>核心技能标签</span>
          <p>{{ form.skills }}</p>
        </div>
        <button type="button" class="btn-primary" @click="editing = true">重新编辑</button>
      </div>

      <form v-else @submit.prevent="save">
        <div class="form-row">
          <div class="form-group">
            <label>姓名</label>
            <input v-model="form.name" placeholder="真实姓名" />
          </div>
          <div class="form-group">
            <label>联系电话</label>
            <input v-model="form.phone" placeholder="手机号码" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>最高学历</label>
            <input v-model="form.education" placeholder="如：本科、硕士" />
          </div>
          <div class="form-group">
            <label>毕业院校</label>
            <input v-model="form.school" placeholder="学校名称" />
          </div>
        </div>
        <div class="form-group">
          <label>工作 / 项目经历</label>
          <textarea v-model="form.experience" placeholder="描述你的工作或项目经历" rows="4"></textarea>
        </div>
        <div class="form-group">
          <label>核心技能标签</label>
          <input v-model="form.skills" placeholder="如：Go, MySQL, Docker, Vue" />
        </div>
        <button type="submit" class="btn-primary" :disabled="loading || uploading || parsing">
          {{ loading ? '保存中...' : savedFlash ? '已保存' : '保存档案' }}
        </button>
      </form>
    </div>
    </div>
  </div>
</template>

<style scoped>
.profile-shell {
  display: grid;
  grid-template-columns: minmax(220px, 300px) minmax(0, 760px);
  gap: 1.25rem;
  align-items: start;
}
.profile-card,
.profile-side {
  min-height: 360px;
}
.profile-side {
  background: linear-gradient(180deg, #FFFFFF 0%, #F8FBFF 100%);
}
.side-title {
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 800;
}
.side-score {
  margin: 0.6rem 0;
  font-size: 2.1rem;
  font-weight: 900;
  color: var(--primary);
}
.profile-side p {
  color: var(--text-secondary);
  font-size: 0.92rem;
  line-height: 1.8;
}
.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
.parse-hint {
  color: var(--primary);
  font-size: 0.85rem;
  margin-top: 0.25rem;
  font-weight: 500;
}
.save-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.85rem 1rem;
  margin-bottom: 1rem;
  border: 1px solid #BBF7D0;
  border-radius: 16px;
  background: #F0FDF4;
  color: #166534;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.save-banner.flash {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(22, 101, 52, 0.12);
}
.save-title {
  font-weight: 700;
}
.save-sub {
  margin-top: 0.15rem;
  font-size: 0.82rem;
  color: #15803D;
}
.save-check {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #22C55E;
  color: white;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 800;
}
.profile-summary {
  display: grid;
  gap: 0.85rem;
}
.summary-row {
  display: grid;
  grid-template-columns: 120px 1fr;
  gap: 1rem;
  align-items: center;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--border);
}
.summary-row span,
.summary-block span {
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 600;
}
.summary-row strong {
  color: var(--text);
  font-weight: 600;
}
.summary-block {
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--border);
}
.summary-block p {
  margin: 0.4rem 0 0;
  color: var(--text);
  line-height: 1.7;
}
.resume-panel {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 0.75rem;
  align-items: center;
  padding: 1rem;
  margin-bottom: 1rem;
  border: 1px solid var(--border);
  border-radius: 16px;
  background: #F8FBFF;
}
.resume-title {
  font-weight: 700;
  color: var(--text);
}
.resume-sub {
  margin-top: 0.2rem;
  color: var(--text-secondary);
  font-size: 0.85rem;
}
.file-input {
  display: none;
}
.parse-status {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-height: 32px;
  margin-bottom: 1rem;
  color: var(--text-secondary);
  font-size: 0.875rem;
}
.parse-status.active {
  color: var(--primary);
}
.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--border);
  flex-shrink: 0;
}
.parse-status.active .status-dot {
  background: var(--primary);
  animation: pulse 1s ease-in-out infinite;
}
.upload-line {
  margin-bottom: 1rem;
  color: var(--text-muted);
  font-size: 0.85rem;
}
@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 0.45; }
  50% { transform: scale(1.6); opacity: 1; }
}
@media (max-width: 720px) {
  .profile-shell {
    grid-template-columns: 1fr;
  }
  .resume-panel {
    grid-template-columns: 1fr;
  }
  .summary-row {
    grid-template-columns: 1fr;
    gap: 0.25rem;
  }
}
</style>
