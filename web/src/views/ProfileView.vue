<template>
  <div>
    <!-- 尝鲜币钱包 -->
    <section class="wallet">
      <div class="wallet-top">
        <div class="user">
          <span class="big-avatar">{{ (auth.user?.nickname || '吃')[0] }}</span>
          <div>
            <div class="nick">{{ auth.user?.nickname }}</div>
            <div class="username">@{{ auth.user?.username }}</div>
          </div>
        </div>
        <button class="btn btn-outline btn-sm" @click="loadAll">刷新</button>
      </div>
      <div class="balances">
        <div class="bal">
          <div class="bal-num">🪙 {{ auth.user?.coin_balance ?? 0 }}</div>
          <div class="bal-label">可用尝鲜币</div>
        </div>
        <div class="bal">
          <div class="bal-num">❄️ {{ auth.user?.frozen_balance ?? 0 }}</div>
          <div class="bal-label">悬赏冻结中</div>
        </div>
      </div>
    </section>

    <!-- Tab 切换 -->
    <div class="tabs">
      <button v-for="t in tabs" :key="t.key" :class="{ active: tab === t.key }" @click="tab = t.key">
        {{ t.label }}
      </button>
    </div>

    <!-- 尝鲜币流水 -->
    <div v-if="tab === 'coins'">
      <div v-if="loadingCoins" class="empty">加载中…</div>
      <div v-else-if="coins.length === 0" class="empty"><span class="emoji">🪙</span>还没有尝鲜币记录</div>
      <div v-for="log in coins" :key="log.id" class="coin-row">
        <span class="coin-icon">{{ typeMeta(log.type).icon }}</span>
        <div class="coin-main">
          <div class="coin-title">{{ typeMeta(log.type).text }}{{ log.remark ? ' · ' + log.remark : '' }}</div>
          <div class="coin-time">{{ formatTime(log.created_at) }}</div>
        </div>
        <div class="coin-amount" :class="log.amount >= 0 ? 'plus' : 'minus'">
          {{ log.amount >= 0 ? '+' : '' }}{{ log.amount }}
          <div class="coin-bal">可用 {{ log.balance_after }} · 冻结 {{ log.frozen_after }}</div>
        </div>
      </div>
    </div>

    <!-- 我发布的心愿 -->
    <div v-if="tab === 'wishes'">
      <div v-if="wishes.length === 0" class="empty"><span class="emoji">📝</span>还没有发布过求代吃心愿</div>
      <div v-else class="feed">
        <WishPostcard
          v-for="w in wishes"
          :key="w.id"
          :wish="w"
          :expanded="expandedId === w.id"
          @toggle-expand="expandedId = expandedId === w.id ? null : w.id"
          @addBounty="onAddBounty"
          @cancel="onCancel"
          @viewAll="openDetail"
          @preview="(u) => (preview = u)"
        />
      </div>
    </div>

    <!-- 我替别人去吃 -->
    <div v-if="tab === 'responses'">
      <div v-if="responses.length === 0" class="empty"><span class="emoji">🥢</span>还没有替谁去吃过，去广场逛逛吧</div>
      <div v-for="r in responses" :key="r.id" class="resp-card">
        <div class="resp-top">
          <b>🍽️ {{ r.wish?.restaurant }} · {{ r.wish?.dish }}</b>
          <span :class="['badge', wishBadge(r.wish?.status)]">{{ wishStatus(r.wish?.status, r) }}</span>
        </div>
        <div class="stars">{{ '★'.repeat(r.rating) }}{{ '☆'.repeat(5 - r.rating) }}</div>
        <p class="resp-text">{{ r.review }}</p>
        <div class="resp-imgs">
          <img v-for="(p, i) in r.photos" :key="i" :src="p" @click="preview = p" />
        </div>
        <div class="resp-foot">{{ formatTime(r.created_at) }} · 发起人 @{{ r.wish?.poster?.nickname }}</div>
      </div>
    </div>

    <WishDetailModal v-if="detailWishId" :wish-id="detailWishId" @close="detailWishId = null" @accepted="loadAll" @preview="(u) => (preview = u)" />
    <div v-if="preview" class="lightbox" @click="preview = ''"><img :src="preview" /></div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { myCoins, myWishes, myResponses, addBounty, cancelWish, getMe } from '../api'
import { useAuth } from '../store/auth'
import { toast } from '../store/toast'
import WishPostcard from '../components/WishPostcard.vue'
import WishDetailModal from '../components/WishDetailModal.vue'

const auth = useAuth()
const tab = ref('coins')
const tabs = [
  { key: 'coins', label: '尝鲜币流水' },
  { key: 'wishes', label: '我发布的心愿' },
  { key: 'responses', label: '我替吃的记录' }
]

const coins = ref([])
const wishes = ref([])
const responses = ref([])
const loadingCoins = ref(false)
const expandedId = ref(null)
const detailWishId = ref(null)
const preview = ref('')

const TYPE_MAP = {
  register: { icon: '🎁', text: '注册赠送' },
  freeze: { icon: '❄️', text: '悬赏冻结' },
  unfreeze: { icon: '↩️', text: '取消退回' },
  settle_spend: { icon: '💸', text: '采纳支出' },
  settle_income: { icon: '🎉', text: '代吃收入' }
}
function typeMeta(t) {
  return TYPE_MAP[t] || { icon: '🪙', text: t }
}
function wishStatus(s, r) {
  if (s === 'settled' && r.wish?.accepted_response_id === r.id) return '已采纳 · 已入账'
  if (s === 'settled') return '已采纳他人'
  if (s === 'canceled') return '心愿已取消'
  return '等待采纳'
}
function wishBadge(s) {
  return { open: 'badge-open', settled: 'badge-settled', canceled: 'badge-canceled' }[s] || 'badge-open'
}
function formatTime(t) {
  const d = new Date(t)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

async function loadCoins() {
  loadingCoins.value = true
  try {
    const { list } = await myCoins({ page: 1, size: 50 })
    coins.value = list
  } finally {
    loadingCoins.value = false
  }
}
async function loadWishes() {
  const { list } = await myWishes()
  wishes.value = list
}
async function loadResponses() {
  const { list } = await myResponses()
  responses.value = list
}
async function loadAll() {
  const { user } = await getMe()
  auth.setUser(user)
  await Promise.all([loadCoins(), loadWishes(), loadResponses()])
}

function openDetail(wish) {
  detailWishId.value = wish.id
}
async function onAddBounty(wish) {
  const amount = Number(prompt(`追加多少枚尝鲜币？（可用 ${auth.user?.coin_balance ?? 0}）`, '10'))
  if (!amount || amount < 1) return
  try {
    await addBounty(wish.id, amount)
    toast(`已追加冻结 🪙${amount}`)
    loadAll()
  } catch (e) {
    toast(e.message)
  }
}
async function onCancel(wish) {
  if (!confirm(`取消后冻结的 🪙${wish.bounty} 将全额退回。确认取消？`)) return
  try {
    await cancelWish(wish.id)
    toast('已取消，悬赏币已退回')
    loadAll()
  } catch (e) {
    toast(e.message)
  }
}

onMounted(loadAll)
</script>

<style scoped>
.wallet {
  background: linear-gradient(135deg, #ff8a50, #ff6b35 60%, #f5a623);
  border-radius: 20px;
  padding: 22px;
  color: #fff;
  box-shadow: 0 10px 28px rgba(255, 107, 53, 0.26);
}
.wallet-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.user {
  display: flex;
  align-items: center;
  gap: 12px;
}
.big-avatar {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.25);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  font-weight: 800;
}
.nick {
  font-size: 18px;
  font-weight: 800;
}
.username {
  font-size: 12px;
  opacity: 0.85;
}
.wallet .btn {
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.4);
}
.balances {
  display: flex;
  gap: 14px;
  margin-top: 18px;
}
.bal {
  flex: 1;
  background: rgba(255, 255, 255, 0.16);
  border-radius: 14px;
  padding: 14px;
}
.bal-num {
  font-size: 24px;
  font-weight: 800;
}
.bal-label {
  font-size: 12px;
  opacity: 0.85;
  margin-top: 2px;
}
.tabs {
  display: flex;
  gap: 8px;
  margin: 18px 0 14px;
}
.tabs button {
  flex: 1;
  background: #fff;
  border: 1px solid var(--line);
  color: var(--ink-2);
  padding: 9px;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 700;
}
.tabs button.active {
  background: var(--brand-soft);
  border-color: var(--brand);
  color: var(--brand);
}
.feed {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.coin-row {
  display: flex;
  align-items: center;
  gap: 12px;
  background: #fff;
  border-radius: 14px;
  padding: 13px 16px;
  margin-bottom: 10px;
  box-shadow: 0 2px 10px rgba(190, 110, 60, 0.06);
}
.coin-icon {
  font-size: 22px;
}
.coin-main {
  flex: 1;
  min-width: 0;
}
.coin-title {
  font-size: 14px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.coin-time {
  font-size: 12px;
  color: var(--ink-3);
}
.coin-amount {
  font-size: 17px;
  font-weight: 800;
  text-align: right;
}
.coin-amount.plus {
  color: var(--green);
}
.coin-amount.minus {
  color: var(--red);
}
.coin-bal {
  font-size: 11px;
  font-weight: 400;
  color: var(--ink-3);
  margin-top: 2px;
}
.resp-card {
  background: #fff;
  border-radius: 14px;
  padding: 15px;
  margin-bottom: 12px;
  box-shadow: 0 2px 10px rgba(190, 110, 60, 0.06);
}
.resp-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}
.stars {
  color: var(--gold);
  letter-spacing: 1px;
  font-size: 14px;
  margin: 6px 0;
}
.resp-text {
  margin: 0 0 10px;
  font-size: 14px;
  color: var(--ink-2);
  white-space: pre-wrap;
}
.resp-imgs {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.resp-imgs img {
  width: 104px;
  height: 104px;
  object-fit: cover;
  border-radius: 10px;
  cursor: zoom-in;
  background: #f5ece5;
}
.resp-foot {
  margin-top: 10px;
  font-size: 12px;
  color: var(--ink-3);
}
.lightbox {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.82);
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 30px;
  cursor: zoom-out;
}
.lightbox img {
  max-width: 100%;
  max-height: 100%;
  border-radius: 12px;
}
</style>
