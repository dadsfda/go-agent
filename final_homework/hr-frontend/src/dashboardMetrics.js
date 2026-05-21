export function buildTrendDays(applications, baseDate = latestApplicationDate(applications) || new Date()) {
  const days = []
  const counts = new Map()

  for (const app of applications || []) {
    const key = dateKey(appDate(app))
    if (key) counts.set(key, (counts.get(key) || 0) + 1)
  }

  for (let offset = 6; offset >= 0; offset--) {
    const date = new Date(baseDate)
    date.setHours(0, 0, 0, 0)
    date.setDate(date.getDate() - offset)
    const key = dateKey(date)
    days.push({
      key,
      label: formatMonthDay(date),
      count: counts.get(key) || 0
    })
  }
  return days
}

export function buildJobRanking(applications) {
  const counts = new Map()
  for (const app of applications || []) {
    const title = String(app?.job?.title || '未知岗位').trim() || '未知岗位'
    counts.set(title, (counts.get(title) || 0) + 1)
  }
  return [...counts.entries()]
    .map(([title, count]) => ({ title, count }))
    .sort((a, b) => b.count - a.count || a.title.localeCompare(b.title, 'zh-CN'))
    .slice(0, 5)
}

export function countOwnedJobs(jobs, userId) {
  const ownerId = Number(userId)
  return (jobs || []).filter(job => Number(job?.ownerHrId || job?.owner_hr_id) === ownerId).length
}

export function buildChartPoints(days, width = 640, height = 220) {
  if (!days?.length) return ''
  const max = Math.max(...days.map(item => item.count), 1)
  const left = 20
  const right = width - 20
  const top = 36
  const bottom = height - 34
  const step = days.length === 1 ? 0 : (right - left) / (days.length - 1)
  return days
    .map((item, index) => {
      const x = left + step * index
      const y = bottom - (item.count / max) * (bottom - top)
      return `${x.toFixed(1)},${y.toFixed(1)}`
    })
    .join(' ')
}

export function formatTrendRange(days) {
  if (!days?.length) return ''
  return `${days[0].label} - ${days[days.length - 1].label}`
}

function dateKey(value) {
  if (!value) return ''
  const date = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function formatMonthDay(value) {
  const date = value instanceof Date ? value : new Date(value)
  return `${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

function latestApplicationDate(applications) {
  let latest = null
  for (const app of applications || []) {
    const date = appDate(app)
    if (date && (!latest || date > latest)) latest = date
  }
  return latest
}

function appDate(app) {
  const value = app?.createdAt || app?.created_at
  if (!value) return null
  const date = value instanceof Date ? value : new Date(value)
  return Number.isNaN(date.getTime()) ? null : date
}
