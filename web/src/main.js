import './styles.css'
import './content-workspace.css'
import { createApiClient } from './api.js'

const api = createApiClient()

const demoAppointments = [
  { id: 'CR-0716-082', patient: '选题《城市夜行》', department: '短视频', doctor: '林编辑', scheduledAt: '2026-07-16T09:30:00+08:00', status: '待制作' },
  { id: 'CR-0716-081', patient: '选题《一周好物》', department: '图文专栏', doctor: '沈编辑', scheduledAt: '2026-07-16T09:45:00+08:00', status: '已排期' },
  { id: 'CR-0716-080', patient: '选题《品牌访谈》', department: '直播栏目', doctor: '赵编辑', scheduledAt: '2026-07-16T10:00:00+08:00', status: '已发布' },
  { id: 'CR-0716-079', patient: '选题《城市观察》', department: '短视频', doctor: '林编辑', scheduledAt: '2026-07-16T10:15:00+08:00', status: '待排期' },
  { id: 'CR-0716-078', patient: '选题《品牌故事》', department: '品牌合作', doctor: '周编辑', scheduledAt: '2026-07-16T10:30:00+08:00', status: '待排期' },
]

const demoFollowups = [
  { id: 'RV-0716-012', patient: '选题《城市夜行》', summary: '标题与封面复核', dueAt: '今天 16:00', status: '待完成' },
  { id: 'RV-0716-011', patient: '选题《一周好物》', summary: '素材版权检查', dueAt: '今天 17:30', status: '待完成' },
  { id: 'RV-0716-010', patient: '选题《品牌访谈》', summary: '发布数据复盘', dueAt: '明天 09:30', status: '待完成' },
  { id: 'RV-0715-009', patient: '选题《城市观察》', summary: '评论区复盘记录', dueAt: '已完成', status: '已完成' },
]

const demoContentItems = [
  { id: 'CF-0718-001', title: '城市夜行：下班后的十五分钟', channel: '短视频', owner: '林编辑', plannedAt: '2026-07-18T09:00:00+08:00', status: '待选题' },
  { id: 'CF-0718-002', title: '一周好物：把桌面整理成工作流', channel: '图文专栏', owner: '沈编辑', plannedAt: '2026-07-18T10:00:00+08:00', status: '写作中', script: { body: '开场钩子、三段主体和结尾行动号召。' } },
  { id: 'CF-0718-003', title: '品牌访谈：小店如何留住老客', channel: '直播栏目', owner: '赵编辑', plannedAt: '2026-07-18T14:00:00+08:00', status: '制作中', script: { body: '开场钩子、三段主体和结尾行动号召。' } },
  { id: 'CF-0718-004', title: '夏日直播：创作者增长公开课', channel: '品牌合作', owner: '周编辑', plannedAt: '2026-07-18T16:00:00+08:00', status: '待审核', script: { body: '公开课流程、嘉宾串词与观众互动问题。' } },
  { id: 'CF-0718-005', title: '通勤装备：轻量化出行清单', channel: '短视频', owner: '林编辑', plannedAt: '2026-07-18T18:00:00+08:00', status: '已发布', publish: { publishedAt: '2026-07-18T18:00:00+08:00', actor: '主编' }, metrics: { views: 18200, likes: 920, comments: 61, shares: 140 } },
  { id: 'CF-0718-006', title: '一张图读懂内容复盘', channel: '图文专栏', owner: '沈编辑', plannedAt: '2026-07-19T09:00:00+08:00', status: '已复盘', publish: { publishedAt: '2026-07-18T18:00:00+08:00', actor: '主编' }, metrics: { views: 12480, likes: 892, comments: 67, shares: 141 } },
]

const demoDashboard = { todayAppointments: 86, averageWaitMinutes: 12, completed: 58, checkedIn: 42, pendingFollowups: 12 }
const statusColors = { 待排期: 'coral', 已排期: 'indigo', 待制作: 'amber', 制作中: 'green', 已发布: 'green', 已取消: 'gray' }
const nav = [
  ['overview', '运营总览', '⌂'],
  ['content', '内容流水线', '▦'],
  ['queue', '选题队列', '▤'],
  ['doctors', '编辑排班', '◉'],
  ['patients', '创作者档案', '♧'],
  ['followups', '复盘任务', '✓'],
  ['mobile', '移动端体验', '⌁'],
]

let appointments = demoAppointments.map((item) => ({ ...item }))
let followupTasks = demoFollowups.map((item) => ({ ...item }))
let dashboard = { ...demoDashboard }
let page = 'overview'
let toast = ''
let toastTimer
let dataSource = '演示数据'
let isSyncing = false
let contentItems = demoContentItems.map((item) => ({ ...item }))
let selectedContentId = contentItems[2].id
let contentStatusFilter = ''
let contentOwnerFilter = ''
let contentKeyword = ''
let contentDraft = { script: '', publishedAt: '2026-07-18T18:00:00+08:00', actor: '主编', views: 0, likes: 0, comments: 0, shares: 0 }

function timeLabel(value) {
  const match = String(value ?? '').match(/T(\d{2}:\d{2})/)
  return match?.[1] || String(value ?? '').slice(0, 5) || '--:--'
}

function normalizeAppointment(item) {
  return {
    id: item.id,
    patientId: item.patientId,
    patient: item.patient || '未命名创作者',
    department: item.department || '待分诊',
    doctor: item.doctor || '待安排',
    scheduledAt: item.scheduledAt || '',
    status: item.status || '待排期',
  }
}

function normalizeFollowup(item) {
  return {
    id: item.id,
    patientId: item.patientId,
    patient: item.patient || '未命名创作者',
    summary: item.summary || '内容复盘任务',
    dueAt: item.dueAt || '--',
    status: item.status || '待完成',
  }
}

function showToast(message) {
  toast = message
  render()
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toast = ''
    render()
  }, 2200)
}

function appointmentAction(appointment) {
  if (appointment.status === '待排期') return `<button class="text-action" data-action="checkin" data-appointment-id="${appointment.id}">加入排期</button>`
  if (appointment.status === '已排期') return `<button class="text-action" data-action="status" data-next-status="待制作" data-appointment-id="${appointment.id}">确认排期</button>`
  if (appointment.status === '待制作') return `<button class="text-action" data-action="status" data-next-status="制作中" data-appointment-id="${appointment.id}">开始制作</button>`
  if (appointment.status === '制作中') return `<button class="text-action" data-action="status" data-next-status="已发布" data-appointment-id="${appointment.id}">发布内容</button>`
  return '<button class="text-action" data-toast="该选题已完成，无需重复操作">查看详情</button>'
}

function header(title) {
  return `<header><span>工作台　/　<strong>${title}</strong></span><span class="header-tools"><span>2026 年 7 月 16 日</span><span class="data-source ${dataSource === 'API 数据' ? 'remote' : ''}">● ${isSyncing ? '同步中' : dataSource}</span><button class="refresh" data-refresh ${isSyncing ? 'disabled' : ''}>↻ 刷新</button></span></header>`
}

function render() {
  const title = nav.find((item) => item[0] === page)?.[1] || '运营总览'
  const content = page === 'overview' ? overview() : page === 'content' ? contentWorkspace() : page === 'queue' ? queue() : page === 'doctors' ? doctors() : page === 'patients' ? patients() : page === 'followups' ? followups() : mobileView()
  document.querySelector('#app').innerHTML = `<div class="shell"><aside><div class="brand"><span>✚</span><div><strong>CreatorFlow</strong><small>创作运营中心</small></div></div><div class="clinic">● 上海静安联合创作　⌄</div><p class="caption">内容运营</p><nav>${nav.map((item) => `<button class="${page === item[0] ? 'active' : ''}" data-page="${item[0]}"><i>${item[2]}</i>${item[1]}${item[0] === 'queue' ? '<em>8</em>' : ''}</button>`).join('')}</nav><div class="user"><b>许</b><span><strong>许汝林</strong><small>运营管理员</small></span></div></aside><main>${header(title)}<section class="heading"><div><p>THURSDAY, JUL 16 · CREATORFLOW</p><h1>${title} <i>✦</i></h1><label>让每一次选题，都有被照顾的下一步。</label></div><button class="primary" data-action="create-appointment">＋ 新建选题</button></section>${content}<footer>CreatorFlow 内容排期与创作者协同 · 演示数据均为虚构内容，发布与复盘操作可在本地 API 重现</footer><div class="toast" ${toast ? '' : 'hidden'}>${toast}</div></main></div>`
  bind()
}

function overview() {
  return `<section class="metrics"><article class="metric dark"><span>今日选题</span><strong>${dashboard.todayAppointments}</strong><small>↗ 较昨日 +14.6%</small></article><article class="metric"><span>平均制作周转</span><strong>${dashboard.averageWaitMinutes}<small> 小时</small></strong><small class="good">较上周 -3 小时</small></article><article class="metric"><span>今日发布</span><strong>${dashboard.completed}<small> 条</small></strong><div class="progress"><i style="width:68%"></i></div></article><article class="metric warm"><span>待复盘</span><strong>${dashboard.pendingFollowups}<small> 条</small></strong><small class="coral">今日需完成</small></article></section><section class="grid"><article class="panel calendar"><div class="panel-head"><div><h2>今日选题队列</h2><p>7 月 16 日 · 周四 · 共 ${dashboard.todayAppointments} 条内容</p></div><button class="link" data-page="queue">查看队列 →</button></div><div class="timeline">${appointments.slice(0, 4).map((appointment) => `<div class="time-row"><span>${timeLabel(appointment.scheduledAt)}</span><i class="time-dot ${statusColors[appointment.status] || 'indigo'}"></i><div><strong>${appointment.patient}</strong><small>${appointment.department} · ${appointment.status}</small></div><b class="status ${statusColors[appointment.status] || 'indigo'}">${appointment.status}</b></div>`).join('')}</div></article><article class="panel"><div class="panel-head"><div><h2>内容渠道负载</h2><p>当前时段排期利用率</p></div><button class="link" data-page="doctors">排班管理 →</button></div><div class="load-list">${[['短视频', '32 / 40', '80%', 'indigo'], ['图文专栏', '18 / 24', '75%', 'coral'], ['直播栏目', '12 / 18', '67%', 'green'], ['品牌合作', '8 / 12', '66%', 'amber']].map((item) => `<div class="load"><div><strong>${item[0]}</strong><span>${item[1]}</span></div><div class="load-bar"><i class="${item[3]}" style="width:${item[2]}"></i></div><b>${item[2]}</b></div>`).join('')}</div></article></section><section class="grid lower"><article class="panel"><div class="panel-head"><div><h2>复盘完成趋势</h2><p>近 7 日任务完成率</p></div><span class="legend">本周平均 84%</span></div><div class="spark"><i style="height:38%"></i><i style="height:58%"></i><i style="height:46%"></i><i style="height:74%"></i><i style="height:66%"></i><i style="height:88%"></i><i class="today" style="height:80%"></i></div><div class="days"><span>周五</span><span>周六</span><span>周日</span><span>周一</span><span>周二</span><span>周三</span><span>今天</span></div></article><article class="panel tasks"><div class="panel-head"><div><h2>待办提醒</h2><p>需要运营人员跟进的事项</p></div></div><div class="task"><span class="task-icon coral">!</span><div><strong>3 个选题需要补充素材</strong><small>选题队列 · 10 分钟前</small></div><button data-page="queue">处理</button></div><div class="task"><span class="task-icon amber">✓</span><div><strong>${dashboard.pendingFollowups} 条复盘今日到期</strong><small>内容复盘 · 32 分钟前</small></div><button data-page="followups">查看</button></div></article></section>`
}

const contentStatuses = ['待选题', '写作中', '制作中', '待审核', '已发布', '已复盘']

function contentStatusClass(status) {
  return { 待选题: 'coral', 写作中: 'amber', 制作中: 'indigo', 待审核: 'purple', 已发布: 'green', 已复盘: 'blue' }[status] || 'gray'
}

function normalizeContent(item) {
  return {
    ...item,
    title: item.title || '未命名选题',
    channel: item.channel || '待配置渠道',
    owner: item.owner || '待分配负责人',
    plannedAt: item.plannedAt || '',
    status: item.status || '待选题',
    events: Array.isArray(item.events) ? item.events : [],
  }
}

function demoContentEvents(item) {
  const order = contentStatuses.slice(0, contentStatuses.indexOf(item.status) + 1)
  return order.map((status, index) => ({
    id: `${item.id}-event-${index + 1}`,
    fromStatus: index ? order[index - 1] : '',
    toStatus: status,
    action: index === 0 ? 'create' : status === '已发布' ? 'publish' : status === '已复盘' ? 'record_metrics' : 'advance',
    actor: status === '已发布' ? '主编' : item.owner,
    createdAt: `2026-07-1${6 + Math.min(index, 3)}T0${9 + index}:00:00+08:00`,
  }))
}

function contentWorkspace() {
  const filtered = contentItems.filter((item) => {
    const matchesStatus = !contentStatusFilter || item.status === contentStatusFilter
    const matchesOwner = !contentOwnerFilter || item.owner === contentOwnerFilter
    const matchesKeyword = !contentKeyword || `${item.title} ${item.channel} ${item.owner}`.toLowerCase().includes(contentKeyword.toLowerCase())
    return matchesStatus && matchesOwner && matchesKeyword
  })
  const selected = contentItems.find((item) => item.id === selectedContentId) || filtered[0] || contentItems[0]
  const events = (selected?.events?.length ? selected.events : demoContentEvents(selected || { id: 'demo', status: '待选题', owner: '林编辑' })).slice().sort((a, b) => String(a.createdAt).localeCompare(String(b.createdAt)))
  const owners = [...new Set(contentItems.map((item) => item.owner))]
  const detail = selected ? `<article class="content-detail">
    <div class="content-detail-head"><div><span class="status ${contentStatusClass(selected.status)}">${selected.status}</span><h2>${selected.title}</h2><p>${selected.channel} · ${selected.owner} · 计划 ${timeLabel(selected.plannedAt)}</p></div><span class="content-id">${selected.id}</span></div>
    <div class="content-detail-grid"><div><span>脚本草稿</span><strong>${selected.script?.body ? '已填写' : '待补充'}</strong></div><div><span>发布时间</span><strong>${selected.publish?.publishedAt ? timeLabel(selected.publish.publishedAt) : '未发布'}</strong></div><div><span>阅读量</span><strong>${selected.metrics?.views?.toLocaleString?.() || '—'}</strong></div><div><span>互动率</span><strong>${selected.metrics ? `${(selected.metrics.likes + selected.metrics.comments + selected.metrics.shares).toLocaleString()} 次` : '—'}</strong></div></div>
    ${selected.status === '待选题' || selected.status === '写作中' || selected.status === '制作中' ? `<section class="content-form"><h3>脚本编辑</h3><textarea data-content-script placeholder="补充分镜、口播与行动号召">${selected.script?.body || ''}</textarea><div class="form-actions"><button class="primary small" data-content-action="save-script" data-content-id="${selected.id}">保存脚本</button>${selected.status === '制作中' ? `<button class="secondary small" data-content-action="submit-review" data-content-id="${selected.id}">提交审核</button>` : ''}</div></section>` : ''}
    ${selected.status === '待审核' ? `<section class="content-form"><h3>发布确认</h3><div class="form-row"><label>发布时间<input data-content-published-at value="${selected.publish?.publishedAt || contentDraft.publishedAt}" /></label><label>审核人<input data-content-actor value="${contentDraft.actor}" /></label></div><button class="primary small" data-content-action="publish" data-content-id="${selected.id}">确认发布</button></section>` : ''}
    ${selected.status === '已发布' || selected.status === '已复盘' ? `<section class="content-form"><h3>指标复盘</h3><div class="metric-inputs"><label>阅读量<input type="number" min="0" data-content-metric="views" value="${selected.metrics?.views || contentDraft.views}" /></label><label>点赞<input type="number" min="0" data-content-metric="likes" value="${selected.metrics?.likes || contentDraft.likes}" /></label><label>评论<input type="number" min="0" data-content-metric="comments" value="${selected.metrics?.comments || contentDraft.comments}" /></label><label>分享<input type="number" min="0" data-content-metric="shares" value="${selected.metrics?.shares || contentDraft.shares}" /></label></div><button class="primary small" data-content-action="metrics" data-content-id="${selected.id}">记录复盘</button></section>` : ''}
    <section class="content-timeline"><div class="subhead"><h3>指标时间线</h3><span>最近活动</span></div><ol>${events.map((event) => `<li><time>${timeLabel(event.createdAt)}</time><div><strong>${event.toStatus || event.action}</strong><small>${event.action} · ${event.actor || '系统'}</small></div></li>`).join('')}</ol></section>
  </article>` : '<article class="content-detail empty">暂无匹配的内容</article>'
  return `<section class="content-workspace"><div class="content-toolbar"><div><span class="eyebrow">CREATOR PIPELINE</span><h2>内容流水线</h2><p>从选题、脚本、审核到发布复盘，所有动作都留在同一条可追踪时间线上。</p></div><div class="toolbar-actions"><span class="fiction-note">虚构演示数据 · API 可切换</span><button class="primary small" data-content-action="create">＋ 新建选题</button></div></div><div class="content-filters"><input data-content-filter="keyword" placeholder="搜索标题、渠道或负责人" value="${contentKeyword}" /><select data-content-filter="status"><option value="">全部状态</option>${contentStatuses.map((status) => `<option value="${status}" ${contentStatusFilter === status ? 'selected' : ''}>${status}</option>`).join('')}</select><select data-content-filter="owner"><option value="">全部负责人</option>${owners.map((owner) => `<option value="${owner}" ${contentOwnerFilter === owner ? 'selected' : ''}>${owner}</option>`).join('')}</select><span class="filter-count">${filtered.length} 条内容</span></div><div class="content-board">${contentStatuses.map((status) => `<section class="content-column"><div class="column-head"><h3>${status}</h3><span>${filtered.filter((item) => item.status === status).length}</span></div>${filtered.filter((item) => item.status === status).map((item) => `<button class="content-card ${selected?.id === item.id ? 'selected' : ''}" data-content-action="select" data-content-id="${item.id}"><span class="status ${contentStatusClass(item.status)}">${item.status}</span><strong>${item.title}</strong><small>${item.channel} · ${item.owner}</small><time>${timeLabel(item.plannedAt)}</time></button>`).join('') || '<p class="column-empty">暂无内容</p>'}</section>`).join('')}</div>${detail}</section>`
}

function queue() {
  return `<section class="panel full"><div class="panel-head"><div><h2>选题队列</h2><p>${dataSource === 'API 数据' ? 'API 实时选题' : '20 条演示选题'} · 支持排期、制作、发布和复盘</p></div><span class="chip">今天　⌄</span></div><div class="table"><div class="th"><span>选题编号 / 创作者</span><span>内容渠道</span><span>时间</span><span>状态</span><span>操作</span></div>${appointments.concat(dataSource === 'API 数据' ? [] : appointments.slice(0, 3)).map((appointment) => `<div class="tr"><span><strong>${appointment.id}</strong><small>${appointment.patient}</small></span><span>${appointment.department}</span><span>${timeLabel(appointment.scheduledAt)}</span><b class="status ${statusColors[appointment.status] || 'indigo'}">${appointment.status}</b><span>${appointmentAction(appointment)}</span></div>`).join('')}</div></section>`
}

function doctors() {
  return `<section class="panel full"><div class="panel-head"><div><h2>编辑排班</h2><p>8 位编辑 · 今日 42 个可选题时段</p></div><button class="primary small" data-toast="排班编辑器已打开">编辑排班</button></div><div class="doctor-grid">${[['林编辑', '短视频', '32 条制作中', 'indigo'], ['沈编辑', '图文专栏', '18 条待复核', 'coral'], ['赵编辑', '直播栏目', '制作中', 'green'], ['周编辑', '品牌合作', '8 条排期中', 'amber'], ['陈编辑', '短视频', '午间休息', 'gray'], ['王编辑', '图文专栏', '6 条排期中', 'indigo']].map((doctor) => `<article><div class="doctor-avatar ${doctor[3]}">${doctor[0][0]}</div><div><strong>${doctor[0]}</strong><small>${doctor[1]}</small></div><span>${doctor[2]}</span><div class="schedule-line"><i style="width:78%"></i></div></article>`).join('')}</div></section>`
}

function patients() {
  return `<section class="panel full"><div class="panel-head"><div><h2>创作者档案</h2><p>30 条虚构档案 · 仅用于界面演示</p></div><button class="link" data-toast="导出任务已创建">导出列表 ↓</button></div><div class="table"><div class="th"><span>创作者 / 编号</span><span>内容渠道</span><span>最近创作</span><span>复盘状态</span><span>操作</span></div>${[['林晓雨', 'CR-2038', '短视频', '07/16', '待复盘'], ['沈明远', 'CR-2037', '图文专栏', '07/15', '进行中'], ['赵思涵', 'CR-2036', '直播栏目', '07/14', '已完成'], ['周子昂', 'CR-2035', '短视频', '07/13', '待复盘'], ['许安然', 'CR-2034', '品牌合作', '07/12', '已完成']].map((creator) => `<div class="tr"><span><strong>${creator[0]}</strong><small>${creator[1]}</small></span><span>${creator[2]}</span><span>${creator[3]}</span><b class="status ${creator[4] === '已完成' ? 'green' : 'coral'}">${creator[4]}</b><button class="text-action" data-toast="${creator[0]} 档案已打开">查看档案</button></div>`).join('')}</div></section>`
}

function followups() {
  return `<section class="panel full"><div class="panel-head"><div><h2>复盘任务</h2><p>${dataSource === 'API 数据' ? 'API 实时复盘' : '12 条待跟进任务'} · 由主编/运营确认后记录</p></div><span class="chip">全部任务　⌄</span></div><div class="follow-list">${followupTasks.map((item) => `<article><span class="task-icon ${item.status === '已完成' ? 'green' : 'coral'}">✓</span><div><strong>${item.id} · ${item.patient}</strong><p>${item.summary}</p><small>${item.dueAt} · ${dataSource === 'API 数据' ? 'API 数据' : '演示任务'}</small></div>${item.status === '已完成' ? '<button class="text-action" data-toast="该复盘已经完成">查看</button>' : `<button class="text-action" data-action="complete-followup" data-followup-id="${item.id}">完成任务</button>`}</article>`).join('')}</div></section>`
}

function mobileView() {
  return `<section class="mobile-panel"><div class="mobile-panel__hero"><span>CREATORFLOW MOBILE</span><h2>我的创作与复盘</h2><p>创作者端可在同一套闭环 API 中完成排期确认、制作和复盘确认。</p><button class="primary" data-action="create-appointment">＋ 创建演示选题</button></div><div class="mobile-list"><h3>今日选题</h3>${appointments.slice(0, 4).map((appointment) => `<article class="mobile-card"><div><small>${timeLabel(appointment.scheduledAt)} · ${appointment.department}</small><strong>${appointment.patient}</strong><span>${appointment.doctor} · ${appointment.status}</span></div><b class="status ${statusColors[appointment.status] || 'indigo'}">${appointment.status}</b>${appointmentAction(appointment)}</article>`).join('')}</div><div class="mobile-list"><h3>我的复盘</h3>${followupTasks.slice(0, 3).map((item) => `<article class="mobile-card"><div><small>${item.dueAt}</small><strong>${item.summary}</strong><span>${item.patient} · ${item.status}</span></div>${item.status === '已完成' ? '<b class="status green">已完成</b>' : `<button class="text-action" data-action="complete-followup" data-followup-id="${item.id}">完成复盘</button>`}</article>`).join('')}</div></section>`
}

async function refreshFromApi({ quiet = false } = {}) {
  if (isSyncing) return
  isSyncing = true
  render()
  try {
    const [nextDashboard, nextAppointments, nextFollowups, nextContent] = await Promise.all([
      api.getDashboard(),
      api.listAppointments({ page: 1, pageSize: 20 }),
      api.listFollowups({ page: 1, pageSize: 20 }),
      api.listContentItems({ page: 1, pageSize: 100 }),
    ])
    dashboard = { ...demoDashboard, ...nextDashboard }
    appointments = (nextAppointments?.list || []).map(normalizeAppointment)
    followupTasks = (nextFollowups?.list || []).map(normalizeFollowup)
    if (Array.isArray(nextContent?.list) && nextContent.list.length) {
      contentItems = nextContent.list.map(normalizeContent)
      selectedContentId = contentItems.find((item) => item.id === selectedContentId)?.id || contentItems[0].id
    }
    dataSource = 'API 数据'
    if (!quiet) toast = '已从 CreatorFlow API 刷新数据'
  } catch (error) {
    dataSource = '演示数据'
    if (!quiet) toast = `API 暂不可用，继续使用演示数据：${error.message}`
  } finally {
    isSyncing = false
    render()
  }
}

function replaceContent(updated) {
  const normalized = normalizeContent(updated)
  contentItems = contentItems.map((item) => item.id === normalized.id ? { ...item, ...normalized } : item)
}

async function selectContent(id) {
  selectedContentId = id
  const local = contentItems.find((item) => item.id === id)
  if (!local) return render()
  render()
  try {
    const detail = await api.getContentItem(id)
    replaceContent(detail)
    await refreshContentEvents(id)
  } catch {
    // Keep deterministic demo detail when the API is unavailable.
  }
  render()
}

async function refreshContentEvents(id) {
  const result = await api.listContentEvents(id)
  const events = Array.isArray(result?.list) ? result.list : []
  contentItems = contentItems.map((item) => item.id === id ? { ...item, events } : item)
  return events
}

async function createContentItem() {
  const input = { title: '城市夜行：夜班人的补给站', channel: '短视频', owner: '林编辑', plannedAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString() }
  try {
    const created = await api.createContentItem(input)
    contentItems = [normalizeContent(created), ...contentItems]
    selectedContentId = created.id
    dataSource = 'API 数据'
    showToast('选题已创建，进入写作队列')
  } catch (error) {
    const local = normalizeContent({ ...input, id: `CF-DEMO-${Date.now().toString().slice(-6)}`, status: '待选题' })
    contentItems = [local, ...contentItems]
    selectedContentId = local.id
    showToast(`接口暂不可用，保留演示选题：${error.message}`)
  }
}

async function contentAction(element) {
  const action = element.dataset.contentAction
  const id = element.dataset.contentId
  if (action === 'select') return selectContent(id)
  if (action === 'create') return createContentItem()
  const item = contentItems.find((content) => content.id === id)
  if (!item) return
  try {
    if (action === 'save-script') {
      const body = document.querySelector('[data-content-script]')?.value?.trim()
      if (!body) return showToast('请先填写脚本内容')
      replaceContent(await api.saveContentScript(id, { body }))
      await refreshContentEvents(id)
      showToast('脚本已保存，状态已推进')
    } else if (action === 'submit-review') {
      replaceContent(await api.submitContentReview(id, '主编'))
      await refreshContentEvents(id)
      showToast('已提交审核队列')
    } else if (action === 'publish') {
      const publishedAt = document.querySelector('[data-content-published-at]')?.value?.trim()
      const actor = document.querySelector('[data-content-actor]')?.value?.trim()
      if (!publishedAt || !actor) return showToast('发布时间和审核人不能为空')
      replaceContent(await api.publishContent(id, { publishedAt, actor }))
      await refreshContentEvents(id)
      showToast('内容已发布，等待数据复盘')
    } else if (action === 'metrics') {
      const values = Object.fromEntries([...document.querySelectorAll('[data-content-metric]')].map((input) => [input.dataset.contentMetric, Number(input.value)]))
      if (Object.values(values).some((value) => !Number.isFinite(value) || value < 0)) return showToast('指标不能为负数')
      replaceContent(await api.recordContentMetrics(id, values))
      await refreshContentEvents(id)
      showToast('复盘指标已记录')
    }
    dataSource = 'API 数据'
  } catch (error) {
    showToast(`接口暂不可用：${error.message}`)
  }
  render()
}

function replaceAppointment(updated) {
  appointments = appointments.map((item) => item.id === updated.id ? normalizeAppointment(updated) : item)
}

async function advanceAppointment(button) {
  const id = button.dataset.appointmentId
  const appointment = appointments.find((item) => item.id === id)
  if (!appointment) return
  const nextStatus = button.dataset.nextStatus
  try {
    const updated = button.dataset.action === 'checkin'
      ? await api.checkinAppointment(id)
      : await api.updateAppointmentStatus(id, nextStatus, '运营人员')
    replaceAppointment(updated)
    dataSource = 'API 数据'
    showToast(`${appointment.patient} 已更新为${updated.status}`)
  } catch (error) {
    dataSource = '演示数据'
    showToast(`接口暂不可用，已保留演示数据：${error.message}`)
  }
}

async function completeFollowup(button) {
  const id = button.dataset.followupId
  const task = followupTasks.find((item) => item.id === id)
  if (!task) return
  try {
    const updated = await api.completeFollowup(id)
    followupTasks = followupTasks.map((item) => item.id === id ? normalizeFollowup(updated) : item)
    dataSource = 'API 数据'
    showToast(`${task.patient} 的复盘已完成`)
  } catch (error) {
    dataSource = '演示数据'
    showToast(`接口暂不可用，已保留演示任务：${error.message}`)
  }
}

async function createAppointment() {
  try {
    const created = await api.createAppointment({ patient: '移动端演示选题', patientId: 'CR-MOBILE-DEMO', department: '短视频', doctor: '林编辑', scheduledAt: new Date().toISOString() })
    appointments = [normalizeAppointment(created), ...appointments]
    dataSource = 'API 数据'
    showToast('选题已创建，可继续在移动端完成排期')
  } catch (error) {
    dataSource = '演示数据'
    showToast(`API 暂不可用，保留演示选题：${error.message}`)
  }
}

function bind() {
  document.querySelectorAll('[data-page]').forEach((element) => element.addEventListener('click', () => {
    page = element.dataset.page
    render()
  }))
  document.querySelectorAll('[data-toast]').forEach((element) => element.addEventListener('click', () => showToast(element.dataset.toast)))
  document.querySelectorAll('[data-refresh]').forEach((element) => element.addEventListener('click', () => refreshFromApi()))
  document.querySelectorAll('[data-content-filter]').forEach((element) => element.addEventListener('input', () => {
    if (element.dataset.contentFilter === 'keyword') contentKeyword = element.value
    if (element.dataset.contentFilter === 'status') contentStatusFilter = element.value
    if (element.dataset.contentFilter === 'owner') contentOwnerFilter = element.value
    render()
  }))
  document.querySelectorAll('[data-content-action]').forEach((element) => element.addEventListener('click', () => contentAction(element)))
  document.querySelectorAll('[data-action]').forEach((element) => element.addEventListener('click', () => {
    if (element.dataset.action === 'checkin' || element.dataset.action === 'status') return advanceAppointment(element)
    if (element.dataset.action === 'complete-followup') return completeFollowup(element)
    if (element.dataset.action === 'create-appointment') return createAppointment()
    return undefined
  }))
}

render()
refreshFromApi({ quiet: true })
