const BASE = 'http://127.0.0.1:8081/api'

let pass = 0, fail = 0
function check(name, cond, extra = '') {
  if (cond) { pass++; console.log(`  ✅ ${name}`) }
  else { fail++; console.log(`  ❌ ${name} ${extra}`) }
}

async function api(path, { method = 'GET', token, body } = {}) {
  const res = await fetch(BASE + path, {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    },
    body: body ? JSON.stringify(body) : undefined
  })
  const data = await res.json().catch(() => ({}))
  return { status: res.status, data }
}

const rnd = Math.floor(Math.random() * 1e6)
const nameA = `alice_${rnd}`, nameB = `bob_${rnd}`

console.log('1) 注册赠送初始尝鲜币')
let r = await api('/register', { method: 'POST', body: { username: nameA, password: '1234', nickname: '馋猫阿粒' } })
check('A 注册成功并赠送 100 币', r.status === 200 && r.data.user.coin_balance === 100, JSON.stringify(r.data))
const tokenA = r.data.token

r = await api('/register', { method: 'POST', body: { username: nameB, password: '1234', nickname: '代吃波波' } })
check('B 注册成功并赠送 100 币', r.status === 200 && r.data.user.coin_balance === 100)
const tokenB = r.data.token

r = await api('/register', { method: 'POST', body: { username: nameA, password: '1234' } })
check('重复用户名被拒绝', r.status === 409)

console.log('2) 发布心愿冻结悬赏')
r = await api('/wishes', { method: 'POST', token: tokenA, body: {
  restaurant: '巷口老陈麻辣烫', dish: '招牌毛血旺', reason: '刷到三次了，想知道到底多辣', address: '城南巷 8 号', bounty: 30
}})
check('发布成功', r.status === 200, JSON.stringify(r.data))
const wishId = r.data.wish.id
check('心愿状态为 open / 悬赏 30', r.data.wish.status === 'open' && r.data.wish.bounty === 30)

r = await api('/me', { token: tokenA })
check('A 余额 70、冻结 30', r.data.user.coin_balance === 70 && r.data.user.frozen_balance === 30,
  `got ${r.data.user.coin_balance}/${r.data.user.frozen_balance}`)

console.log('3) 余额不足时发布被拒绝')
r = await api('/wishes', { method: 'POST', token: tokenA, body: {
  restaurant: '贵价店', dish: '龙', reason: 'x', bounty: 100
}})
check('返回 402 且余额不变', r.status === 402)
r = await api('/me', { token: tokenA })
check('A 余额仍为 70 / 冻结 30', r.data.user.coin_balance === 70 && r.data.user.frozen_balance === 30)

console.log('4) 追加悬赏')
r = await api(`/wishes/${wishId}/bounty`, { method: 'POST', token: tokenA, body: { amount: 10 } })
check('追加成功，悬赏变 40', r.status === 200 && r.data.wish.bounty === 40, JSON.stringify(r.data))
r = await api('/me', { token: tokenA })
check('A 余额 60、冻结 40', r.data.user.coin_balance === 60 && r.data.user.frozen_balance === 40)

console.log('5) Feed 流与详情')
r = await api('/wishes')
check('Feed 包含新心愿且带发布人', r.data.list.some(w => w.id === wishId && w.poster.nickname === '馋猫阿粒'))
r = await api(`/wishes/${wishId}`)
check('详情可取', r.status === 200 && r.data.wish.restaurant === '巷口老陈麻辣烫')

console.log('6) 替 TA 去吃')
r = await api(`/wishes/${wishId}/responses`, { method: 'POST', token: tokenB, body: {
  photos: ['/uploads/demo1.jpg', '/uploads/demo2.jpg'], review: '麻辣鲜香，鸭血很嫩，排队 20 分钟值得，分量足！', rating: 5
}})
check('B 提交回应成功', r.status === 200, JSON.stringify(r.data))
const respId = r.data.response.id

r = await api(`/wishes/${wishId}/responses`, { method: 'POST', token: tokenB, body: {
  photos: ['/uploads/x.jpg'], review: '再次回应', rating: 3
}})
check('同一人不能重复回应', r.status === 400)

r = await api(`/wishes/${wishId}/responses`, { method: 'POST', token: tokenA, body: {
  photos: ['/uploads/x.jpg'], review: '自己回应', rating: 3
}})
check('不能替自己代吃', r.status === 400)

r = await api(`/wishes/${wishId}`)
check('详情能看到回应与代吃人', r.data.responses.length >= 1 && r.data.responses[0].taster.nickname === '代吃波波')

console.log('7) 确认采纳并结算（事务）')
r = await api(`/wishes/${wishId}/accept/${respId}`, { method: 'POST', token: tokenA })
check('采纳成功', r.status === 200, JSON.stringify(r.data))
r = await api('/me', { token: tokenA })
check('A 余额 60、冻结清零', r.data.user.coin_balance === 60 && r.data.user.frozen_balance === 0,
  `got ${r.data.user.coin_balance}/${r.data.user.frozen_balance}`)
r = await api('/me', { token: tokenB })
check('B 余额变为 140（100+40）', r.data.user.coin_balance === 140, `got ${r.data.user.coin_balance}`)

r = await api(`/wishes/${wishId}`)
check('心愿变为 settled 且绑定采纳回应', r.data.wish.status === 'settled' && r.data.wish.accepted_response_id === respId)

r = await api(`/wishes/${wishId}/accept/${respId}`, { method: 'POST', token: tokenA })
check('不能重复采纳', r.status === 400)

console.log('8) 取消心愿退回冻结币')
r = await api('/wishes', { method: 'POST', token: tokenA, body: {
  restaurant: '测试退回店', dish: '退退面', reason: 'cancel path', bounty: 20
}})
const wish2 = r.data.wish.id
r = await api(`/wishes/${wish2}/cancel`, { method: 'POST', token: tokenA })
check('无回应时可取消', r.status === 200, JSON.stringify(r.data))
r = await api('/me', { token: tokenA })
check('取消后冻结币退回（余额 60、冻结 0）', r.data.user.coin_balance === 60 && r.data.user.frozen_balance === 0,
  `got ${r.data.user.coin_balance}/${r.data.user.frozen_balance}`)

r = await api('/wishes', { method: 'POST', token: tokenA, body: {
  restaurant: '有回应店', dish: '不能取消', reason: 'x', bounty: 10
}})
const wish3 = r.data.wish.id
await api(`/wishes/${wish3}/responses`, { method: 'POST', token: tokenB, body: { photos: ['/uploads/y.jpg'], review: '吃了', rating: 4 } })
r = await api(`/wishes/${wish3}/cancel`, { method: 'POST', token: tokenA })
check('已有回应时不能取消', r.status === 400)

console.log('9) 个人中心记录')
r = await api('/my/coins', { token: tokenA })
const types = r.data.list.map(l => l.type)
check('流水含 register/freeze/unfreeze/settle_spend',
  ['register','freeze','unfreeze','settle_spend'].every(t => types.includes(t)), JSON.stringify(types))
check('每条流水都有变动后余额快照', r.data.list.every(l => typeof l.balance_after === 'number'))

r = await api('/my/wishes', { token: tokenA })
check('我的心愿 3 条', r.data.list.length === 3, `got ${r.data.list.length}`)

r = await api('/my/responses', { token: tokenB })
check('B 的代吃记录 2 条', r.data.list.length === 2, `got ${r.data.list.length}`)
check('代吃记录带心愿信息', r.data.list[0].wish && r.data.list[0].wish.restaurant)

r = await api('/wishes?mine=1', { token: tokenB })
check('Feed mine 过滤只返回自己的', r.data.list.every(w => w.mine))

console.log(`\n结果: ${pass} 通过, ${fail} 失败`)
process.exit(fail ? 1 : 0)
