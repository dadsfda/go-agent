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
  { path: '/', label: '\u9996\u9875' },
  { path: '/profile', label: '\u4e2a\u4eba\u8d44\u6599' },
  { path: '/applications', label: '\u6211\u7684\u6295\u9012' }
]

window.__showToast = showToast
</script>

<template>
  <div class="candidate-shell">
    <header class="candidate-header">
      <router-link to="/" class="brand">
        <span class="brand-mark">+</span>
        <span>&#26234;&#33021;&#25307;&#32856;&#31995;&#32479;</span>
      </router-link>

      <nav class="top-nav">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="nav-item"
          :class="{ active: route.path === item.path }"
        >
          {{ item.label }}
        </router-link>
      </nav>

      <div class="header-actions">
        <template v-if="auth.loggedIn">
          <span class="user-email">{{ auth.user?.email }}</span>
          <button class="btn-ghost btn-sm" @click="handleLogout">&#36864;&#20986;</button>
        </template>
        <router-link v-else to="/login" class="login-link">&#30331;&#24405; / &#27880;&#20876;</router-link>
      </div>
    </header>

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
.candidate-shell {
  min-height: 100vh;
  background: var(--bg);
}

/* ============ Glass Header ============ */
.candidate-header {
  position: sticky;
  top: 0;
  z-index: 20;
  height: 62px;
  display: grid;
  grid-template-columns: minmax(170px, 240px) 1fr minmax(160px, 240px);
  align-items: center;
  gap: 1rem;
  padding: 0 1.5rem;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border);
  box-shadow: var(--shadow-sm);
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  color: var(--text);
  text-decoration: none;
  white-space: nowrap;
}
.brand-mark {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: var(--primary-gradient);
  color: #fff;
  font-family: var(--font-heading);
  font-size: 0.85rem;
  font-weight: 800;
  line-height: 1;
  box-shadow: 0 3px 10px rgba(22, 163, 74, 0.2);
}
.brand span:last-child {
  font-family: var(--font-heading);
  font-weight: 700;
}

.top-nav {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.3rem;
}
.nav-item {
  position: relative;
  display: inline-flex;
  align-items: center;
  height: 62px;
  padding: 0 1rem;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.85rem;
  font-weight: 600;
  transition: color var(--transition);
}
.nav-item:hover { color: var(--primary); }
.nav-item.active { color: var(--primary); font-weight: 700; }
.nav-item.active::after {
  content: "";
  position: absolute;
  left: 50%;
  bottom: 0;
  width: 36px;
  height: 3px;
  border-radius: 999px 999px 0 0;
  background: var(--primary-gradient);
  transform: translateX(-50%);
  box-shadow: 0 2px 8px rgba(22, 163, 74, 0.2);
}

.header-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.7rem;
  min-width: 0;
}
.user-email {
  color: var(--text-secondary);
  font-size: 0.8rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.login-link {
  color: var(--text);
  text-decoration: none;
  font-size: 0.84rem;
  font-weight: 600;
  padding: 0.45rem 1rem;
  border-radius: 10px;
  transition: all var(--transition);
}
.login-link:hover {
  color: var(--primary);
  background: var(--primary-light);
}

.main-content {
  width: min(1200px, calc(100% - 2rem));
  margin: 0 auto;
  padding: 1.3rem 0 1.8rem;
}

@media (max-width: 760px) {
  .candidate-header {
    height: auto;
    grid-template-columns: 1fr;
    padding: 0.85rem 1rem;
    gap: 0.6rem;
  }
  .top-nav {
    justify-content: flex-start;
    gap: 0.5rem;
    overflow-x: auto;
  }
  .nav-item {
    height: 40px;
    flex: 0 0 auto;
  }
  .header-actions {
    justify-content: flex-start;
  }
}
</style>
