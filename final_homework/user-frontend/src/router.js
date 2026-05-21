import { createRouter, createWebHistory } from 'vue-router'
import { auth } from './stores/auth.js'

const routes = [
  { path: '/login', name: 'login', component: () => import('./views/Login.vue') },
  { path: '/', name: 'jobs', component: () => import('./views/Jobs.vue') },
  { path: '/profile', name: 'profile', component: () => import('./views/Profile.vue'), meta: { auth: true } },
  { path: '/resume', redirect: '/profile' },
  { path: '/applications', name: 'applications', component: () => import('./views/Applications.vue'), meta: { auth: true } }
]

export const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  if (to.meta.auth && !auth.loggedIn) return { name: 'login' }
  if (to.name === 'login' && auth.loggedIn) return { name: 'jobs' }
})
