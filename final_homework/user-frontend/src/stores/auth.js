import { reactive } from 'vue'

export const auth = reactive({
  token: localStorage.getItem('candidateToken') || '',
  user: JSON.parse(localStorage.getItem('candidateUser') || 'null'),

  get loggedIn() {
    return Boolean(this.token)
  },

  setSession(data) {
    this.token = data.token
    this.user = data.user
    localStorage.setItem('candidateToken', this.token)
    localStorage.setItem('candidateUser', JSON.stringify(this.user))
  },

  logout() {
    this.token = ''
    this.user = null
    localStorage.removeItem('candidateToken')
    localStorage.removeItem('candidateUser')
  }
})
