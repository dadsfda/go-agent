const API = '/api'

export async function request(path, options = {}) {
  const headers = options.headers || {}
  if (!(options.body instanceof FormData)) headers['Content-Type'] = 'application/json'
  const token = localStorage.getItem('candidateToken')
  if (token) headers.Authorization = `Bearer ${token}`
  const res = await fetch(`${API}${path}`, { ...options, headers })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    if (res.status === 401) {
      localStorage.removeItem('candidateToken')
      localStorage.removeItem('candidateUser')
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
  if (s.includes('候选人档案必填字段不能为空')) return '请完整填写姓名、电话、学历、学校、经历和技能'
  if (s.includes('技能标签不能超过')) return s
  if (s.includes('请先完善结构化个人档案')) return '请先保存完整个人资料，再投递岗位'
  if (s.includes('请先上传合规简历')) return '请先上传 PDF、DOC 或 DOCX 简历，再投递岗位'
  if (s.includes('请勿重复投递')) return '该岗位已经投递过了'
  if (s.includes('简历仅支持')) return '简历仅支持真实 PDF、DOC、DOCX 文件，请检查文件格式'
  if (s.includes('TENCENT_SECRET') || s.includes('COS')) return '简历上传配置异常，请检查 config.yaml 中的 COS 配置'
  return s || '请求失败，请稍后重试'
}
