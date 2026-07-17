import test from 'node:test'
import assert from 'node:assert/strict'

import { createApiClient } from '../src/api.js'

function response(data, status = 200) {
  return {
    ok: status >= 200 && status < 300,
    status,
    async json() {
      return { code: 0, message: 'ok', data }
    },
  }
}

test('defaults to /api/v1 and adds an idempotency key to writes', async () => {
  const requests = []
  const client = createApiClient({
    fetchImpl: async (url, init) => {
      requests.push({ url, init })
      return response({ id: 'CR-1', status: '已排期' })
    },
  })

  const appointment = await client.checkinAppointment('CR-1')

  assert.equal(appointment.id, 'CR-1')
  assert.equal(requests[0].url, '/api/v1/appointments/CR-1/checkin')
  assert.equal(requests[0].init.method, 'POST')
  assert.match(requests[0].init.headers['Idempotency-Key'], /^cf-/)
})

test('uses a configured API origin without duplicating the API path', async () => {
  const requests = []
  const client = createApiClient({
    baseUrl: 'http://localhost:8080/api/v1/',
    fetchImpl: async (url) => {
      requests.push(url)
      return response({ list: [], total: 0 })
    },
  })

  await client.listAppointments({ page: 1, pageSize: 20 })

  assert.equal(requests[0], 'http://localhost:8080/api/v1/appointments?page=1&pageSize=20')
})

test('rejects non-zero API envelopes so callers can keep demo data', async () => {
  const client = createApiClient({
    fetchImpl: async () => ({
      ok: false,
      status: 409,
      async json() {
        return { code: 409, message: '状态不可推进', data: null }
      },
    }),
  })

  await assert.rejects(() => client.updateAppointmentStatus('CR-1', '待制作'), /状态不可推进/)
})

test('exposes mobile lifecycle and follow-up operations through the same client', async () => {
  const paths = []
  const client = createApiClient({
    fetchImpl: async (url) => {
      paths.push(url)
      return response({ id: 'ok' })
    },
  })

  await client.createAppointment({ patient: '选题《城市夜行》', department: '短视频' })
  await client.checkinAppointment('CR-1')
  await client.updateAppointmentStatus('CR-1', '待制作')
  await client.updateAppointmentStatus('CR-1', '制作中')
  await client.updateAppointmentStatus('CR-1', '已发布')
  await client.completeFollowup('FW-1')

  assert.deepEqual(paths, [
    '/api/v1/appointments',
    '/api/v1/appointments/CR-1/checkin',
    '/api/v1/appointments/CR-1/status',
    '/api/v1/appointments/CR-1/status',
    '/api/v1/appointments/CR-1/status',
    '/api/v1/followups/FW-1/complete',
  ])
})

test('exposes content pipeline list, review, publish and metrics operations', async () => {
  const calls = []
  const client = createApiClient({
    fetchImpl: async (url, init) => {
      calls.push({ url, init })
      return response({ id: 'CF-1', status: '已复盘' })
    },
  })

  await client.listContentItems({ status: '待审核', owner: '林编辑', publishedAt: '2026-07-18' })
  await client.getContentItem('CF-1')
  await client.listContentEvents('CF-1')
  await client.createContentItem({ title: '城市夜行', channel: '短视频', owner: '林编辑', plannedAt: '2026-07-18' })
  await client.saveContentScript('CF-1', { body: '脚本' })
  await client.submitContentReview('CF-1', '主编')
  await client.publishContent('CF-1', { publishedAt: '2026-07-18T18:00:00+08:00', actor: '主编' })
  await client.recordContentMetrics('CF-1', { views: 100, likes: 12, comments: 3, shares: 4 })

  assert.deepEqual(calls.map(({ url }) => url), [
    '/api/v1/content-items?status=%E5%BE%85%E5%AE%A1%E6%A0%B8&owner=%E6%9E%97%E7%BC%96%E8%BE%91&publishedAt=2026-07-18',
    '/api/v1/content-items/CF-1',
    '/api/v1/content-items/CF-1/events',
    '/api/v1/content-items',
    '/api/v1/content-items/CF-1/script',
    '/api/v1/content-items/CF-1/submit-review',
    '/api/v1/content-items/CF-1/publish',
    '/api/v1/content-items/CF-1/metrics',
  ])
  for (const call of calls.filter(({ init }) => init.method === 'POST')) {
    assert.match(call.init.headers['Idempotency-Key'], /^cf-/)
  }
})
