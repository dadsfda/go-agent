const API = '/api'

export async function request(path, options = {}) {
  const headers = options.headers || {}
  if (!(options.body instanceof FormData)) headers['Content-Type'] = 'application/json'
  const token = localStorage.getItem('hrToken')
  if (token) headers.Authorization = `Bearer ${token}`
  const res = await fetch(`${API}${path}`, { ...options, headers })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    if (res.status === 401) {
      localStorage.removeItem('hrToken')
      localStorage.removeItem('hrUser')
      window.location.href = '/login'
      throw new Error('登录已过期，请重新登录')
    }
    throw new Error(formatError(data.error || '请求失败'))
  }
  return data
}

function formatError(msg) {
  const s = String(msg || '')
    .replace(/^rpc error: code = [A-Za-z]+ desc = /, '')
    .replace(/^rpc error: code = [A-Za-z]+ /, '')
    .trim()
  if (s.includes('账号或密码错误')) return '账号或密码错误，请检查邮箱、密码和角色'
  if (s.includes('账号已存在') || s.includes('Duplicate entry') || s.includes('duplicate')) return '该账号已注册，请直接登录或换一个邮箱'
  if (s.includes('邮箱和密码不能为空')) return '请填写邮箱和密码'
  if (s.includes('无权限访问') || s.includes('角色权限不足')) return '当前账号无权限执行该操作'
  if (s.includes('无权操作他人岗位')) return '只能管理自己创建的岗位'
  if (s.includes('岗位名称和描述不能为空')) return '请填写岗位名称和岗位描述'
  if (s.includes('DeepSeek') || s.includes('Eino')) return 'AI 配置异常，请检查 config.yaml 中的 DeepSeek 配置'
  return s || '请求失败，请稍后重试'
}
