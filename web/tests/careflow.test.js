import test from 'node:test'; import assert from 'node:assert/strict'; import { readFile } from 'node:fs/promises'
test('CreatorFlow has content queue, schedule and review data', async()=>{const source=await readFile(new URL('../src/main.js',import.meta.url),'utf8'); assert.match(source,/今日选题队列/); assert.match(source,/编辑排班/); assert.match(source,/复盘任务/); assert.match(source,/CR-0716-082/)})

test('CreatorFlow binds real API actions while keeping a demo fallback', async()=>{const source=await readFile(new URL('../src/main.js',import.meta.url),'utf8'); assert.match(source,/createApiClient/); assert.match(source,/data-action="checkin"/); assert.match(source,/data-action="status"/); assert.match(source,/data-action="complete-followup"/); assert.match(source,/refreshFromApi/); assert.match(source,/演示数据/)})

test('Vite proxies the default API path to the local Go service', async()=>{const source=await readFile(new URL('../vite.config.js',import.meta.url),'utf8'); assert.match(source,/server/); assert.match(source,/proxy/); assert.match(source,/localhost:8080/)})

test('CreatorFlow admin exposes Kanban filters, review queue, publish form and metrics timeline', async()=>{
  const source=await readFile(new URL('../src/main.js',import.meta.url),'utf8')
  const api=await readFile(new URL('../src/api.js',import.meta.url),'utf8')
  assert.match(source,/内容流水线/)
  assert.match(api,/content-items/)
  assert.match(source,/待审核/)
  assert.match(source,/发布确认/)
  assert.match(source,/指标时间线/)
  assert.match(source,/data-content-filter/)
  assert.match(source,/data-content-action/)
  assert.match(source,/refreshContentEvents/)
  assert.match(source,/listContentEvents\(id\)/)
  assert.match(source,/event\.actor/)
})
