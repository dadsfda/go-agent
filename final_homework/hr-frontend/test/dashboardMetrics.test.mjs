import assert from 'node:assert/strict'
import { buildJobRanking, buildTrendDays, countOwnedJobs } from '../src/dashboardMetrics.js'

const apps = [
  { createdAt: '2026-05-15T09:00:00+08:00', job: { title: '后端工程师' } },
  { createdAt: '2026-05-15T18:00:00+08:00', job: { title: '后端工程师' } },
  { createdAt: '2026-05-16T10:00:00+08:00', job: { title: '前端工程师' } },
  { createdAt: '2026-05-17T10:00:00+08:00', job: { title: '后端工程师' } }
]

const trend = buildTrendDays(apps, new Date('2026-05-17T12:00:00+08:00'))
assert.equal(trend.length, 7)
assert.deepEqual(
  trend.slice(-3).map(item => ({ label: item.label, count: item.count })),
  [
    { label: '05-15', count: 2 },
    { label: '05-16', count: 1 },
    { label: '05-17', count: 1 }
  ]
)

const ranking = buildJobRanking(apps)
assert.deepEqual(ranking, [
  { title: '后端工程师', count: 3 },
  { title: '前端工程师', count: 1 }
])

const historicalTrend = buildTrendDays(apps)
assert.deepEqual(
  historicalTrend.slice(-3).map(item => ({ label: item.label, count: item.count })),
  [
    { label: '05-15', count: 2 },
    { label: '05-16', count: 1 },
    { label: '05-17', count: 1 }
  ]
)

assert.equal(countOwnedJobs([
  { id: 1, ownerHrId: 7 },
  { id: 2, ownerHrId: 8 },
  { id: 3, owner_hr_id: 7 }
], 7), 2)
