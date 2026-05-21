<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { request } from '../api.js'
import { auth } from '../stores/auth.js'

const emit = defineEmits(['toast'])
const router = useRouter()
const email = ref('')
const password = ref('')
const loading = ref(false)

async function handleAuth(mode) {
  if (!email.value.trim()) return emit('toast', { message: '请输入邮箱', type: 'error' })
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value)) return emit('toast', { message: '邮箱格式不正确', type: 'error' })
  if (!password.value) return emit('toast', { message: '请输入密码', type: 'error' })
  if (mode === 'register' && password.value.length < 6) return emit('toast', { message: '密码长度不能少于 6 位', type: 'error' })

  loading.value = true
  try {
    const data = await request(`/auth/${mode}`, {
      method: 'POST',
      body: JSON.stringify({ role: 'hr', email: email.value.trim(), password: password.value })
    })
    auth.setSession(data)
    emit('toast', { message: mode === 'login' ? '登录成功' : '注册成功', type: 'success' })
    router.push('/')
  } catch (e) {
    emit('toast', { message: e.message, type: 'error' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-card">
    <div class="login-brand">
      <div class="login-icon">HR</div>
      <h1>智能招聘管理端</h1>
      <p>HR 账号登录，管理岗位与候选人</p>
    </div>
    <form @submit.prevent="handleAuth('login')">
      <div class="form-group">
        <label>邮箱</label>
        <input v-model="email" type="email" placeholder="HR 邮箱" />
      </div>
      <div class="form-group">
        <label>密码</label>
        <input v-model="password" type="password" placeholder="密码" />
      </div>
      <div class="login-actions">
        <button type="submit" class="btn-primary" :disabled="loading" style="flex:1">
          {{ loading ? '登录中...' : '登录' }}
        </button>
        <button type="button" class="btn-secondary" :disabled="loading" @click="handleAuth('register')">
          注册 HR
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.login-card {
  background: linear-gradient(180deg, #FFFFFF 0%, #F8FBFF 100%);
  border: 1px solid rgba(220, 229, 242, 0.92);
  border-radius: 20px;
  padding: 2.4rem;
  width: 100%;
  max-width: 430px;
  box-shadow: var(--shadow-md);
}
.login-brand {
  text-align: center;
  margin-bottom: 2rem;
}
.login-icon {
  width: 58px;
  height: 58px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.9rem;
  border-radius: 18px;
  background: linear-gradient(135deg, #3B82F6, #1D4ED8);
  color: #fff;
  font-size: 1rem;
  font-weight: 900;
  box-shadow: 0 16px 30px rgba(37, 99, 235, 0.25);
}
.login-brand h1 { font-size: 1.4rem; color: var(--text); margin-bottom: 0.25rem; }
.login-brand p { font-size: 0.85rem; color: var(--text-muted); }
.login-actions { display: flex; gap: 0.75rem; margin-top: 1.5rem; }
</style>
