<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { auth } from './stores/auth.js'

const router = useRouter()
const route = useRoute()
const toasts = ref([])
let toastId = 0

function showToast(payload, type = 'info') {
  const msg = typeof payload === 'object' ? payload.message : payload
  const t = typeof payload === 'object' ? payload.type : type
  const id = ++toastId
  toasts.value.push({ id, message: msg, type: t })
  setTimeout(() => { toasts.value = toasts.value.filter(x => x.id !== id) }, 3000)
}

function handleLogout() {
  auth.logout()
  router.push('/')
}

const navItems = [
  { path: '/', label: '浏览岗位', icon: '💼' },
  { path: '/profile', label: '个人档案', icon: '📋' },
  { path: '/applications', label: '我的投递', icon: '📮' }
]

window.__showToast = showToast
</script>

<template>
  <div class="app-layout">
    <aside class="sidebar">
      <div class="sidebar-brand">
        <div class="brand-icon">🎯</div>
        <div>
          <div class="brand-title">智能招聘</div>
          <div class="brand-sub">{{ auth.loggedIn ? '候选人端' : '游客模式' }}</div>
        </div>
      </div>

      <nav class="sidebar-nav">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="nav-item"
          :class="{ active: route.path === item.path }"
        >
          <span class="nav-icon">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <template v-if="auth.loggedIn">
          <div class="user-info">
            <div class="user-avatar">{{ (auth.user?.email || 'C')[0].toUpperCase() }}</div>
            <div class="user-email">{{ auth.user?.email }}</div>
          </div>
          <button class="btn-ghost btn-sm" @click="handleLogout">退出</button>
        </template>
        <router-link v-else to="/login" class="btn-primary btn-sm" style="text-decoration:none;width:100%;text-align:center">
          登录 / 注册
        </router-link>
      </div>
    </aside>

    <main class="main-content">
      <router-view @toast="showToast" />
    </main>
  </div>

  <div class="toast-container">
    <div v-for="t in toasts" :key="t.id" class="toast" :class="'toast-' + t.type">
      {{ t.message }}
    </div>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: var(--sidebar-width);
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(14px);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 10;
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1.25rem;
  border-bottom: 1px solid var(--border);
}
.brand-icon { font-size: 1.75rem; }
.brand-title { font-weight: 800; font-size: 1rem; color: var(--text); }
.brand-sub { font-size: 0.75rem; color: var(--text-muted); }

.sidebar-nav {
  flex: 1;
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 0.85rem;
  border-radius: 14px;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all var(--transition);
}
.nav-item:hover { background: rgba(37, 99, 235, 0.08); color: var(--text); transform: translateX(2px); }
.nav-item.active {
  background: linear-gradient(135deg, #EAF2FF 0%, #F5F9FF 100%);
  color: var(--primary);
  box-shadow: inset 0 0 0 1px rgba(37, 99, 235, 0.12);
}
.nav-icon { font-size: 1.1rem; }

.sidebar-footer {
  padding: 1rem;
  border-top: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.user-info { display: flex; align-items: center; gap: 0.5rem; min-width: 0; }
.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--primary-light);
  color: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.8rem;
  flex-shrink: 0;
}
.user-email {
  font-size: 0.8rem;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100px;
}

.main-content {
  flex: 1;
  margin-left: var(--sidebar-width);
  padding: 2.25rem;
  min-height: 100vh;
}
</style>
