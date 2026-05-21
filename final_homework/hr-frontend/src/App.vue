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
  router.push('/login')
}

const navItems = [
  { path: '/', label: '\u9996\u9875', icon: '\ud83c\udfe0' },
  { path: '/jobs', label: '\u5c97\u4f4d\u7ba1\u7406', icon: '\ud83d\udcbc' },
  { path: '/applications', label: '\u5019\u9009\u4eba\u53f0\u8d26', icon: '\ud83d\udccb' },
  { path: '/ai', label: 'AI \u667a\u80fd\u95ee\u7b54', icon: '\ud83e\udd16' }
]

window.__showToast = showToast
</script>

<template>
  <div v-if="!auth.loggedIn" class="login-layout">
    <div class="login-title-pill">HR &#31649;&#29702;&#31471;</div>
    <router-view @toast="showToast" />
  </div>

  <div v-else class="app-layout">
    <aside class="sidebar">
      <div class="sidebar-brand">
        <div class="brand-mark">HR</div>
        <div>
          <div class="brand-title">&#26234;&#33021;&#25307;&#32856;</div>
          <div class="brand-sub">&#31649;&#29702;&#21518;&#21488;</div>
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
        <div class="user-info">
          <div class="user-avatar">{{ (auth.user?.email || 'H')[0].toUpperCase() }}</div>
          <div class="user-email">{{ auth.user?.email }}</div>
        </div>
        <button class="btn-ghost btn-sm" @click="handleLogout">&#36864;&#20986;</button>
      </div>
    </aside>

    <main class="main-content">
      <div class="topbar">
        <button class="btn-ghost btn-sm" @click="router.back()">&lt; &#36820;&#22238;</button>
        <div class="topbar-user">
          <span>{{ auth.user?.email }}</span>
          <span class="role-dot">HR</span>
        </div>
      </div>
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
/* ============ Login Layout ============ */
.login-layout {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 4rem 9vw;
  overflow: hidden;
  background: linear-gradient(135deg, #EFF6FF 0%, #F0F9FF 40%, #F8FAFC 100%);
}
/* Animated floating orbs */
.login-layout::before {
  content: "";
  position: absolute;
  left: -10%;
  top: -25%;
  width: 650px;
  height: 650px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(37, 99, 235, 0.1) 0%, transparent 70%);
  animation: floatOrb1 18s ease-in-out infinite;
  filter: blur(40px);
}
.login-layout::after {
  content: "";
  position: absolute;
  right: -8%;
  bottom: -20%;
  width: 550px;
  height: 550px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(59, 130, 246, 0.08) 0%, transparent 70%);
  animation: floatOrb2 22s ease-in-out infinite;
  filter: blur(50px);
}
@keyframes floatOrb1 {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(40px, -30px) scale(1.05); }
  66% { transform: translate(-20px, 20px) scale(0.95); }
}
@keyframes floatOrb2 {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(-30px, 25px) scale(0.95); }
  66% { transform: translate(25px, -15px) scale(1.05); }
}
.login-title-pill {
  position: absolute;
  left: 0;
  top: 0;
  padding: 0.9rem 1.6rem;
  border-radius: 0 0 16px 0;
  background: var(--primary-gradient);
  color: #fff;
  font-family: var(--font-heading);
  font-size: 1.25rem;
  font-weight: 700;
  box-shadow: 0 8px 24px rgba(37, 99, 235, 0.2);
  z-index: 2;
}

/* ============ App Layout ============ */
.app-layout {
  display: flex;
  min-height: 100vh;
  background: var(--bg);
}

/* ============ Sidebar ============ */
.sidebar {
  width: var(--sidebar-width);
  background: #FFFFFF;
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  position: fixed;
  inset: 0 auto 0 0;
  z-index: 10;
  box-shadow: 1px 0 3px rgba(0,0,0,0.04);
}
.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  padding: 1.15rem 1.1rem;
  border-bottom: 1px solid var(--border-light);
}
.brand-mark {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  background: var(--primary-gradient);
  color: #fff;
  font-family: var(--font-heading);
  font-weight: 800;
  font-size: 0.82rem;
  box-shadow: 0 4px 14px rgba(37, 99, 235, 0.25);
  flex-shrink: 0;
}
.brand-title {
  font-family: var(--font-heading);
  font-weight: 700;
  font-size: 0.95rem;
  color: var(--text);
}
.brand-sub {
  font-size: 0.72rem;
  color: var(--text-muted);
  margin-top: 0.05rem;
}

/* ============ Sidebar Nav ============ */
.sidebar-nav {
  flex: 1;
  padding: 1rem 0.85rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  min-height: 42px;
  padding: 0.6rem 0.8rem;
  border-radius: 10px;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.84rem;
  font-weight: 600;
  transition: all var(--transition);
  position: relative;
}
.nav-item:hover {
  background: rgba(37, 99, 235, 0.06);
  color: var(--primary);
}
.nav-item.active {
  background: var(--primary-gradient);
  color: #fff;
  box-shadow: 0 4px 16px rgba(37, 99, 235, 0.28);
}
/* Glow trail on active */
.nav-item.active::after {
  content: '';
  position: absolute;
  inset: -2px;
  border-radius: 12px;
  background: var(--primary-gradient);
  opacity: 0.15;
  filter: blur(8px);
  z-index: -1;
}
.nav-icon {
  width: 20px;
  text-align: center;
  font-size: 0.95rem;
  flex-shrink: 0;
}

/* ============ Sidebar Footer ============ */
.sidebar-footer {
  padding: 0.9rem;
  border-top: 1px solid var(--border-light);
  display: grid;
  gap: 0.65rem;
}
.user-info {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  min-width: 0;
}
.user-avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  background: var(--primary-light);
  color: var(--primary);
  display: grid;
  place-items: center;
  font-weight: 700;
  font-size: 0.78rem;
  flex-shrink: 0;
  border: 2px solid rgba(37, 99, 235, 0.1);
}
.user-email {
  font-size: 0.76rem;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ============ Main Content ============ */
.main-content {
  flex: 1;
  margin-left: var(--sidebar-width);
  padding: 1.1rem 1.3rem 1.5rem;
  min-height: 100vh;
  background: var(--bg);
}
.topbar {
  height: 46px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.85rem;
}
.topbar-user {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  color: var(--text-secondary);
  font-size: 0.82rem;
}
.role-dot {
  padding: 0.22rem 0.6rem;
  border-radius: 999px;
  background: var(--primary-light);
  color: var(--primary);
  font-family: var(--font-heading);
  font-weight: 700;
  font-size: 0.72rem;
}

@media (max-width: 760px) {
  .login-layout {
    justify-content: center;
    padding: 4rem 1rem;
  }
}
</style>
