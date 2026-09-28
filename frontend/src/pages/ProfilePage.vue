<template>
  <div class="profile-page" v-if="auth.user">
    <!-- 资产卡 -->
    <section class="asset-card">
      <div class="asset-top">
        <div class="asset-user">
          <span class="asset-avatar">{{ auth.user.avatar_emoji }}</span>
          <div>
            <strong>{{ auth.user.nickname }}</strong>
            <small>@{{ auth.user.username }}</small>
          </div>
        </div>
        <button class="btn btn-plain btn-sm" @click="logout">退出登录</button>
      </div>
      <div class="asset-balance">
        <div class="bal-main">
          <span class="bal-label">尝鲜币余额</span>
          <span class="bal-num">{{ auth.user.coins }}</span>
        </div>
        <div class="bal-frozen">
          <span class="bal-label">冻结中</span>
          <span class="bal-frozen-num">{{ auth.user.frozen_coins }}</span>
        </div>
      </div>
      <div class="asset-tip">
        🪙 帮别人代吃被采纳即可赚币 · 发求代吃会冻结悬赏 · 取消心愿全额退回
      </div>
    </section>

    <!-- Tab 切换 -->
    <div class="tabs">
      <button v-for="t in tabs" :key="t.key" :class="{ on: tab === t.key }" @click="switchTab(t.key)">
        {{ t.label }}
      </button>
    </div>

    <!-- 我的流水 -->
    <section v-if="tab === 'txn'">
      <div v-if="!txns.length" class="empty"><span class="emoji">🧾</span>还没有尝鲜币流水</div>
      <div v-for="t in txns" :key="t.id" class="txn-item">
        <span :class="['txn-ico', t.amount > 0 ? 'in' : t.amount < 0 ? 'out' : 'info']">{{ txnIcon(t.type) }}</span>
        <div class="txn-body">
          <strong>{{ txnTypeText(t.type) }}</strong>
          <small>{{ t.remark }}</small>
          <small class="muted">{{ relTime(t.created_at) }} · 余额 {{ t.balance_after }}</small>
        </div>
        <span :class="['txn-amount', t.amount > 0 ? 'plus' : t.amount < 0 ? 'minus' : 'zero']">
          {{ t.amount > 0 ? '+' : '' }}{{ t.amount }}
        </span>
      </div>
    </section>

    <!-- 我发布的 -->
    <section v-if="tab === 'mine'">
      <div v-if="!myWishes.length" class="empty">
        <span class="emoji">🙏</span>
        <p>还没发过求代吃</p>
        <RouterLink to="/" class="btn btn-ghost">去首页发一个</RouterLink>
      </div>
      <div v-for="w in myWishes" :key="w.id" class="mini-card">
        <div class="mini-head">
          <strong>🏮 {{ w.restaurant_name }}</strong>
          <span :class="['tag', 'tag-' + w.status]">{{ statusText(w.status) }}</span>
        </div>
        <p class="mini-menu">{{ w.dish_list }}</p>
        <div class="mini-foot">
          <span>悬赏 <span class="coin">{{ w.reward_coins }}</span></span>
          <span class="muted small">{{ w.response_count }} 人代吃 · {{ relTime(w.created_at) }}</span>
        </div>
      </div>
    </section>

    <!-- 我代吃的 -->
    <section v-if="tab === 'tasted'">
      <div v-if="!myResponses.length" class="empty">
        <span class="emoji">👅</span>
        <p>还没有替别人去吃过</p>
        <RouterLink to="/" class="btn btn-ghost">去 Feed 找找心仪单子</RouterLink>
      </div>
      <div v-for="r in myResponses" :key="r.id" class="mini-card">
        <div class="mini-head">
          <strong>👅 替吃了「{{ r.restaurant_name }}」</strong>
          <span v-if="r.is_adopted" class="tag tag-settled">✓ 已采纳 +{{ r.reward_coins }}</span>
          <span v-else-if="r.wish_status === 'open'" class="tag tag-open">等待采纳</span>
          <span v-else class="tag tag-canceled">未采纳</span>
        </div>
        <p class="mini-comment">“{{ r.comment }}”</p>
        <div class="mini-foot">
          <span>{{ '⭐'.repeat(r.rating) }}</span>
          <span class="muted small">{{ relTime(r.created_at) }}</span>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import http from '../api'
import { auth } from '../store'

const router = useRouter()
const tab = ref('txn')
const tabs = [
  { key: 'txn', label: '🧾 尝鲜币流水' },
  { key: 'mine', label: '🙏 我发布的' },
  { key: 'tasted', label: '👅 我代吃的' }
]
const txns = ref([])
const myWishes = ref([])
const myResponses = ref([])

async function loadAll() {
  const [a, b, c] = await Promise.all([
    http.get('/api/my/transactions'),
    http.get('/api/my/wishes'),
    http.get('/api/my/responses')
  ])
  txns.value = a.data.items
  myWishes.value = b.data.items
  myResponses.value = c.data.items
}

function switchTab(k) {
  tab.value = k
}

function logout() {
  if (!confirm('确定退出登录吗？')) return
  auth.logout()
  router.push('/')
}

function relTime(t) {
  const d = new Date(t)
  const diff = (Date.now() - d.getTime()) / 1000
  if (diff < 60) return '刚刚'
  if (diff < 3600) return Math.floor(diff / 60) + ' 分钟前'
  if (diff < 86400) return Math.floor(diff / 3600) + ' 小时前'
  if (diff < 86400 * 7) return Math.floor(diff / 86400) + ' 天前'
  return `${d.getMonth() + 1}月${d.getDate()}日`
}

const typeMap = {
  register_gift: '注册赠送',
  wish_freeze: '发布心愿冻结',
  append_freeze: '追加悬赏冻结',
  settle_income: '代吃悬赏入账',
  settle_expense: '悬赏结算完成',
  cancel_refund: '取消心愿退回'
}
const txnTypeText = (t) => typeMap[t] || t
function txnIcon(t) {
  return {
    register_gift: '🎁',
    wish_freeze: '🧊',
    append_freeze: '➕',
    settle_income: '💰',
    settle_expense: '✅',
    cancel_refund: '↩️'
  }[t] || '🪙'
}
const statusText = (s) => ({ open: '求代吃中', settled: '已结算', canceled: '已取消' })[s] || s

onMounted(loadAll)
</script>

<style scoped>
.profile-page { padding-top: 18px; }

.asset-card {
  background: linear-gradient(135deg, var(--brand) 0%, #f0743f 100%);
  border-radius: 22px;
  padding: 22px;
  color: #fff;
  box-shadow: 0 14px 30px rgba(232, 85, 45, 0.28);
  margin-bottom: 20px;
}
.asset-top { display: flex; align-items: center; justify-content: space-between; }
.asset-user { display: flex; align-items: center; gap: 12px; }
.asset-avatar {
  width: 50px; height: 50px; border-radius: 50%;
  background: rgba(255, 255, 255, 0.22);
  display: flex; align-items: center; justify-content: center;
  font-size: 27px;
}
.asset-user strong { display: block; font-size: 17px; }
.asset-user small { opacity: 0.85; font-size: 12px; }
.asset-card .btn-plain {
  background: rgba(255, 255, 255, 0.18);
  border-color: rgba(255, 255, 255, 0.35);
  color: #fff;
}
.asset-balance { display: flex; align-items: flex-end; gap: 30px; margin: 22px 0 12px; }
.bal-label { display: block; font-size: 12.5px; opacity: 0.9; margin-bottom: 2px; }
.bal-num { font-size: 42px; font-weight: 800; line-height: 1; }
.bal-frozen-num { font-size: 24px; font-weight: 700; opacity: 0.95; }
.asset-tip {
  font-size: 12px;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 10px;
  padding: 8px 12px;
  opacity: 0.95;
}

.tabs { display: flex; gap: 8px; margin-bottom: 16px; overflow-x: auto; }
.tabs button {
  border: 1.5px solid var(--line);
  background: #fff;
  border-radius: 999px;
  padding: 8px 15px;
  font-size: 13.5px;
  color: var(--ink-soft);
  cursor: pointer;
  white-space: nowrap;
}
.tabs button.on { background: var(--ink); border-color: var(--ink); color: #fff; font-weight: 700; }

/* 流水 */
.txn-item {
  display: flex;
  align-items: center;
  gap: 12px;
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 15px;
  padding: 13px 15px;
  margin-bottom: 10px;
}
.txn-ico {
  width: 40px; height: 40px; flex-shrink: 0;
  border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  font-size: 19px;
  background: #f6ede5;
}
.txn-ico.in { background: var(--ok-soft); }
.txn-ico.out { background: var(--brand-soft); }
.txn-body { flex: 1; display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.txn-body strong { font-size: 14px; }
.txn-body small { font-size: 12px; color: var(--ink-soft); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.txn-amount { font-weight: 800; font-size: 17px; flex-shrink: 0; }
.txn-amount.plus { color: var(--ok); }
.txn-amount.minus { color: var(--brand-dark); }
.txn-amount.zero { color: #b3a597; font-size: 14px; }

/* 迷你卡 */
.mini-card {
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 15px;
  padding: 14px 16px;
  margin-bottom: 10px;
}
.mini-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: 6px; }
.mini-menu { margin: 0 0 8px; font-size: 13.5px; color: var(--ink-soft); }
.mini-comment { margin: 0 0 8px; font-size: 13.5px; line-height: 1.6; color: #4a3c30; }
.mini-foot { display: flex; align-items: center; justify-content: space-between; font-size: 13px; }
</style>
