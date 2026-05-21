import { reactive } from 'vue'

export const auth = reactive({
  token: localStorage.getItem('hrToken') || '',
  user: JSON.parse(localStorage.getItem('hrUser') || 'null'),

  get loggedIn() {
    return Boolean(this.token)
  },

  setSession(data) {
    this.token = data.token
    this.user = data.user
    localStorage.setItem('hrToken', this.token)
    localStorage.setItem('hrUser', JSON.stringify(this.user))
  },

  logout() {
    this.token = ''
    this.user = null
    localStorage.removeItem('hrToken')
    localStorage.removeItem('hrUser')
  }
})
