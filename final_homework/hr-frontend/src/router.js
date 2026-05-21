import { createRouter, createWebHistory } from 'vue-router'
import { auth } from './stores/auth.js'

const routes = [
  { path: '/login', name: 'login', component: () => import('./views/Login.vue') },
  { path: '/', name: 'dashboard', component: () => import('./views/Dashboard.vue'), meta: { auth: true } },
  { path: '/jobs', name: 'jobs', component: () => import('./views/JobManage.vue'), meta: { auth: true } },
  { path: '/applications', name: 'applications', component: () => import('./views/Applications.vue'), meta: { auth: true } },
  { path: '/ai', name: 'ai', component: () => import('./views/AIChat.vue'), meta: { auth: true } }
]

export const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  if (to.meta.auth && !auth.loggedIn) return { name: 'login' }
  if (to.name === 'login' && auth.loggedIn) return { name: 'dashboard' }
})
